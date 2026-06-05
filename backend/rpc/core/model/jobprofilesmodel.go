package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ JobProfilesModel = (*customJobProfilesModel)(nil)

type (
	// JobProfilesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customJobProfilesModel.
	JobProfilesModel interface {
		jobProfilesModel
	}

	customJobProfilesModel struct {
		*defaultJobProfilesModel
	}
)

// NewJobProfilesModel returns a model for the database table.
func NewJobProfilesModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) JobProfilesModel {
	return &customJobProfilesModel{
		defaultJobProfilesModel: newJobProfilesModel(conn, c, opts...),
	}
}
