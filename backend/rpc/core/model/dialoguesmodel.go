package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DialoguesModel = (*customDialoguesModel)(nil)

type (
	// DialoguesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDialoguesModel.
	DialoguesModel interface {
		dialoguesModel
	}

	customDialoguesModel struct {
		*defaultDialoguesModel
	}
)

// NewDialoguesModel returns a model for the database table.
func NewDialoguesModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) DialoguesModel {
	return &customDialoguesModel{
		defaultDialoguesModel: newDialoguesModel(conn, c, opts...),
	}
}
