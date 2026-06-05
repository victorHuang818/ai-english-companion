package logic

import (
	"context"
	"fmt"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"
	"ai_interview/rpc/core/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPracticeContextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPracticeContextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPracticeContextLogic {
	return &GetPracticeContextLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPracticeContextLogic) GetPracticeContext(in *core.GetPracticeContextReq) (*core.GetPracticeContextResp, error) {
	// 1. 查询会话
	session, err := l.svcCtx.PracticeSessionModel.FindOne(l.ctx, in.SessionId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("session not found: %s", in.SessionId)
		}
		l.Errorf("GetPracticeContext FindSession error: %v", err)
		return nil, err
	}

	// 2. 并发查询用户档案和场景信息
	type profileResult struct {
		profile *model.UserProfiles
		err     error
	}
	type scenarioResult struct {
		scenario *model.Scenarios
		err      error
	}

	profileCh := make(chan profileResult, 1)
	scenarioCh := make(chan scenarioResult, 1)

	go func() {
		profile, err := l.svcCtx.UserProfileModel.FindOne(l.ctx, session.UserProfileId)
		profileCh <- profileResult{profile, err}
	}()

	go func() {
		scenario, err := l.svcCtx.ScenarioModel.FindOne(l.ctx, session.ScenarioId)
		scenarioCh <- scenarioResult{scenario, err}
	}()

	profileRes := <-profileCh
	scenarioRes := <-scenarioCh

	if profileRes.err != nil {
		l.Errorf("GetPracticeContext FindProfile error: %v", profileRes.err)
		return nil, profileRes.err
	}
	if scenarioRes.err != nil {
		l.Errorf("GetPracticeContext FindScenario error: %v", scenarioRes.err)
		return nil, scenarioRes.err
	}

	// 3. 组装上下文并返回
	return &core.GetPracticeContextResp{
		SessionId:      in.SessionId,
		EnglishLevel:   profileRes.profile.EnglishLevel,
		LearningTarget: profileRes.profile.LearningTarget,
		ScenarioName:   scenarioRes.scenario.Name,
		ScenarioDesc:   scenarioRes.scenario.Description.String,
		UserId:         session.UserId,
	}, nil
}
