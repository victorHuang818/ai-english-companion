package logic

import (
	"context"
	"fmt"

	"ai_interview/interview_api/internal/svc"
	"ai_interview/interview_api/internal/types"
	"ai_interview/rpc/core/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取面试历史列表
func NewListSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSessionsLogic {
	return &ListSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSessionsLogic) ListSessions() (resp *types.ListSessionsResp, err error) {
	userId, ok := l.ctx.Value("userId").(string)
	if !ok || userId == "" {
		return nil, fmt.Errorf("unauthorized")
	}

	rpcResp, err := l.svcCtx.CoreRpc.ListSessions(l.ctx, &core.ListSessionsReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	sessions := make([]types.InterviewSession, 0, len(rpcResp.Sessions))
	for _, s := range rpcResp.Sessions {
		sessions = append(sessions, types.InterviewSession{
			SessionId:    s.Id,
			JobTitle:     s.JobTitle,
			Status:       s.Status,
			CreatedAt:    s.CreatedAt,
			OverallScore: s.OverallScore,
		})
	}

	return &types.ListSessionsResp{
		Sessions: sessions,
	}, nil
}
