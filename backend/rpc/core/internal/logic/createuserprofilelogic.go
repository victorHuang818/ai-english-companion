package logic

import (
	"context"

	"ai_companion/rpc/core/core"
	"ai_companion/rpc/core/internal/svc"
	"ai_companion/rpc/core/model"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserProfileLogic {
	return &CreateUserProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateUserProfileLogic) CreateUserProfile(in *core.CreateUserProfileReq) (*core.CreateUserProfileResp, error) {
	id := uuid.New().String()
	_, err := l.svcCtx.UserProfileModel.Insert(l.ctx, &model.UserProfiles{
		Id:             id,
		UserId:         in.UserId,
		EnglishLevel:   in.EnglishLevel,
		LearningTarget: in.LearningTarget,
	})
	if err != nil {
		l.Errorf("CreateUserProfile error: %v", err)
		return nil, err
	}

	return &core.CreateUserProfileResp{
		Id:     id,
		Status: "parsed",
	}, nil
}
