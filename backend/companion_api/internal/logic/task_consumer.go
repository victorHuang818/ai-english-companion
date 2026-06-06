package logic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"ai_companion/common/prompts"
	"ai_companion/companion_api/internal/svc"
	"ai_companion/rpc/ai/ai"
	"ai_companion/rpc/core/core"
	"ai_companion/rpc/core/coreclient"
	"ai_companion/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Turn struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type CompanionTask struct {
	SessionId    string `json:"session_id"`
	UserId       string `json:"user_id"`
	Question     string `json:"question"`
	Answer       string `json:"answer"`
	WavBase64    string `json:"wav_base64"`
	NextQuestion string `json:"next_question"`
	History      []Turn `json:"history"`
	TotalTokens  int64  `json:"total_tokens"`
}

func StartConsumer(svcCtx *svc.ServiceContext) {
	logx.Infof("Starting Redis Stream task consumer background worker...")
	stream := "english_practice_tasks"
	group := "english_practice_consumers_group"
	consumer := fmt.Sprintf("consumer-%d", time.Now().UnixNano())

	// 1. 确保 Stream 和消费组存在 (使用 go-zero 的 typed 方法)
	_, err := svcCtx.RedisClient.XGroupCreateMkStream(stream, group, "0")
	if err != nil {
		// 忽略重复创建消费组的错误
		logx.Infof("XGroupCreateMkStream info/status: %v (safe to ignore if already exists)", err)
	}

	// 2. 对于阻塞命令 (XREADGROUP with BLOCK)，go-zero 要求创建专用的 BlockingNode 以免耗尽连接池
	node, err := redis.CreateBlockingNode(svcCtx.RedisClient)
	if err != nil {
		logx.Errorf("Failed to create Redis blocking node for consumer: %v", err)
		return
	}

	go func() {
		defer node.Close() // 退出时释放阻塞节点连接
		ctx := context.Background()

		for {
			// 使用 XReadGroup 读取消息
			res, err := svcCtx.RedisClient.XReadGroup(node, group, consumer, 10, 1*time.Second, false, stream, ">")
			if err != nil {
				logx.Errorf("Redis XReadGroup error: %v, sleeping 2s...", err)
				time.Sleep(2 * time.Second)
				continue
			}

			if len(res) == 0 {
				continue
			}

			// 处理消息
			for _, s := range res {
				for _, msg := range s.Messages {
					payload, ok := msg.Values["payload"].(string)
					if ok && payload != "" {
						var task CompanionTask
						if err := json.Unmarshal([]byte(payload), &task); err == nil {
							// 执行异步落库与打分评估
							processTask(ctx, svcCtx, task)
						} else {
							logx.Errorf("Failed to unmarshal task payload: %v", err)
						}
					}

					// 确认消息消费完成
					_, err = svcCtx.RedisClient.XAck(stream, group, msg.ID)
					if err != nil {
						logx.Errorf("Failed to XACK message %s: %v", msg.ID, err)
					}
				}
			}
		}
	}()
}

func processTask(ctx context.Context, svcCtx *svc.ServiceContext, task CompanionTask) {
	logx.Infof("Processing companion task asynchronously for session: %s", task.SessionId)

	// Decode WAV bytes from Base64
	var wav []byte
	if task.WavBase64 != "" {
		var err error
		wav, err = base64.StdEncoding.DecodeString(task.WavBase64)
		if err != nil {
			logx.Errorf("Failed to decode audio Base64: %v", err)
		}
	}

	// 1. 调用 Core RPC 统一写入用户回答记录并上传音频至 MinIO
	var userResp *core.AddDialogueResp
	var err error
	if task.Answer != "" {
		userResp, err = svcCtx.CoreRpc.AddDialogue(ctx, &coreclient.AddDialogueReq{
			SessionId:    task.SessionId,
			Role:         "user",
			Content:      task.Answer,
			AudioContent: wav,
			Evaluation:   "evaluating",
		})
		if err != nil {
			logx.Errorf("Consumer failed to save user dialogue: %v", err)
		}
	}

	// 2. 紧接着立即写入面试官的新提问
	if task.NextQuestion != "" {
		_, err = svcCtx.CoreRpc.AddDialogue(ctx, &coreclient.AddDialogueReq{
			SessionId: task.SessionId,
			Role:      "interviewer",
			Content:   task.NextQuestion,
		})
		if err != nil {
			logx.Errorf("Consumer failed to save companion teacher dialogue: %v", err)
		}
	}

	// 3. 在后台调用大模型进行评估 (GetAiCommentary)
	if userResp != nil && userResp.Id != "" {
		go func(dialogueId string) {
			// 获取简历与岗位上下文
			ctxResp, err := svcCtx.CoreRpc.GetPracticeContext(context.Background(), &coreclient.GetPracticeContextReq{
				SessionId: task.SessionId,
			})
			if err != nil {
				logx.Errorf("Consumer failed to get practice context: %v", err)
				return
			}

			// 格式化前两轮的对话上下文
			var historyBuilder strings.Builder
			for idx, turn := range task.History {
				historyBuilder.WriteString(fmt.Sprintf("回合 %d:\n", idx+1))
				historyBuilder.WriteString(fmt.Sprintf("  陪练老师提问: %s\n", turn.Question))
				historyBuilder.WriteString(fmt.Sprintf("  用户回答: %s\n", turn.Answer))
			}
			historyStr := historyBuilder.String()

			// 组装系统 Prompt
			userProfileContext := fmt.Sprintf("English Level: %s\nLearning Target: %s", ctxResp.EnglishLevel, ctxResp.LearningTarget)
			commentaryPrompt := fmt.Sprintf(
				prompts.AiCommentatorPrompt,
				userProfileContext,
				ctxResp.ScenarioName,
				ctxResp.ScenarioDesc,
				historyStr,
				task.Question,
				"【当前用户回答已在附带音频中传入，请直接分析并评估用户的语音表达、自信度、流畅度以及回答内容本身】",
			)

			var commentary string
			var rpcErr error
			if svcCtx.Config.Mock {
				// 模拟计算耗时
				time.Sleep(50 * time.Millisecond)
				commentary = `{
					"dimensions": {
						"fluency": "口语表达流畅度较好，表达自然且逻辑连贯。",
						"relevance": "回答切题度良好，重点抓得准。",
						"logic": "逻辑严密，采用了总分结构进行回答。",
						"depth": "专业词汇掌握扎实，展现出了较深的技术经验。"
					},
					"scores": {
						"fluency": 85,
						"vocabulary": 88,
						"grammar": 80,
						"pronunciation": 90
					}
				}`
			} else {
				commentaryCtx, commentaryCancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer commentaryCancel()
				commentaryResp, err := svcCtx.AiRpc.GetAiCommentary(commentaryCtx, &ai.GetAiCommentaryReq{
					SessionId:    task.SessionId,
					Context:      commentaryPrompt,
					AudioContent: wav,
				})
				if err == nil && commentaryResp != nil {
					commentary = commentaryResp.Commentary
				} else {
					rpcErr = err
				}
			}

			if rpcErr == nil && commentary != "" {
				// Clean the commentary to remove any lists or bullet points
				cleanedCommentary := cleanDialogueEvaluationJSON(commentary)
				// 评估结果返回后，异步更新用户回答的 evaluation 字段
				_, err = svcCtx.CoreRpc.UpdateDialogueEvaluation(context.Background(), &coreclient.UpdateDialogueEvaluationReq{
					DialogueId: dialogueId,
					Evaluation: cleanedCommentary,
				})
				if err != nil {
					logx.Errorf("Consumer failed to update dialogue evaluation: %v", err)
				}
			} else {
				errMsg := "Unknown error"
				if rpcErr != nil {
					logx.Errorf("Consumer GetAiCommentary RPC error: %v", rpcErr)
					errMsg = rpcErr.Error()
				} else if commentary == "" {
					logx.Errorf("Consumer GetAiCommentary returned empty commentary")
					errMsg = "Empty response from AI"
				}

				// 🌟 写入失败时的降级 JSON，避免状态永久卡在 "evaluating"
				fallbackObj := map[string]interface{}{
					"dimensions": map[string]string{
						"fluency":        fmt.Sprintf("评估请求失败: %s", errMsg),
						"relevance":      "评估请求失败",
						"logic":          "评估请求失败",
						"depth":          "评估请求失败",
					},
					"scores": map[string]int{
						"fluency":       0,
						"vocabulary":    0,
						"grammar":       0,
						"pronunciation": 0,
					},
				}
				fallbackBytes, _ := json.Marshal(fallbackObj)
				fallbackCommentary := string(fallbackBytes)

				_, err = svcCtx.CoreRpc.UpdateDialogueEvaluation(context.Background(), &coreclient.UpdateDialogueEvaluationReq{
					DialogueId: dialogueId,
					Evaluation: fallbackCommentary,
				})
				if err != nil {
					logx.Errorf("Consumer failed to update fallback dialogue evaluation: %v", err)
				}
			}
		}(userResp.Id)
	}

	// 4. 用户 Token 额度扣减
	if task.TotalTokens > 0 {
		go func(tokens int64) {
			_, err = svcCtx.UserRpc.DeductToken(context.Background(), &user.DeductTokenReq{
				Id:     task.UserId,
				Amount: uint32(tokens),
			})
			if err != nil {
				logx.Errorf("Consumer failed to deduct token: %v", err)
			}
		}(task.TotalTokens)
	}
}

func cleanParagraph(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// Remove markers at the very beginning of the string
	reStart := regexp.MustCompile(`^\s*([-[*•]|\d+[\.)]|\(\d+\))\s*`)
	text = reStart.ReplaceAllString(text, "")

	// Remove markers in the middle of the string
	reMiddle := regexp.MustCompile(`\s*([-[*•]|\d+[\.)]|\(\d+\))\s+`)
	text = reMiddle.ReplaceAllString(text, " ")

	// Normalize spacing
	words := strings.Fields(text)
	text = strings.Join(words, " ")

	return text
}

func cleanDialogueEvaluationJSON(jsonStr string) string {
	cleanJSON := strings.ReplaceAll(jsonStr, "```json", "")
	cleanJSON = strings.ReplaceAll(cleanJSON, "```", "")
	cleanJSON = strings.TrimSpace(cleanJSON)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(cleanJSON), &data); err != nil {
		return jsonStr // return original if not parseable
	}

	// Clean dimensions
	if dims, ok := data["dimensions"].(map[string]interface{}); ok {
		for k, v := range dims {
			if strVal, ok := v.(string); ok {
				dims[k] = cleanParagraph(strVal)
			}
		}
	}

	// Clean overall_comment if it exists
	if comment, ok := data["overall_comment"].(string); ok {
		data["overall_comment"] = cleanParagraph(comment)
	}

	// Re-marshal to JSON string
	if cleanedBytes, err := json.Marshal(data); err == nil {
		return string(cleanedBytes)
	}
	return jsonStr
}
