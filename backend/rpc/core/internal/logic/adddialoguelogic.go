package logic

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"
	"ai_interview/rpc/core/model"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
)

type AddDialogueLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddDialogueLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddDialogueLogic {
	return &AddDialogueLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Interview Dialogues
func (l *AddDialogueLogic) AddDialogue(in *core.AddDialogueReq) (*core.AddDialogueResp, error) {
	dialogueID := uuid.New().String()

	audioUrl := sql.NullString{Valid: false}
	if in.AudioObjectKey != "" {
		audioUrl = sql.NullString{String: in.AudioObjectKey, Valid: true}
	}

	// 🌟 如果传入了原始音频字节，代表需要直接将音频合并文件存储至 MinIO 并绑定落库
	if len(in.AudioContent) > 0 {
		objectKey := fmt.Sprintf("audios/%s.wav", uuid.New().String())
		_, err := l.svcCtx.MinioClient.PutObject(
			l.ctx,
			l.svcCtx.Config.Minio.BucketName,
			objectKey,
			bytes.NewReader(in.AudioContent),
			int64(len(in.AudioContent)),
			minio.PutObjectOptions{ContentType: "audio/wav"},
		)
		if err != nil {
			l.Errorf("Failed to upload dialogue audio to MinIO: %v", err)
			return nil, err
		}
		audioUrl = sql.NullString{String: objectKey, Valid: true}
	}

	evaluation := sql.NullString{Valid: false}
	if in.Evaluation != "" {
		evaluation = sql.NullString{String: in.Evaluation, Valid: true}
	}

	data := &model.InterviewDialogues{
		Id:                 dialogueID,
		InterviewSessionId: in.SessionId,
		Role:               in.Role,
		AudioUrl:           audioUrl,
		Content:            in.Content,
		Evaluation:         evaluation,
	}

	_, err := l.svcCtx.InterviewDialogueModel.Insert(l.ctx, data)
	if err != nil {
		l.Errorf("Failed to insert dialogue: %v", err)
		return nil, err
	}

	return &core.AddDialogueResp{
		Id: dialogueID,
	}, nil
}
