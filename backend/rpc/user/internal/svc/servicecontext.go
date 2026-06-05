package svc

import (
	"ai_interview/rpc/user/internal/config"
	"ai_interview/rpc/user/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UsersModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUsersModel(sqlConn, c.CacheRedis),
	}
}
