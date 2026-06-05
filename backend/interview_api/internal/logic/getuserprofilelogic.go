package logic

import (
	"context"

	"ai_interview/interview_api/internal/svc"
	"ai_interview/interview_api/internal/types"
	"ai_interview/rpc/core/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户档案信息
func NewGetUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserProfileLogic {
	return &GetUserProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserProfileLogic) GetUserProfile(req *types.GetUserProfileReq) (resp *types.GetUserProfileResp, err error) {
	rpcResp, err := l.svcCtx.CoreRpc.GetUserProfile(l.ctx, &coreclient.GetUserProfileReq{
		Id: req.Id,
	})
	if err != nil {
		l.Errorf("Failed to call CoreRpc.GetUserProfile: %v", err)
		return nil, err
	}

	return &types.GetUserProfileResp{
		Id:             rpcResp.Id,
		EnglishLevel:   rpcResp.EnglishLevel,
		LearningTarget: rpcResp.LearningTarget,
		CreatedAt:      0,
	}, nil
}
