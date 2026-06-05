package logic

import (
	"context"

	"ai_interview/interview_api/internal/svc"
	"ai_interview/interview_api/internal/types"
	"ai_interview/pkg/error_pkg"
	"ai_interview/rpc/core/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建口语练习会话
func NewCreateSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSessionLogic {
	return &CreateSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSessionLogic) CreateSession(req *types.CreateSessionReq) (resp *types.CreateSessionResp, err error) {
	userId, ok := l.ctx.Value("userId").(string)
	if !ok || userId == "" {
		return nil, error_pkg.NewCodeError(error_pkg.RequestParamError, "未获取到有效用户身份")
	}

	rpcResp, err := l.svcCtx.CoreRpc.CreateSession(l.ctx, &coreclient.CreateSessionReq{
		UserId:        userId,
		UserProfileId: req.UserProfileId,
		ScenarioId:    req.ScenarioId,
	})
	if err != nil {
		l.Errorf("CoreRpc.CreateSession error: %v", err)
		return nil, err
	}

	return &types.CreateSessionResp{
		SessionId: rpcResp.Id,
	}, nil
}
