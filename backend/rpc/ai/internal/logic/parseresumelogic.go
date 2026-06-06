package logic

import (
	"context"
	"errors"

	"ai_companion/rpc/ai/ai"
	"ai_companion/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ParseResumeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewParseResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ParseResumeLogic {
	return &ParseResumeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ParseResume is deprecated and legacy. PDF resume parsing is now done asynchronously 
// via conductor-oss/markitdown in the core service.
func (l *ParseResumeLogic) ParseResume(in *ai.ParseResumeReq) (*ai.ParseResumeResp, error) {
	l.Infof("ParseResume RPC endpoint was called, but it is deprecated and legacy.")
	return nil, errors.New("ParseResume is deprecated, use the core service with markitdown instead")
}
