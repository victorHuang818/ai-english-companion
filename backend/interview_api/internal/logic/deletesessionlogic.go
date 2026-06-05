package logic

import (
	"context"

	"ai_interview/interview_api/internal/svc"
	"ai_interview/interview_api/internal/types"
	"ai_interview/rpc/core/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除口语练习记录
func NewDeleteSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSessionLogic {
	return &DeleteSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSessionLogic) DeleteSession(req *types.DeleteSessionReq) (resp *types.DeleteSessionResp, err error) {
	rpcResp, err := l.svcCtx.CoreRpc.DeleteSession(l.ctx, &coreclient.DeleteSessionReq{
		SessionId: req.SessionId,
	})
	if err != nil {
		l.Errorf("CoreRpc.DeleteSession error: %v", err)
		return nil, err
	}

	return &types.DeleteSessionResp{
		Success: rpcResp.Success,
	}, nil
}
