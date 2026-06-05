package logic

import (
	"context"
	"database/sql"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSessionStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSessionStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSessionStatusLogic {
	return &UpdateSessionStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateSessionStatusLogic) UpdateSessionStatus(in *core.UpdateSessionStatusReq) (*core.UpdateSessionStatusResp, error) {
	session, err := l.svcCtx.InterviewSessionModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Errorf("Failed to find session %s: %v", in.Id, err)
		return nil, err
	}

	session.Status = in.Status
	if in.EvaluationReport != "" {
		session.EvaluationReport = sql.NullString{
			String: in.EvaluationReport,
			Valid:  true,
		}
	}

	err = l.svcCtx.InterviewSessionModel.Update(l.ctx, session)
	if err != nil {
		l.Errorf("Failed to update session %s: %v", in.Id, err)
		return nil, err
	}

	return &core.UpdateSessionStatusResp{Success: true}, nil
}
