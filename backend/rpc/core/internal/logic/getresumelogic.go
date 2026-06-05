package logic

import (
	"context"
	"fmt"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"
	"ai_interview/rpc/core/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetResumeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResumeLogic {
	return &GetResumeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetResumeLogic) GetResume(in *core.GetResumeReq) (*core.GetResumeResp, error) {
	resume, err := l.svcCtx.ResumeModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("resume not found")
		}
		return nil, err
	}

	pdfUrlStr := ""
	if resume.PdfUrl.Valid {
		pdfUrlStr = resume.PdfUrl.String
	}

	return &core.GetResumeResp{
		Id:      resume.Id,
		UserId:  resume.UserId,
		Content: resume.Content,
		PdfUrl:  pdfUrlStr, // 返回原始的 Object Key
	}, nil
}
