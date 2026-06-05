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

type CreateJobProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建面试岗位
func NewCreateJobProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateJobProfileLogic {
	return &CreateJobProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateJobProfileLogic) CreateJobProfile(req *types.CreateJobProfileReq) (resp *types.CreateJobProfileResp, err error) {
	userId, ok := l.ctx.Value("userId").(string)
	if !ok || userId == "" {
		return nil, fmt.Errorf("unauthorized: missing user id in context")
	}

	rpcResp, err := l.svcCtx.CoreRpc.CreateJobProfile(l.ctx, &core.CreateJobProfileReq{
		CreatorId:   userId,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateJobProfileResp{
		Id: rpcResp.Id,
	}, nil
}

