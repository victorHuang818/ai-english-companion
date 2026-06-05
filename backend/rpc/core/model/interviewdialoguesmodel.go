package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ InterviewDialoguesModel = (*customInterviewDialoguesModel)(nil)

type (
	// InterviewDialoguesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customInterviewDialoguesModel.
	InterviewDialoguesModel interface {
		interviewDialoguesModel
	}

	customInterviewDialoguesModel struct {
		*defaultInterviewDialoguesModel
	}
)

// NewInterviewDialoguesModel returns a model for the database table.
func NewInterviewDialoguesModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) InterviewDialoguesModel {
	return &customInterviewDialoguesModel{
		defaultInterviewDialoguesModel: newInterviewDialoguesModel(conn, c, opts...),
	}
}
