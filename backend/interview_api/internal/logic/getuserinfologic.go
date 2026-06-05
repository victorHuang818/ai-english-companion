// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"

	"ai_interview/interview_api/internal/svc"
	"ai_interview/interview_api/internal/types"
	"ai_interview/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取当前用户信息
func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo() (resp *types.GetUserInfoResp, err error) {
	userId, ok := l.ctx.Value("userId").(string)
	if !ok || userId == "" {
		return nil, fmt.Errorf("unauthorized: missing user id in context")
	}

	rpcResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.GetUserInfoReq{
		Id: userId,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetUserInfoResp{
		Id:              rpcResp.Id,
		Username:        rpcResp.Username,
		Email:           rpcResp.Email,
		DailyFreeTokens: rpcResp.DailyFreeTokens,
		RechargeTokens:  rpcResp.RechargeTokens,
	}, nil
}

