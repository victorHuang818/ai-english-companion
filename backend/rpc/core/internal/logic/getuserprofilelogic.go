package logic

import (
	"context"
	"fmt"

	"ai_companion/rpc/core/core"
	"ai_companion/rpc/core/internal/svc"
	"ai_companion/rpc/core/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserProfileLogic {
	return &GetUserProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserProfileLogic) GetUserProfile(in *core.GetUserProfileReq) (*core.GetUserProfileResp, error) {
	profile, err := l.svcCtx.UserProfileModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("user profile not found")
		}
		l.Errorf("GetUserProfile error: %v", err)
		return nil, err
	}

	return &core.GetUserProfileResp{
		Id:             profile.Id,
		UserId:         profile.UserId,
		EnglishLevel:   profile.EnglishLevel,
		LearningTarget: profile.LearningTarget,
	}, nil
}
