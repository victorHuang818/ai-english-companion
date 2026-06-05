package logic

import (
	"context"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSessionLogic {
	return &DeleteSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteSessionLogic) DeleteSession(in *core.DeleteSessionReq) (*core.DeleteSessionResp, error) {
	// 1. Delete all dialogues under this session first (completely delete)
	queryDialogues := "DELETE FROM interview_dialogues WHERE interview_session_id = ?"
	_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx, queryDialogues, in.SessionId)
	if err != nil {
		l.Errorf("Failed to delete dialogues for session %s: %v", in.SessionId, err)
		return nil, err
	}

	// 2. Delete the session itself using cache-enabled Delete method (completely delete)
	err = l.svcCtx.InterviewSessionModel.Delete(l.ctx, in.SessionId)
	if err != nil {
		l.Errorf("Failed to delete session %s: %v", in.SessionId, err)
		return nil, err
	}

	return &core.DeleteSessionResp{Success: true}, nil
}
