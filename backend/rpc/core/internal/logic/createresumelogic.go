package logic

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"
	"ai_interview/rpc/core/model"

	"github.com/conductor-oss/markitdown"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateResumeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateResumeLogic {
	return &CreateResumeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateResumeLogic) CreateResume(in *core.CreateResumeReq) (*core.CreateResumeResp, error) {
	// 1. 验证文件在 OSS 中是否存在
	exists, err := l.svcCtx.OssClient.FileExists(l.ctx, in.ObjectKey)
	if err != nil {
		l.Errorf("Failed to check if file exists in OSS: %v", err)
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("file does not exist in OSS/MinIO: %s", in.ObjectKey)
	}

	// 2. 生成简历唯一标识 (UUID)
	resumeID := uuid.New().String()

	// 3. 构建初始 "parsing" 状态内容
	initialContent := "parsing"

	// 4. 立刻在数据库创建一条记录 (PdfUrl 字段存的是 object_key)
	data := &model.Resumes{
		Id:      resumeID,
		UserId:  in.UserId,
		Content: initialContent,
		PdfUrl: sql.NullString{
			String: in.ObjectKey,
			Valid:  true,
		},
	}

	_, err = l.svcCtx.ResumeModel.Insert(l.ctx, data)
	if err != nil {
		l.Errorf("Failed to insert initial resume to database: %v", err)
		return nil, err
	}

	// 5. 开启异步硬核解析 Goroutine (后台默默干活)
	go func(resID, userID, objKey string, initialData *model.Resumes) {
		bgCtx := context.Background()
		logx.WithContext(bgCtx).Infof("Starting asynchronous resume parsing for resume %s (OSS Key: %s)", resID, objKey)

		// 5.1 从 OSS 拉取文件二进制数据
		pdfBytes, err := l.svcCtx.OssClient.DownloadFile(bgCtx, objKey)
		if err != nil {
			logx.WithContext(bgCtx).Errorf("Failed to download PDF from OSS for resume %s: %v", resID, err)
			// 更新内容为失败状态描述
			initialData.Content = fmt.Sprintf("# Resume Parsing Failed\n\nFailed to download resume file from OSS key: `%s`. Error: %v", objKey, err)
			_ = l.svcCtx.ResumeModel.Update(bgCtx, initialData)
			return
		}

		// 5.2 使用 conductor-oss/markitdown 进行 PDF 转 Markdown 解析
		m := markitdown.New()
		reader := bytes.NewReader(pdfBytes)
		res, err := m.ConvertReader(reader, markitdown.StreamInfo{
			Extension: ".pdf",
			MIMEType:  "application/pdf",
		})

		var parsedMarkdown string
		if err != nil {
			logx.WithContext(bgCtx).Errorf("MarkItDown parsing error: %v, falling back to basic extraction", err)
			// 备用兜底内容
			parsedMarkdown = fmt.Sprintf("# Candidate Resume (Basic Extracted)\n\n- **User ID**: %s\n- **OSS Key**: %s\n\n*(Standard MarkItDown parsing failed; metadata extracted.)*", userID, objKey)
		} else {
			parsedMarkdown = res.Markdown
		}

		// 5.3 更新数据库中的 content 为解析后的 Markdown 内容
		initialData.Content = parsedMarkdown
		err = l.svcCtx.ResumeModel.Update(bgCtx, initialData)
		if err != nil {
			logx.WithContext(bgCtx).Errorf("Failed to update database with parsed resume content: %v", err)
		} else {
			logx.WithContext(bgCtx).Infof("Successfully parsed and saved resume %s content (Markdown length: %d)", resID, len(parsedMarkdown))
		}
	}(resumeID, in.UserId, in.ObjectKey, data)

	// 6. 立刻返回给前端 "parsing" 状态响应
	return &core.CreateResumeResp{
		Id:     resumeID,
		Status: "parsing",
	}, nil
}
