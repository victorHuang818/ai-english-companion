package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ InterviewSessionsModel = (*customInterviewSessionsModel)(nil)

type (
	// InterviewSessionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customInterviewSessionsModel.
	InterviewSessionsModel interface {
		interviewSessionsModel
	}

	customInterviewSessionsModel struct {
		*defaultInterviewSessionsModel
	}
)

// NewInterviewSessionsModel returns a model for the database table.
func NewInterviewSessionsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) InterviewSessionsModel {
	return &customInterviewSessionsModel{
		defaultInterviewSessionsModel: newInterviewSessionsModel(conn, c, opts...),
	}
}
