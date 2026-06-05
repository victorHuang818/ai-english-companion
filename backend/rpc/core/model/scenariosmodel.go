package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ScenariosModel = (*customScenariosModel)(nil)

type (
	// ScenariosModel is an interface to be customized, add more methods here,
	// and implement the added methods in customScenariosModel.
	ScenariosModel interface {
		scenariosModel
	}

	customScenariosModel struct {
		*defaultScenariosModel
	}
)

// NewScenariosModel returns a model for the database table.
func NewScenariosModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ScenariosModel {
	return &customScenariosModel{
		defaultScenariosModel: newScenariosModel(conn, c, opts...),
	}
}
