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
	startTime := time.Now()

	// 结构化入口日志：type=evaluation，只记录音频大小，不记录字节内容
	logx.Infow("[evaluation] starting AI commentary evaluation",
		logx.Field("type", "evaluation"),
		logx.Field("session_id", in.SessionId),
		logx.Field("model", modelName),
		logx.Field("audio_bytes", len(in.AudioContent)),
		logx.Field("has_audio", len(in.AudioContent) > 0),
	)

	// 🌟 1. 如果包含了音频字节数据，使用符合 OpenAI / DashScope 标准的 HTTP 接口发送语音多模态请求
	// 这样可以彻底避免 go-openai SDK 库版本落后、不支持 "input_audio" 结构体的编译问题
	if len(in.AudioContent) > 0 {
		audioBase64 := "data:audio/wav;base64," + base64.StdEncoding.EncodeToString(in.AudioContent)

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

		// 使用独立 context，避免 go-zero RPC 框架超时 cancel 传播到 HTTP 层
		httpCtx, httpCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer httpCancel()
		req, err := http.NewRequestWithContext(httpCtx, "POST", apiURL, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create http request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.ApiKey))

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logx.Errorw("[evaluation] multimodal HTTP request failed",
				logx.Field("type", "evaluation"),
				logx.Field("session_id", in.SessionId),
				logx.Field("model", modelName),
				logx.Field("duration_ms", time.Since(startTime).Milliseconds()),
				logx.Field("error", err.Error()),
			)
			return nil, err
		}
		defer resp.Body.Close()

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read AI response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			logx.Errorw("[evaluation] API returned non-200 status",
				logx.Field("type", "evaluation"),
				logx.Field("session_id", in.SessionId),
				logx.Field("model", modelName),
				logx.Field("http_status", resp.StatusCode),
				logx.Field("duration_ms", time.Since(startTime).Milliseconds()),
			)
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

		logx.Infow("[evaluation] multimodal commentary completed",
			logx.Field("type", "evaluation"),
			logx.Field("session_id", in.SessionId),
			logx.Field("model", modelName),
			logx.Field("duration_ms", time.Since(startTime).Milliseconds()),
			logx.Field("result", "ok"),
		)
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

	// 使用独立 context，避免 go-zero RPC 框架超时 cancel 传播到 HTTP 层
	textCtx, textCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer textCancel()
	resp, err := client.CreateChatCompletion(
		textCtx,
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
		logx.Errorw("[evaluation] text-only commentary failed",
			logx.Field("type", "evaluation"),
			logx.Field("session_id", in.SessionId),
			logx.Field("model", modelName),
			logx.Field("duration_ms", time.Since(startTime).Milliseconds()),
			logx.Field("error", err.Error()),
		)
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI generated no commentary")
	}

	logx.Infow("[evaluation] text-only commentary completed",
		logx.Field("type", "evaluation"),
		logx.Field("session_id", in.SessionId),
		logx.Field("model", modelName),
		logx.Field("duration_ms", time.Since(startTime).Milliseconds()),
		logx.Field("result", "ok"),
	)
	return &ai.GetAiCommentaryResp{
		Commentary: resp.Choices[0].Message.Content,
	}, nil
}
