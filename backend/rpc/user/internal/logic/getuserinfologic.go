package logic

import (
	"context"
	"fmt"

	"ai_companion/rpc/user/internal/svc"
	"ai_companion/rpc/user/model"
	"ai_companion/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoReq) (*user.GetUserInfoResp, error) {
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}

	return &user.GetUserInfoResp{
		Id:              u.Id,
		Username:        u.Username,
		Email:           u.Email,
		DailyFreeTokens: uint32(u.DailyFreeTokens),
		RechargeTokens:  u.RechargeTokens,
	}, nil
}


