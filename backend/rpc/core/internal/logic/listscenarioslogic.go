package logic

import (
	"context"
	"time"

	"ai_companion/rpc/core/core"
	"ai_companion/rpc/core/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListScenariosLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListScenariosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListScenariosLogic {
	return &ListScenariosLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type dbScenarioItem struct {
	Id          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatorId   string    `db:"creator_id"`
	CreatedAt   time.Time `db:"created_at"`
}

func (l *ListScenariosLogic) ListScenarios(in *core.ListScenariosReq) (*core.ListScenariosResp, error) {
	var list []dbScenarioItem
	query := `
		SELECT id, name, description, creator_id, created_at
		FROM scenarios
		WHERE creator_id = '00000000-0000-0000-0000-000000000000' OR creator_id = ?
		ORDER BY created_at DESC`

	err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &list, query, in.UserId)
	if err != nil {
		l.Errorf("Failed to query scenarios for user %s: %v", in.UserId, err)
		return nil, err
	}

	scenarios := make([]*core.ScenarioItem, 0, len(list))
	for _, item := range list {
		scenarios = append(scenarios, &core.ScenarioItem{
			Id:          item.Id,
			Name:        item.Name,
			Description: item.Description,
			CreatorId:   item.CreatorId,
			CreatedAt:   item.CreatedAt.Unix(),
		})
	}

	return &core.ListScenariosResp{
		Scenarios: scenarios,
	}, nil
}
