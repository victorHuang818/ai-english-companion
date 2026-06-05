package logic

import (
	"context"
	"fmt"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/core/internal/svc"
	"ai_interview/rpc/core/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInterviewContextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInterviewContextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInterviewContextLogic {
	return &GetInterviewContextLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 面试上下文查询 (供 BFF WebSocket 阶段一次性调用，绕过流式链路)
func (l *GetInterviewContextLogic) GetInterviewContext(in *core.GetInterviewContextReq) (*core.GetInterviewContextResp, error) {
	// 1. 查询面试会话，获取 resume_id 和 job_profile_id
	session, err := l.svcCtx.InterviewSessionModel.FindOne(l.ctx, in.SessionId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("session not found: %s", in.SessionId)
		}
		l.Errorf("GetInterviewContext FindSession error: %v", err)
		return nil, err
	}

	// 2. 并发查询简历和岗位信息 (两者互不依赖，可以并发)
	type resumeResult struct {
		resume *model.Resumes
		err    error
	}
	type jobResult struct {
		job *model.JobProfiles
		err error
	}

	resumeCh := make(chan resumeResult, 1)
	jobCh := make(chan jobResult, 1)

	go func() {
		resume, err := l.svcCtx.ResumeModel.FindOne(l.ctx, session.ResumeId)
		resumeCh <- resumeResult{resume, err}
	}()

	go func() {
		job, err := l.svcCtx.JobProfileModel.FindOne(l.ctx, session.JobProfileId)
		jobCh <- jobResult{job, err}
	}()

	resumeRes := <-resumeCh
	jobRes := <-jobCh

	if resumeRes.err != nil {
		l.Errorf("GetInterviewContext FindResume error: %v", resumeRes.err)
		return nil, resumeRes.err
	}
	if jobRes.err != nil {
		l.Errorf("GetInterviewContext FindJobProfile error: %v", jobRes.err)
		return nil, jobRes.err
	}

	// 3. 组装上下文并返回
	return &core.GetInterviewContextResp{
		SessionId:       in.SessionId,
		ResumeContent:   resumeRes.resume.Content,
		JobProfileName:  jobRes.job.Name,
		JobProfileDesc:  jobRes.job.Description.String,
		UserId:          session.UserId,
	}, nil
}
