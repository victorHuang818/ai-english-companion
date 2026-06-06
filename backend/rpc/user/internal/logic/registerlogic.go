package logic

import (
	"context"
	"fmt"

	"ai_companion/rpc/user/internal/svc"
	"ai_companion/rpc/user/model"
	"ai_companion/rpc/user/user"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	// 1. 检查用户是否已存在
	_, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.Username)
	if err == nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	_, err = l.svcCtx.UserModel.FindOneByEmail(l.ctx, in.Email)
	if err == nil {
		return nil, fmt.Errorf("邮箱已被注册")
	}

	// 2. 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. 插入数据库
	userId := uuid.New().String()
	_, err = l.svcCtx.UserModel.Insert(l.ctx, &model.Users{
		Id:              userId,
		Username:        in.Username,
		Email:           in.Email,
		Password:        string(hashedPassword),
		DailyFreeTokens: 1000, // 初始赠送额度
	})
	if err != nil {
		return nil, err
	}

	return &user.RegisterResp{
		Id: userId,
	}, nil
}

