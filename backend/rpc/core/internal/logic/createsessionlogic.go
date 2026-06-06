package logic

import (
	"context"

	"ai_companion/rpc/core/core"
	"ai_companion/rpc/core/internal/svc"
	"ai_companion/rpc/core/model"

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

// Practice Sessions
func (l *CreateSessionLogic) CreateSession(in *core.CreateSessionReq) (*core.CreateSessionResp, error) {
	sessionID := uuid.New().String()

	data := &model.PracticeSessions{
		Id:            sessionID,
		UserId:        in.UserId,
		UserProfileId: in.UserProfileId,
		ScenarioId:    in.ScenarioId,
		Status:        "in_progress",
	}

	_, err := l.svcCtx.PracticeSessionModel.Insert(l.ctx, data)
	if err != nil {
		l.Errorf("Failed to create practice session: %v", err)
		return nil, err
	}

	return &core.CreateSessionResp{
		Id: sessionID,
	}, nil
}
