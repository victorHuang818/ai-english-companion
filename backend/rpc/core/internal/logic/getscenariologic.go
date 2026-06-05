package logic

import (
	"context"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetScenarioLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetScenarioLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScenarioLogic {
	return &GetScenarioLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetScenarioLogic) GetScenario(in *core.GetScenarioReq) (*core.GetScenarioResp, error) {
	res, err := l.svcCtx.ScenarioModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Errorf("GetScenario error: %v", err)
		return nil, err
	}

	return &core.GetScenarioResp{
		Id:          res.Id,
		Name:        res.Name,
		Description: res.Description.String,
	}, nil
}
