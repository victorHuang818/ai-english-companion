package logic

import (
	"context"
	"time"

	"ai_interview/interview_api/internal/svc"
	"ai_interview/interview_api/internal/types"
	"ai_interview/rpc/core/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取简历详情
func NewGetResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResumeLogic {
	return &GetResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetResumeLogic) GetResume(req *types.GetResumeReq) (resp *types.GetResumeResp, err error) {
	rpcResp, err := l.svcCtx.CoreRpc.GetResume(l.ctx, &core.GetResumeReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}

	downloadUrl := ""
	if rpcResp.PdfUrl != "" {
		expiry := time.Minute * 5
		urlStr, err := l.svcCtx.OssClient.GeneratePresignedDownloadURL(l.ctx, rpcResp.PdfUrl, expiry)
		if err != nil {
			l.Errorf("Failed to generate presigned download URL for resume: %v", err)
		} else {
			downloadUrl = urlStr
		}
	}

	status := "parsed"
	rawText := rpcResp.Content
	if rpcResp.Content == "parsing" {
		status = "parsing"
		rawText = ""
	}

	return &types.GetResumeResp{
		Id:          rpcResp.Id,
		RawText:     rawText,
		Status:      status,
		CreatedAt:   0, // 后续如果 DB 有时间可以加上
		DownloadUrl: downloadUrl,
	}, nil
}
