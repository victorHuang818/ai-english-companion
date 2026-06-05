package logic

import (
	"context"
	"fmt"

	"ai_interview/rpc/ai/ai"
	"ai_interview/rpc/ai/internal/svc"

	"github.com/sashabaranov/go-openai"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetAiSuggestionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAiSuggestionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAiSuggestionLogic {
	return &GetAiSuggestionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAiSuggestionLogic) GetAiSuggestion(in *ai.GetAiSuggestionReq) (*ai.GetAiSuggestionResp, error) {
	cfg := l.svcCtx.Config.AiSuggestion
	if cfg.ApiKey == "" {
		return nil, fmt.Errorf("AiSuggestion ApiKey is not configured")
	}

	// 1. 初始化 OpenAI 兼容客户端
	config := openai.DefaultConfig(cfg.ApiKey)
	if cfg.BaseUrl != "" {
		config.BaseURL = cfg.BaseUrl
	}
	client := openai.NewClientWithConfig(config)

	// 2. 执行请求 (BFF 已完成模板注入)
	resp, err := client.CreateChatCompletion(
		l.ctx,
		openai.ChatCompletionRequest{
			Model: cfg.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: in.Context,
				},
			},
		},
	)

	if err != nil {
		l.Errorf("AI Suggestion API error (Vendor: %s, Model: %s): %v", cfg.Vendor, cfg.Model, err)
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI generated no suggestion")
	}

	return &ai.GetAiSuggestionResp{
		Suggestion: resp.Choices[0].Message.Content,
	}, nil
}

