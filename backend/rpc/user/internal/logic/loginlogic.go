package logic

import (
	"context"
	"fmt"
	"time"

	"ai_companion/pkg/auth"
	"ai_companion/rpc/user/internal/svc"
	"ai_companion/rpc/user/model"
	"ai_companion/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	// 1. 查找用户
	u, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, in.Email)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}

	// 2. 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password))
	if err != nil {
		return nil, fmt.Errorf("密码错误")
	}

	// 3. 生成 Token
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.JwtAuth.AccessExpire
	token, err := auth.GetJwtToken(l.svcCtx.Config.JwtAuth.AccessSecret, now, accessExpire, u.Id)
	if err != nil {
		return nil, err
	}

	return &user.LoginResp{
		Id:    u.Id,
		Token: token,
	}, nil
}

