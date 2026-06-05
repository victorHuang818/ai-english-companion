package logic

import (
	"context"
	"database/sql"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"
	"ai_interview/rpc/core/model"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateScenarioLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateScenarioLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateScenarioLogic {
	return &CreateScenarioLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Scenarios
func (l *CreateScenarioLogic) CreateScenario(in *core.CreateScenarioReq) (*core.CreateScenarioResp, error) {
	id := uuid.New().String()
	_, err := l.svcCtx.ScenarioModel.Insert(l.ctx, &model.Scenarios{
		Id:        id,
		CreatorId: in.CreatorId,
		Name:      in.Name,
		Description: sql.NullString{
			String: in.Description,
			Valid:  in.Description != "",
		},
	})
	if err != nil {
		l.Errorf("CreateScenario error: %v", err)
		return nil, err
	}

	return &core.CreateScenarioResp{
		Id: id,
	}, nil
}
