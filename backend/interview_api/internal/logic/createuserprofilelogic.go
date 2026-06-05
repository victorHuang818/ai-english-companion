package logic

import (
	"context"

	"ai_interview/interview_api/internal/svc"
	"ai_interview/interview_api/internal/types"
	"ai_interview/rpc/core/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 提交并设置用户背景档案
func NewCreateUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserProfileLogic {
	return &CreateUserProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserProfileLogic) CreateUserProfile(req *types.CreateUserProfileReq) (resp *types.CreateUserProfileResp, err error) {
	userId, _ := l.ctx.Value("userId").(string)

	rpcResp, err := l.svcCtx.CoreRpc.CreateUserProfile(l.ctx, &coreclient.CreateUserProfileReq{
		UserId:         userId,
		EnglishLevel:   req.EnglishLevel,
		LearningTarget: req.LearningTarget,
	})
	if err != nil {
		l.Errorf("Failed to call CoreRpc.CreateUserProfile: %v", err)
		return nil, err
	}

	return &types.CreateUserProfileResp{
		Id:     rpcResp.Id,
		Status: rpcResp.Status,
	}, nil
}
