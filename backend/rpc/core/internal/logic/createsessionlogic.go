package logic

import (
	"context"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"
	"ai_interview/rpc/core/model"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSessionLogic {
	return &CreateSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Interview Sessions
func (l *CreateSessionLogic) CreateSession(in *core.CreateSessionReq) (*core.CreateSessionResp, error) {
	sessionID := uuid.New().String()

	data := &model.InterviewSessions{
		Id:             sessionID,
		UserId:         in.UserId,
		ResumeId:       in.ResumeId,
		JobProfileId:   in.JobProfileId,
		Status:         "in_progress",
	}

	_, err := l.svcCtx.InterviewSessionModel.Insert(l.ctx, data)
	if err != nil {
		l.Errorf("Failed to create interview session: %v", err)
		return nil, err
	}

	return &core.CreateSessionResp{
		Id: sessionID,
	}, nil
}
