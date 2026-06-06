package logic

import (
	"context"
	"fmt"

	"ai_companion/rpc/user/internal/svc"
	"ai_companion/rpc/user/model"
	"ai_companion/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductTokenLogic {
	return &DeductTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeductTokenLogic) DeductToken(in *user.DeductTokenReq) (*user.DeductTokenResp, error) {
	// 1. 获取用户信息
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}

	amount := uint64(in.Amount)
	totalTokens := u.DailyFreeTokens + u.RechargeTokens
	if totalTokens < amount {
		return &user.DeductTokenResp{Success: false}, nil
	}

	// 2. 扣减逻辑：优先每日免费，不足部分扣充值
	if u.DailyFreeTokens >= amount {
		u.DailyFreeTokens -= amount
	} else {
		u.RechargeTokens -= (amount - u.DailyFreeTokens)
		u.DailyFreeTokens = 0
	}

	// 3. 更新数据库
	err = l.svcCtx.UserModel.Update(l.ctx, u)
	if err != nil {
		return nil, err
	}

	return &user.DeductTokenResp{Success: true}, nil
}


