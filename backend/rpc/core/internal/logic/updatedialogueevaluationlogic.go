package logic

import (
	"context"
	"database/sql"
	"fmt"

	"ai_companion/rpc/core/core"
	"ai_companion/rpc/core/internal/svc"
	"ai_companion/rpc/core/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDialogueEvaluationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDialogueEvaluationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDialogueEvaluationLogic {
	return &UpdateDialogueEvaluationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDialogueEvaluationLogic) UpdateDialogueEvaluation(in *core.UpdateDialogueEvaluationReq) (*core.UpdateDialogueEvaluationResp, error) {
	dialogue, err := l.svcCtx.DialogueModel.FindOne(l.ctx, in.DialogueId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("dialogue not found: %s", in.DialogueId)
		}
		l.Errorf("Failed to find dialogue %s: %v", in.DialogueId, err)
		return nil, err
	}

	dialogue.Evaluation = sql.NullString{
		String: in.Evaluation,
		Valid:  true,
	}

	err = l.svcCtx.DialogueModel.Update(l.ctx, dialogue)
	if err != nil {
		l.Errorf("Failed to update dialogue evaluation for %s: %v", in.DialogueId, err)
		return nil, err
	}

	return &core.UpdateDialogueEvaluationResp{
		Success: true,
	}, nil
}
