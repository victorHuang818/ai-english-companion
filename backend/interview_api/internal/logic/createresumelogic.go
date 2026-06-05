// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"

	"ai_interview/interview_api/internal/svc"
	"ai_interview/interview_api/internal/types"
	"ai_interview/rpc/core/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 提交通知，触发简历解析入库
func NewCreateResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateResumeLogic {
	return &CreateResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateResumeLogic) CreateResume(req *types.CreateResumeReq) (resp *types.CreateResumeResp, err error) {
	userId, ok := l.ctx.Value("userId").(string)
	if !ok || userId == "" {
		return nil, fmt.Errorf("unauthorized: missing user id in context")
	}

	rpcResp, err := l.svcCtx.CoreRpc.CreateResume(l.ctx, &core.CreateResumeReq{
		UserId:    userId,
		ObjectKey: req.ObjectKey,
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateResumeResp{
		Id:     rpcResp.Id,
		Status: rpcResp.Status,
	}, nil
}

