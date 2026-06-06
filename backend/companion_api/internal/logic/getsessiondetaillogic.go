package logic

import (
	"context"

	"ai_companion/companion_api/internal/svc"
	"ai_companion/companion_api/internal/types"
	"ai_companion/rpc/core/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSessionDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取口语练习详情
func NewGetSessionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSessionDetailLogic {
	return &GetSessionDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSessionDetailLogic) GetSessionDetail(req *types.GetSessionDetailReq) (resp *types.GetSessionDetailResp, err error) {
	rpcResp, err := l.svcCtx.CoreRpc.GetSessionDetail(l.ctx, &coreclient.GetSessionDetailReq{
		SessionId: req.SessionId,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetSessionDetailResp{
		SessionId:    rpcResp.SessionId,
		ScenarioName: rpcResp.ScenarioName,
		Transcript:   rpcResp.Transcript,
		OverallScore: rpcResp.OverallScore,
		Commentary:   rpcResp.EvaluationReport,
	}, nil
}
