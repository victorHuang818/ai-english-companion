// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"ai_companion/companion_api/internal/svc"
	"ai_companion/companion_api/internal/types"
	"ai_companion/rpc/core/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserProfilesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取当前用户的档案列表
func NewListUserProfilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserProfilesLogic {
	return &ListUserProfilesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserProfilesLogic) ListUserProfiles() (resp *types.ListUserProfilesResp, err error) {
	userId, _ := l.ctx.Value("userId").(string)

	rpcResp, err := l.svcCtx.CoreRpc.ListUserProfiles(l.ctx, &coreclient.ListUserProfilesReq{
		UserId: userId,
	})
	if err != nil {
		l.Errorf("Failed to list user profiles: %v", err)
		return nil, err
	}

	profiles := make([]types.UserProfileItem, 0, len(rpcResp.Profiles))
	for _, p := range rpcResp.Profiles {
		profiles = append(profiles, types.UserProfileItem{
			Id:             p.Id,
			EnglishLevel:   p.EnglishLevel,
			LearningTarget: p.LearningTarget,
			CreatedAt:      p.CreatedAt,
		})
	}

	return &types.ListUserProfilesResp{
		Profiles: profiles,
	}, nil
}
