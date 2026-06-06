package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DialoguesModel = (*customDialoguesModel)(nil)

type (
	// DialoguesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDialoguesModel.
	DialoguesModel interface {
		dialoguesModel
		InsertWithTime(ctx context.Context, data *Dialogues) (sql.Result, error)
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

func (m *customDialoguesModel) InsertWithTime(ctx context.Context, data *Dialogues) (sql.Result, error) {
	dialoguesIdKey := fmt.Sprintf("%s%v", cacheDialoguesIdPrefix, data.Id)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (id, practice_session_id, role, audio_url, content, evaluation, created_at) values (?, ?, ?, ?, ?, ?, ?)", m.table)
		return conn.ExecCtx(ctx, query, data.Id, data.PracticeSessionId, data.Role, data.AudioUrl, data.Content, data.Evaluation, data.CreatedAt)
	}, dialoguesIdKey)
	return ret, err
}
