package logic

import (
	"context"
	"fmt"
	"path"
	"time"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateUploadUrlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateUploadUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateUploadUrlLogic {
	return &GenerateUploadUrlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// User Profile & OSS
func (l *GenerateUploadUrlLogic) GenerateUploadUrl(in *core.GenerateUploadUrlReq) (*core.GenerateUploadUrlResp, error) {
	ext := path.Ext(in.Filename)
	if ext == "" {
		ext = ".pdf"
	}
	objectKey := fmt.Sprintf("profiles/%s%s", uuid.New().String(), ext)

	expiry := time.Minute * 15
	presignedURL, err := l.svcCtx.OssClient.GeneratePresignedUploadURL(l.ctx, objectKey, expiry)
	if err != nil {
		l.Errorf("Failed to generate presigned upload url: %v", err)
		return nil, err
	}

	return &core.GenerateUploadUrlResp{
		UploadUrl: presignedURL,
		ObjectKey: objectKey,
	}, nil
}
