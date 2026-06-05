// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"ai_interview/interview_api/internal/svc"
	"ai_interview/interview_api/internal/types"
	"ai_interview/pkg/error_pkg"
	"ai_interview/rpc/core/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建面试会话 (第一步走 HTTP)
func NewCreateSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSessionLogic {
	return &CreateSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSessionLogic) CreateSession(req *types.CreateSessionReq) (resp *types.CreateSessionResp, err error) {
	// 1. 从 context 获取 user_id (由全局网关透传并在中间件注入到 ctx)
	userId, ok := l.ctx.Value("userId").(string)
	if !ok || userId == "" {
		return nil, error_pkg.NewCodeError(error_pkg.RequestParamError, "未获取到有效用户身份")
	}

	// 2. 调用 core_rpc 创建会话
	rpcResp, err := l.svcCtx.CoreRpc.CreateSession(l.ctx, &core.CreateSessionReq{
		UserId:       userId,
		ResumeId:     req.ResumeId,
		JobProfileId: req.JobProfileId,
	})
	if err != nil {
		l.Errorf("CoreRpc.CreateSession error: %v", err)
		return nil, err
	}

	return &types.CreateSessionResp{
		SessionId: rpcResp.Id,
	}, nil
}

