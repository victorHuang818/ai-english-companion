package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PracticeSessionsModel = (*customPracticeSessionsModel)(nil)

type (
	// PracticeSessionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPracticeSessionsModel.
	PracticeSessionsModel interface {
		practiceSessionsModel
	}

	customPracticeSessionsModel struct {
		*defaultPracticeSessionsModel
	}
)

// NewPracticeSessionsModel returns a model for the database table.
func NewPracticeSessionsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) PracticeSessionsModel {
	return &customPracticeSessionsModel{
		defaultPracticeSessionsModel: newPracticeSessionsModel(conn, c, opts...),
	}
}
