package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ResumesModel = (*customResumesModel)(nil)

type (
	// ResumesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customResumesModel.
	ResumesModel interface {
		resumesModel
	}

	customResumesModel struct {
		*defaultResumesModel
	}
)

// NewResumesModel returns a model for the database table.
func NewResumesModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ResumesModel {
	return &customResumesModel{
		defaultResumesModel: newResumesModel(conn, c, opts...),
	}
}
