package logic

import (
	"context"
	"fmt"
	"time"

	"ai_companion/rpc/ai/ai"
	"ai_companion/rpc/ai/internal/svc"

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

	startTime := time.Now()

	// 结构化入口日志：type=hint
	logx.Infow("[hint] starting AI suggestion generation",
		logx.Field("type", "hint"),
		logx.Field("session_id", in.SessionId),
		logx.Field("model", cfg.Model),
	)

	// 1. 初始化 OpenAI 兼容客户端
	config := openai.DefaultConfig(cfg.ApiKey)
	if cfg.BaseUrl != "" {
		config.BaseURL = cfg.BaseUrl
	}
	client := openai.NewClientWithConfig(config)

	// 2. 执行请求 (BFF 已完成模板注入)
	// 使用独立 context，避免 go-zero RPC 框架的超时 cancel 传播到 HTTP 层
	httpCtx, httpCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer httpCancel()

	resp, err := client.CreateChatCompletion(
		httpCtx,
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
		logx.Errorw("[hint] AI suggestion failed",
			logx.Field("type", "hint"),
			logx.Field("session_id", in.SessionId),
			logx.Field("model", cfg.Model),
			logx.Field("duration_ms", time.Since(startTime).Milliseconds()),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI generated no suggestion")
	}

	logx.Infow("[hint] AI suggestion completed",
		logx.Field("type", "hint"),
		logx.Field("session_id", in.SessionId),
		logx.Field("model", cfg.Model),
		logx.Field("duration_ms", time.Since(startTime).Milliseconds()),
		logx.Field("result", "ok"),
	)
	return &ai.GetAiSuggestionResp{
		Suggestion: resp.Choices[0].Message.Content,
	}, nil
}
