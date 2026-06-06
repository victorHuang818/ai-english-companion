package logic

import (
	"context"

	"ai_companion/rpc/core/core"
	"ai_companion/rpc/core/internal/svc"

	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSessionLogic {
	return &DeleteSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteSessionLogic) DeleteSession(in *core.DeleteSessionReq) (*core.DeleteSessionResp, error) {
	// 1. Query all dialogues with non-empty audio_url to delete from MinIO
	var audioUrls []string
	queryFindAudios := "SELECT audio_url FROM dialogues WHERE practice_session_id = ? AND audio_url IS NOT NULL AND audio_url != ''"
	err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &audioUrls, queryFindAudios, in.SessionId)
	if err != nil {
		l.Errorf("Failed to query dialogue audio URLs for session %s: %v", in.SessionId, err)
	}

	// 2. Delete the audio objects from MinIO
	for _, objectKey := range audioUrls {
		if objectKey != "" {
			err := l.svcCtx.MinioClient.RemoveObject(
				l.ctx,
				l.svcCtx.Config.Minio.BucketName,
				objectKey,
				minio.RemoveObjectOptions{},
			)
			if err != nil {
				l.Errorf("Failed to remove MinIO object %s for session %s: %v", objectKey, in.SessionId, err)
			} else {
				l.Infof("Successfully removed MinIO object %s", objectKey)
			}
		}
	}

	// 3. Delete all dialogues under this session first
	queryDialogues := "DELETE FROM dialogues WHERE practice_session_id = ?"
	_, err = l.svcCtx.SqlConn.ExecCtx(l.ctx, queryDialogues, in.SessionId)
	if err != nil {
		l.Errorf("Failed to delete dialogues for session %s: %v", in.SessionId, err)
		return nil, err
	}

	// 4. Delete the session itself
	err = l.svcCtx.PracticeSessionModel.Delete(l.ctx, in.SessionId)
	if err != nil {
		l.Errorf("Failed to delete session %s: %v", in.SessionId, err)
		return nil, err
	}

	return &core.DeleteSessionResp{Success: true}, nil
}
