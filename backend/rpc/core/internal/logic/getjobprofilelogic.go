package logic

import (
	"context"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetJobProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetJobProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetJobProfileLogic {
	return &GetJobProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetJobProfileLogic) GetJobProfile(in *core.GetJobProfileReq) (*core.GetJobProfileResp, error) {
	// todo: add your logic here and delete this line

	return &core.GetJobProfileResp{}, nil
}
