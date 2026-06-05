// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"ai_interview/interview_api/internal/svc"
	"ai_interview/interview_api/internal/types"
	"ai_interview/rpc/core/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateUploadUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取上传简历的预签名URL
func NewGenerateUploadUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateUploadUrlLogic {
	return &GenerateUploadUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateUploadUrlLogic) GenerateUploadUrl(req *types.GenerateUploadUrlReq) (resp *types.GenerateUploadUrlResp, err error) {
	rpcResp, err := l.svcCtx.CoreRpc.GenerateUploadUrl(l.ctx, &core.GenerateUploadUrlReq{
		Filename: req.Filename,
	})
	if err != nil {
		return nil, err
	}

	return &types.GenerateUploadUrlResp{
		UploadUrl: rpcResp.UploadUrl,
		ObjectKey: rpcResp.ObjectKey,
	}, nil
}

