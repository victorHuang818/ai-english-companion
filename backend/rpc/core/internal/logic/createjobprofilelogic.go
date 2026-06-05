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

type CreateJobProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateJobProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateJobProfileLogic {
	return &CreateJobProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Job Profiles
func (l *CreateJobProfileLogic) CreateJobProfile(in *core.CreateJobProfileReq) (*core.CreateJobProfileResp, error) {
	id := uuid.New().String()
	_, err := l.svcCtx.JobProfileModel.Insert(l.ctx, &model.JobProfiles{
		Id:        id,
		CreatorId: in.CreatorId,
		Name:      in.Name,
		Description: sql.NullString{
			String: in.Description,
			Valid:  in.Description != "",
		},
	})
	if err != nil {
		l.Errorf("CreateJobProfile error: %v", err)
		return nil, err
	}

	return &core.CreateJobProfileResp{
		Id: id,
	}, nil
}

