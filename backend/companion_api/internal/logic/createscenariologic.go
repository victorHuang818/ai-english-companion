package logic

import (
	"context"

	"ai_companion/companion_api/internal/svc"
	"ai_companion/companion_api/internal/types"
	"ai_companion/rpc/core/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateScenarioLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建练习场景
func NewCreateScenarioLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateScenarioLogic {
	return &CreateScenarioLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateScenarioLogic) CreateScenario(req *types.CreateScenarioReq) (resp *types.CreateScenarioResp, err error) {
	userId, _ := l.ctx.Value("userId").(string)

	rpcResp, err := l.svcCtx.CoreRpc.CreateScenario(l.ctx, &coreclient.CreateScenarioReq{
		CreatorId:   userId,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		l.Errorf("Failed to call CoreRpc.CreateScenario: %v", err)
		return nil, err
	}

	return &types.CreateScenarioResp{
		Id: rpcResp.Id,
	}, nil
}
