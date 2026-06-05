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

// Resume & OSS
func (l *GenerateUploadUrlLogic) GenerateUploadUrl(in *core.GenerateUploadUrlReq) (*core.GenerateUploadUrlResp, error) {
	// 1. 生成唯一的文件名 (OSS Object Key)，防止同名覆盖
	ext := path.Ext(in.Filename)
	if ext == "" {
		ext = ".pdf" // 默认后缀
	}
	objectKey := fmt.Sprintf("resumes/%s%s", uuid.New().String(), ext)

	// 2. 使用 OSS 通用工具包生成预签名 PUT URL (有效期 15 分钟)
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
