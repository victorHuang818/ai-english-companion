package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"ai_companion/rpc/core/core"
	"ai_companion/rpc/core/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSessionDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSessionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSessionDetailLogic {
	return &GetSessionDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSessionDetailLogic) GetSessionDetail(in *core.GetSessionDetailReq) (*core.GetSessionDetailResp, error) {
	// 1. Query the session
	session, err := l.svcCtx.PracticeSessionModel.FindOne(l.ctx, in.SessionId)
	if err != nil {
		l.Errorf("Failed to find session %s: %v", in.SessionId, err)
		return nil, err
	}

	// 2. Query scenario name
	var scenarioName string
	queryScenario := "SELECT name FROM scenarios WHERE id = ?"
	err = l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &scenarioName, queryScenario, session.ScenarioId)
	if err != nil {
		l.Errorf("Failed to query scenario for session %s: %v", in.SessionId, err)
		scenarioName = "未知场景"
	}

	// 3. Query all dialogues
	type dbDialogue struct {
		Role       string         `db:"role"`
		Content    string         `db:"content"`
		AudioUrl   sql.NullString `db:"audio_url"`
		Evaluation sql.NullString `db:"evaluation"`
	}
	var dialogues []dbDialogue
	queryDialogues := "SELECT role, content, audio_url, evaluation FROM dialogues WHERE practice_session_id = ? ORDER BY created_at ASC, role ASC"
	err = l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &dialogues, queryDialogues, in.SessionId)
	if err != nil {
		l.Errorf("Failed to query dialogues for session %s: %v", in.SessionId, err)
	}

	// Format transcript as JSON array
	type transcriptItem struct {
		Role       string                 `json:"role"`
		Content    string                 `json:"content"`
		AudioUrl   string                 `json:"audio_url,omitempty"`
		Evaluation map[string]interface{} `json:"evaluation,omitempty"`
	}

	var transcriptItems []transcriptItem
	for _, d := range dialogues {
		role := "ai"
		if d.Role == "user" {
			role = "user"
		}
		item := transcriptItem{
			Role:    role,
			Content: d.Content,
		}
		if d.AudioUrl.Valid && d.AudioUrl.String != "" {
			presigned, err := l.svcCtx.OssClient.GeneratePresignedDownloadURL(l.ctx, d.AudioUrl.String, time.Hour)
			if err == nil {
				item.AudioUrl = presigned
			} else {
				l.Errorf("Failed to generate presigned download URL for object %s: %v", d.AudioUrl.String, err)
			}
		}
		if d.Role == "user" && d.Evaluation.Valid && d.Evaluation.String != "" {
			var eval map[string]interface{}
			cleanJSON := strings.ReplaceAll(d.Evaluation.String, "```json", "")
			cleanJSON = strings.ReplaceAll(cleanJSON, "```", "")
			cleanJSON = strings.TrimSpace(cleanJSON)
			if err := json.Unmarshal([]byte(cleanJSON), &eval); err == nil {
				item.Evaluation = eval
			}
		}
		transcriptItems = append(transcriptItems, item)
	}
	transcriptBytes, _ := json.Marshal(transcriptItems)
	transcriptJSON := string(transcriptBytes)

	// 4. Consolidate evaluation report
	var evaluationReport string
	var overallScore int32

	if session.EvaluationReport.Valid && session.EvaluationReport.String != "" {
		evaluationReport = session.EvaluationReport.String
		var report map[string]interface{}
		if err := json.Unmarshal([]byte(evaluationReport), &report); err == nil {
			if val, ok := report["overall_score"]; ok {
				if fVal, ok := val.(float64); ok {
					overallScore = int32(fVal)
				}
			} else if val, ok := report["score"]; ok {
				if fVal, ok := val.(float64); ok {
					overallScore = int32(fVal)
				}
			}
		}
	} else {
		var err error
		evaluationReport, overallScore, err = ConsolidateReport(l.ctx, l.svcCtx, in.SessionId)
		if err != nil {
			l.Errorf("Failed to consolidate report for session %s: %v", in.SessionId, err)
		}
	}

	return &core.GetSessionDetailResp{
		SessionId:        session.Id,
		ScenarioName:     scenarioName,
		Transcript:       transcriptJSON,
		OverallScore:     overallScore,
		EvaluationReport: evaluationReport,
	}, nil
}
