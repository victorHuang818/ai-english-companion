package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"ai_companion/rpc/core/core"
	"ai_companion/rpc/core/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSessionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSessionsLogic {
	return &ListSessionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type dbSessionItem struct {
	Id               string         `db:"id"`
	ScenarioName     string         `db:"scenario_name"`
	Status           string         `db:"status"`
	CreatedAt        time.Time      `db:"created_at"`
	EvaluationReport sql.NullString `db:"evaluation_report"`
}

func (l *ListSessionsLogic) ListSessions(in *core.ListSessionsReq) (*core.ListSessionsResp, error) {
	var list []dbSessionItem
	query := `
		SELECT s.id, COALESCE(j.name, '未知场景') AS scenario_name, s.status, s.created_at, s.evaluation_report
		FROM practice_sessions s
		LEFT JOIN scenarios j ON s.scenario_id = j.id
		WHERE s.user_id = ?
		ORDER BY s.created_at DESC`
	
	err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &list, query, in.UserId)
	if err != nil {
		l.Errorf("Failed to query sessions for user %s: %v", in.UserId, err)
		return nil, err
	}

	sessions := make([]*core.SessionItem, 0, len(list))
	for _, item := range list {
		var score int32 = 0
		if item.EvaluationReport.Valid && item.EvaluationReport.String != "" {
			var report map[string]interface{}
			if err := json.Unmarshal([]byte(item.EvaluationReport.String), &report); err == nil {
				if val, ok := report["overall_score"]; ok {
					if fVal, ok := val.(float64); ok {
						score = int32(fVal)
					}
				} else if val, ok := report["score"]; ok {
					if fVal, ok := val.(float64); ok {
						score = int32(fVal)
					}
				}
			}
		}

		sessions = append(sessions, &core.SessionItem{
			Id:           item.Id,
			ScenarioName: item.ScenarioName,
			Status:       item.Status,
			CreatedAt:    item.CreatedAt.Unix(),
			OverallScore: score,
		})
	}

	return &core.ListSessionsResp{
		Sessions: sessions,
	}, nil
}
