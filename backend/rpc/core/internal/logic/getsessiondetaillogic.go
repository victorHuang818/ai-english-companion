package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"

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
	session, err := l.svcCtx.InterviewSessionModel.FindOne(l.ctx, in.SessionId)
	if err != nil {
		l.Errorf("Failed to find session %s: %v", in.SessionId, err)
		return nil, err
	}

	// 2. Query job title
	var jobTitle string
	queryJob := "SELECT name FROM job_profiles WHERE id = ?"
	err = l.svcCtx.SqlConn.QueryRowCtx(l.ctx, &jobTitle, queryJob, session.JobProfileId)
	if err != nil {
		l.Errorf("Failed to query job profile for session %s: %v", in.SessionId, err)
		jobTitle = "未知岗位"
	}

	// 3. Query all dialogues
	type dbDialogue struct {
		Role       string         `db:"role"`
		Content    string         `db:"content"`
		Evaluation sql.NullString `db:"evaluation"`
	}
	var dialogues []dbDialogue
	queryDialogues := "SELECT role, content, evaluation FROM interview_dialogues WHERE interview_session_id = ? ORDER BY created_at ASC"
	err = l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &dialogues, queryDialogues, in.SessionId)
	if err != nil {
		l.Errorf("Failed to query dialogues for session %s: %v", in.SessionId, err)
	}

	// Format transcript
	var transcriptBuilder strings.Builder
	for _, d := range dialogues {
		if d.Role == "interviewer" {
			transcriptBuilder.WriteString(fmt.Sprintf("AI:%s\n", d.Content))
		} else {
			transcriptBuilder.WriteString(fmt.Sprintf("User:%s\n", d.Content))
		}
	}

	// 4. Consolidate evaluation report if session is completed (or even in progress) and empty
	var evaluationReport string
	var overallScore int32

	if session.EvaluationReport.Valid && session.EvaluationReport.String != "" {
		evaluationReport = session.EvaluationReport.String
	} else {
		// Consolidate from dialogues
		var totalScore float64
		var count int
		var fluencyList []string
		var relevanceList []string
		var logicList []string
		var depthList []string
		var starList []string
		var overallCommentList []string

		for _, d := range dialogues {
			if d.Role == "user" && d.Evaluation.Valid && d.Evaluation.String != "" {
				var round map[string]interface{}
				cleanJSON := strings.ReplaceAll(d.Evaluation.String, "```json", "")
				cleanJSON = strings.ReplaceAll(cleanJSON, "```", "")
				cleanJSON = strings.TrimSpace(cleanJSON)
				if err := json.Unmarshal([]byte(cleanJSON), &round); err == nil {
					// Extract score
					var s float64
					if val, ok := round["score"]; ok {
						if fVal, ok := val.(float64); ok {
							s = fVal
						}
					} else if val, ok := round["overall_score"]; ok {
						if fVal, ok := val.(float64); ok {
							s = fVal
						}
					}
					if s > 0 {
						totalScore += s
						count++
					}

					// Extract dimensions
					if dims, ok := round["dimensions"].(map[string]interface{}); ok {
						if val, ok := dims["fluency"].(string); ok && val != "" {
							fluencyList = append(fluencyList, val)
						}
						if val, ok := dims["relevance"].(string); ok && val != "" {
							relevanceList = append(relevanceList, val)
						}
						if val, ok := dims["logic"].(string); ok && val != "" {
							logicList = append(logicList, val)
						}
						if val, ok := dims["depth"].(string); ok && val != "" {
							depthList = append(depthList, val)
						}
						if val, ok := dims["star_alignment"].(string); ok && val != "" {
							starList = append(starList, val)
						}
					}

					// Extract overall comment
					if comment, ok := round["overall_comment"].(string); ok && comment != "" {
						overallCommentList = append(overallCommentList, comment)
					}
				}
			}
		}

		if count > 0 {
			overallScore = int32(totalScore / float64(count))
			// Join comments
			joinFeedback := func(list []string) string {
				if len(list) == 0 {
					return "评估未见异常。"
				}
				if len(list) == 1 {
					return list[0]
				}
				// Format as bullet points or joined strings
				var sb strings.Builder
				for i, item := range list {
					sb.WriteString(fmt.Sprintf("%d. %s ", i+1, item))
				}
				return sb.String()
			}

			reportObj := map[string]interface{}{
				"score":         overallScore,
				"overall_score": overallScore,
				"dimensions": map[string]string{
					"fluency":        joinFeedback(fluencyList),
					"relevance":      joinFeedback(relevanceList),
					"logic":          joinFeedback(logicList),
					"depth":          joinFeedback(depthList),
					"star_alignment": joinFeedback(starList),
				},
				"overall_comment": joinFeedback(overallCommentList),
			}

			reportBytes, _ := json.Marshal(reportObj)
			evaluationReport = string(reportBytes)

			// Save evaluation report back to session
			session.EvaluationReport = sql.NullString{String: evaluationReport, Valid: true}
			_ = l.svcCtx.InterviewSessionModel.Update(l.ctx, session)
		} else {
			// Return default empty evaluation
			reportObj := map[string]interface{}{
				"score":         0,
				"overall_score": 0,
				"dimensions": map[string]string{
					"fluency":        "本次面试未检测到有效的回答记录。",
					"relevance":      "本次面试未检测到有效的回答记录。",
					"logic":          "本次面试未检测到有效的回答记录。",
					"depth":          "本次面试未检测到有效的回答记录。",
					"star_alignment": "本次面试未检测到有效的回答记录。",
				},
				"overall_comment": "未进行充分的对话，无法生成最终评估报告。",
			}
			reportBytes, _ := json.Marshal(reportObj)
			evaluationReport = string(reportBytes)
			overallScore = 0
		}
	}

	// Calculate overall score from evaluationReport if we got it from DB
	if overallScore == 0 && evaluationReport != "" {
		var report map[string]interface{}
		if err := json.Unmarshal([]byte(evaluationReport), &report); err == nil {
			if val, ok := report["score"]; ok {
				if fVal, ok := val.(float64); ok {
					overallScore = int32(fVal)
				}
			} else if val, ok := report["overall_score"]; ok {
				if fVal, ok := val.(float64); ok {
					overallScore = int32(fVal)
				}
			}
		}
	}

	return &core.GetSessionDetailResp{
		SessionId:        session.Id,
		JobTitle:         jobTitle,
		Transcript:       transcriptBuilder.String(),
		OverallScore:     overallScore,
		EvaluationReport: evaluationReport,
	}, nil
}
