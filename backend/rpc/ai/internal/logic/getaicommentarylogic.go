package logic

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"ai_companion/rpc/ai/ai"
	"ai_companion/rpc/ai/internal/svc"

	"github.com/sashabaranov/go-openai"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetAiCommentaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAiCommentaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAiCommentaryLogic {
	return &GetAiCommentaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAiCommentaryLogic) GetAiCommentary(in *ai.GetAiCommentaryReq) (*ai.GetAiCommentaryResp, error) {
	cfg := l.svcCtx.Config.AiCommentator
	if cfg.ApiKey == "" {
		return nil, fmt.Errorf("AiCommentator ApiKey is not configured")
	}

	modelName := cfg.Model

	// 🌟 1. 如果包含了音频字节数据，使用符合 OpenAI / DashScope 标准的 HTTP 接口发送语音多模态请求
	// 这样可以彻底避免 go-openai SDK 库版本落后、不支持 "input_audio" 结构体的编译问题
	if len(in.AudioContent) > 0 {
		audioBase64 := base64.StdEncoding.EncodeToString(in.AudioContent)

		payload := map[string]interface{}{
			"model": modelName,
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": in.Context,
						},
						{
							"type": "input_audio",
							"input_audio": map[string]string{
								"data":   audioBase64,
								"format": "wav",
							},
						},
					},
				},
			},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal multi-modal payload: %v", err)
		}

		baseURL := "https://api.openai.com/v1"
		if cfg.BaseUrl != "" {
			baseURL = cfg.BaseUrl
		}
		apiURL := fmt.Sprintf("%s/chat/completions", baseURL)

		req, err := http.NewRequestWithContext(l.ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create http request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.ApiKey))

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			l.Errorf("Multimodal AI Commentary HTTP request failed (Vendor: %s, Model: %s): %v", cfg.Vendor, modelName, err)
			return nil, err
		}
		defer resp.Body.Close()

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read AI response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			l.Errorf("Multimodal AI Commentary API returned error status %d: %s", resp.StatusCode, string(respBytes))
			return nil, fmt.Errorf("AI API returned status %d", resp.StatusCode)
		}

		// 解析 OpenAI 返回体
		var chatResp struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(respBytes, &chatResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal AI response: %v", err)
		}

		if len(chatResp.Choices) == 0 {
			return nil, fmt.Errorf("multimodal AI generated no commentary")
		}

		return &ai.GetAiCommentaryResp{
			Commentary: chatResp.Choices[0].Message.Content,
		}, nil
	}

	// 🌟 2. 回退逻辑：执行普通的纯文本 Chat 请求 (BFF 已完成模板注入)
	config := openai.DefaultConfig(cfg.ApiKey)
	if cfg.BaseUrl != "" {
		config.BaseURL = cfg.BaseUrl
	}
	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(
		l.ctx,
		openai.ChatCompletionRequest{
			Model: modelName,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: in.Context,
				},
			},
		},
	)

	if err != nil {
		l.Errorf("AI Commentary API error (Vendor: %s, Model: %s): %v", cfg.Vendor, modelName, err)
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI generated no commentary")
	}

	return &ai.GetAiCommentaryResp{
		Commentary: resp.Choices[0].Message.Content,
	}, nil
}
