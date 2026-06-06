package logic

import (
	"context"
	"time"

	"ai_companion/rpc/core/core"
	"ai_companion/rpc/core/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserProfilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserProfilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserProfilesLogic {
	return &ListUserProfilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type dbUserProfileItem struct {
	Id             string    `db:"id"`
	EnglishLevel   string    `db:"english_level"`
	LearningTarget string    `db:"learning_target"`
	CreatedAt      time.Time `db:"created_at"`
}

func (l *ListUserProfilesLogic) ListUserProfiles(in *core.ListUserProfilesReq) (*core.ListUserProfilesResp, error) {
	var list []dbUserProfileItem
	query := `
		SELECT id, english_level, learning_target, created_at
		FROM user_profiles
		WHERE user_id = ?
		ORDER BY created_at DESC`

	err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &list, query, in.UserId)
	if err != nil {
		l.Errorf("Failed to query user profiles for user %s: %v", in.UserId, err)
		return nil, err
	}

	profiles := make([]*core.UserProfileItem, 0, len(list))
	for _, item := range list {
		profiles = append(profiles, &core.UserProfileItem{
			Id:             item.Id,
			EnglishLevel:   item.EnglishLevel,
			LearningTarget: item.LearningTarget,
			CreatedAt:      item.CreatedAt.Unix(),
		})
	}

	return &core.ListUserProfilesResp{
		Profiles: profiles,
	}, nil
}
