// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"ai_companion/companion_api/internal/svc"
	"ai_companion/companion_api/internal/types"
	"ai_companion/rpc/core/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListScenariosLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取场景列表
func NewListScenariosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListScenariosLogic {
	return &ListScenariosLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListScenariosLogic) ListScenarios() (resp *types.ListScenariosResp, err error) {
	userId, _ := l.ctx.Value("userId").(string)

	rpcResp, err := l.svcCtx.CoreRpc.ListScenarios(l.ctx, &coreclient.ListScenariosReq{
		UserId: userId,
	})
	if err != nil {
		l.Errorf("Failed to list scenarios: %v", err)
		return nil, err
	}

	scenarios := make([]types.ScenarioItem, 0, len(rpcResp.Scenarios))
	for _, s := range rpcResp.Scenarios {
		scenarios = append(scenarios, types.ScenarioItem{
			Id:          s.Id,
			Name:        s.Name,
			Description: s.Description,
			CreatorId:   s.CreatorId,
			CreatedAt:   s.CreatedAt,
		})
	}

	return &types.ListScenariosResp{
		Scenarios: scenarios,
	}, nil
}
