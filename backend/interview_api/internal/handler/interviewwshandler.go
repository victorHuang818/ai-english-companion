package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"ai_interview/common/prompts"
	"ai_interview/interview_api/internal/logic"
	"ai_interview/interview_api/internal/svc"
	"ai_interview/rpc/ai/ai"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/user/user"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

// sync.Pool 复用音频 slice 内存
var audioBufferPool = sync.Pool{
	New: func() any {
		// 预分配 512KB 大小的缓冲区
		return make([]byte, 0, 512*1024)
	},
}

// sync.Pool 复用 JSON 解析结构体
var realtimeResponsePool = sync.Pool{
	New: func() any {
		return &RealtimeResponse{}
	},
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// RealtimeResponse 统一的实时交互协议响应结构体（桥接并兼容各家大模型底层流式数据格式）
type RealtimeResponse struct {
	ServerContent *struct {
		ModelTurn *struct {
			Parts []struct {
				Text       string `json:"text"`
				InlineData *struct {
					MimeType string `json:"mimeType"`
					Data     string `json:"data"`
				} `json:"inlineData"`
			} `json:"parts"`
		} `json:"modelTurn"`
		InputTranscription *struct {
			Text string `json:"text"`
		} `json:"inputTranscription"`
		OutputTranscription *struct {
			Text string `json:"text"`
		} `json:"outputTranscription"`
		TurnComplete bool `json:"turnComplete"`
	} `json:"serverContent"`
	UsageMetadata *struct {
		TotalTokenCount int64 `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func InterviewWSHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")
		if userId == "" {
			logx.Error("Missing X-User-ID in header")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()

		_, msg, err := conn.ReadMessage()
		if err != nil {
			logx.Errorf("Read first message error: %v", err)
			return
		}

		var firstMsg struct {
			SessionId string `json:"session_id"`
		}
		if err := json.Unmarshal(msg, &firstMsg); err != nil {
			logx.Errorf("Unmarshal session_id error: %v", err)
			return
		}

		ctxResp, err := svcCtx.CoreRpc.GetInterviewContext(r.Context(), &core.GetInterviewContextReq{
			SessionId: firstMsg.SessionId,
		})
		if err != nil {
			logx.Errorf("GetInterviewContext rpc error: %v", err)
			return
		}

		// 🌟 依据配置文件 yaml 的 RealtimeInterviewer.Provider 动态创建适配器，零侵入无缝切流！
		var realtimeConn RealtimeInterviewerConn
		provider := svcCtx.Config.RealtimeInterviewer.Provider
		systemInstruction := fmt.Sprintf(prompts.InterviewerPrompt, ctxResp.JobProfileName, ctxResp.ResumeContent, ctxResp.JobProfileDesc, "暂无对话记录")

		if svcCtx.Config.Mock {
			logx.Infof("Initializing Mock Realtime Interviewer for load testing")
			realtimeConn, err = NewMockRealtimeConn()
		} else if strings.ToLower(provider) == "qwen" {
			logx.Infof("Initializing Qwen-Omni-Realtime Interviewer with model: %s", svcCtx.Config.RealtimeInterviewer.Qwen.Model)
			realtimeConn, err = NewQwenRealtimeConn(
				svcCtx.Config.RealtimeInterviewer.Qwen.ApiKey,
				svcCtx.Config.RealtimeInterviewer.Qwen.Model,
			)
		} else {
			logx.Infof("Initializing Gemini Live Interviewer with model: %s", svcCtx.Config.RealtimeInterviewer.Gemini.Model)
			apiKey := svcCtx.Config.RealtimeInterviewer.Gemini.ApiKey
			model := svcCtx.Config.RealtimeInterviewer.Gemini.Model
			realtimeConn, err = NewGeminiRealtimeConn(apiKey, model)
		}

		if err != nil {
			logx.Errorf("Failed to dial realtime interviewer: %v", err)
			return
		}
		defer realtimeConn.Close()

		// 2. 初始化配置会话
		if err := realtimeConn.Initialize(systemInstruction); err != nil {
			logx.Errorf("Failed to initialize realtime session: %v", err)
			return
		}

		errCh := make(chan error, 2)
		var writeMu sync.Mutex

		var (
			fullInterviewerText string
			fullUserTranscript  string

			// 滑动窗口状态机变量
			currentQuestion string
			currentAnswer   string
			nextQuestion    string
			history         []logic.Turn
			historyMu       sync.Mutex

			// 🌟 音频收集与 AI 语音评估上下文
			audioMu               sync.Mutex
			isFirstResponseOfTurn = true
			activeWavBytes        []byte
		)

		// 从对象池借用音频缓冲区并确保重置长度
		userAudioBufferRaw := audioBufferPool.Get().([]byte)
		userAudioBuffer := userAudioBufferRaw[:0]
		defer func() {
			audioMu.Lock()
			audioBufferPool.Put(userAudioBuffer)
			audioMu.Unlock()
		}()

		// Goroutine A: 前端 -> Realtime AI Interviewer
		go func() {
			for {
				mt, data, err := conn.ReadMessage()
				if err != nil {
					errCh <- err
					return
				}

				if mt == websocket.BinaryMessage {
					// 🌟 收集前端传来的原始 PCM 音频碎片，用于当前回合的完整回答音频合并
					audioMu.Lock()
					userAudioBuffer = append(userAudioBuffer, data...)
					audioMu.Unlock()

					// 使用通用接口发送音频流
					if err := realtimeConn.SendAudioChunk(data); err != nil {
						errCh <- err
						return
					}
				}
			}
		}()

		// Goroutine B: Realtime AI Interviewer -> 前端 (含 AI 辅助提示和滑动窗口 AI 评论员)
		go func() {
			for {
				mt, data, _, _, _, _, err := realtimeConn.ReadMessage()
				if err != nil {
					errCh <- err
					return
				}
				writeMu.Lock()
				conn.WriteMessage(mt, data)
				writeMu.Unlock()

				gResp := realtimeResponsePool.Get().(*RealtimeResponse)
				*gResp = RealtimeResponse{} // 重置字段以防脏数据

				if err := json.Unmarshal(data, gResp); err == nil {
					if gResp.ServerContent != nil {
						// 🌟 捕捉本轮 Live API 响应的第一块数据，意味着用户回答已完毕，立即取出并清空已录制的 PCM 音频并发起语音评估
						historyMu.Lock()
						if isFirstResponseOfTurn {
							isFirstResponseOfTurn = false

							audioMu.Lock()
							var pcmBytes []byte
							if len(userAudioBuffer) > 0 {
								pcmBytes = make([]byte, len(userAudioBuffer))
								copy(pcmBytes, userAudioBuffer)
								userAudioBuffer = userAudioBuffer[:0] // 仅重置长度，保持底层容量以重用内存
							}
							audioMu.Unlock()

							if len(pcmBytes) > 0 {
								// 将 PCM 合并为包含 44 字节标准头部的 WAV 文件字节流
								wavBytes := pcmToWav(pcmBytes, 16000, 1, 16)
								activeWavBytes = wavBytes
							}
						}
						historyMu.Unlock()

						// 1. 收集用户回答文本
						if gResp.ServerContent.InputTranscription != nil {
							text := gResp.ServerContent.InputTranscription.Text
							if text != "" {
								fullUserTranscript += text
								historyMu.Lock()
								currentAnswer += text // 收集用户回答文本碎片，拼接成完整的回答 A_n
								historyMu.Unlock()
							}
						}

						// 2. 收集面试官提问文本
						if gResp.ServerContent.OutputTranscription != nil {
							text := gResp.ServerContent.OutputTranscription.Text
							if text != "" {
								fullInterviewerText += text
								historyMu.Lock()
								nextQuestion += text // 流式拼接下一轮问题 Q_{n+1}
								historyMu.Unlock()
							}
						}

						// 3. 捕捉 TurnComplete 信号，代表当前回合数据流（用户回答 A_n 和新问题 Q_{n+1}已全部完整接收完毕）
						if gResp.ServerContent.TurnComplete {
							historyMu.Lock()

							var suggestionHistoryStr string

							if currentAnswer != "" {
								evaluatedQuestion := currentQuestion
								evaluatedAnswer := currentAnswer
								wavBytes := activeWavBytes

								// 拷贝历史数据切片
								var historyCopy []logic.Turn
								for _, h := range history {
									historyCopy = append(historyCopy, logic.Turn{
										Question: h.Question,
										Answer:   h.Answer,
									})
								}

								// 将当前本轮追加进历史窗口，供后续回合使用
								history = append(history, logic.Turn{Question: evaluatedQuestion, Answer: evaluatedAnswer})
								if len(history) > 2 {
									history = history[len(history)-2:] // 仅保留最近2轮作为滑动窗口
								}

								// 将音频转换为 Base64，以便在 Redis Stream 中进行传输
								var wavBase64 string
								if len(wavBytes) > 0 {
									wavBase64 = base64.StdEncoding.EncodeToString(wavBytes)
								}

								// 组装并发布任务至 Redis Stream 队列，实现高吞吐异步落库与评估
								task := logic.InterviewTask{
									SessionId:    firstMsg.SessionId,
									UserId:       userId,
									Question:     evaluatedQuestion,
									Answer:       evaluatedAnswer,
									WavBase64:    wavBase64,
									NextQuestion: nextQuestion,
									History:      historyCopy,
									TotalTokens:  0,
								}

								taskBytes, err := json.Marshal(task)
								if err == nil {
									_, xerr := svcCtx.RedisClient.XAdd("interview_tasks", false, "*", map[string]string{"payload": string(taskBytes)})
									if xerr != nil {
										logx.Errorf("Failed to XADD task to Redis: %v", xerr)
									} else {
										logx.Infof("Successfully queued interview task in Redis Stream for session %s", firstMsg.SessionId)
									}
								} else {
									logx.Errorf("Failed to marshal interview task: %v", err)
								}
							}

							// 格式化包含本轮最新对话的历史窗口，用于破题锦囊 (AI 辅助提示) 的上下文背景
							var suggestionHistoryBuilder strings.Builder
							for idx, turn := range history {
								suggestionHistoryBuilder.WriteString(fmt.Sprintf("回合 %d:\n", idx+1))
								suggestionHistoryBuilder.WriteString(fmt.Sprintf("  面试官提问: %s\n", turn.Question))
								suggestionHistoryBuilder.WriteString(fmt.Sprintf("  候选人回答: %s\n", turn.Answer))
							}
							suggestionHistoryStr = suggestionHistoryBuilder.String()

							// 在完整接收完新问题后，触发 AI 辅助提示（破题锦囊）
							if nextQuestion != "" {
								go func(nq, h string) {
									suggestionPrompt := fmt.Sprintf(
										prompts.AiSuggestionPrompt,
										ctxResp.JobProfileName,
										ctxResp.JobProfileDesc,
										ctxResp.ResumeContent,
										h,
										nq,
									)

									var suggestion string
									var rpcErr error
									if svcCtx.Config.Mock {
										time.Sleep(30 * time.Millisecond)
										suggestion = "- 阐述你在高并发下的限流设计思路。\n- 使用STAR法则：描述背景S、挑战T、行动A、结果R。"
									} else {
										suggestionResp, err := svcCtx.AiRpc.GetAiSuggestion(context.Background(), &ai.GetAiSuggestionReq{
											SessionId: firstMsg.SessionId,
											Context:   suggestionPrompt,
										})
										if err == nil && suggestionResp != nil {
											suggestion = suggestionResp.Suggestion
										} else {
											rpcErr = err
										}
									}

									if rpcErr != nil {
										logx.Errorf("GetAiSuggestion RPC error: %v", rpcErr)
										return
									}

									if suggestion != "" {
										writeMu.Lock()
										conn.WriteJSON(map[string]interface{}{
											"aiSuggestion": map[string]interface{}{
												"suggestion": suggestion,
											},
										})
										writeMu.Unlock()
									}
								}(nextQuestion, suggestionHistoryStr)
							}

							// 无论是否触发评估，都完成本轮向下一轮问题的交接并重置
							currentQuestion = nextQuestion
							currentAnswer = ""
							nextQuestion = ""

							// 🌟 重置录音标记，为下一轮做准备
							isFirstResponseOfTurn = true
							activeWavBytes = nil

							historyMu.Unlock()
						}
					}
					// 计费逻辑
					if gResp.UsageMetadata != nil {
						totalTokens := gResp.UsageMetadata.TotalTokenCount
						go func(t int64) {
							svcCtx.UserRpc.DeductToken(context.Background(), &user.DeductTokenReq{
								Id:     userId,
								Amount: uint32(t),
							})
						}(totalTokens)
					}
				}
				realtimeResponsePool.Put(gResp)
			}
		}()

		reason := <-errCh
		logx.Infof("WebSocket closed: %v", reason)

		// 🌟 WebSocket 断开时，异步更新数据库中的面试会话状态为已完成 (completed)
		if firstMsg.SessionId != "" {
			go func(sid string) {
				bgCtx := context.Background()
				_, err := svcCtx.CoreRpc.UpdateSessionStatus(bgCtx, &core.UpdateSessionStatusReq{
					Id:     sid,
					Status: "completed",
				})
				if err != nil {
					logx.Errorf("Failed to update session status to completed: %v", err)
				} else {
					logx.Infof("Session %s updated to completed status", sid)
				}
			}(firstMsg.SessionId)
		}
	}
}

// pcmToWav 将原始 PCM 字节流转换为包含标准 44 字节 WAV 头部的完整音频流
// sampleRate: 16000, numChannels: 1 (单声道), bitsPerSample: 16
func pcmToWav(pcmData []byte, sampleRate int, numChannels int, bitsPerSample int) []byte {
	buf := new(bytes.Buffer)

	// RIFF Header
	buf.Write([]byte("RIFF"))
	// File Size (total size of the file minus 8 bytes)
	fileSize := uint32(36 + len(pcmData))
	binary.Write(buf, binary.LittleEndian, fileSize)

	// WAVE Header
	buf.Write([]byte("WAVE"))

	// fmt chunk
	buf.Write([]byte("fmt "))
	chunkSize := uint32(16)
	binary.Write(buf, binary.LittleEndian, chunkSize)

	audioFormat := uint16(1) // PCM = 1
	binary.Write(buf, binary.LittleEndian, audioFormat)

	binary.Write(buf, binary.LittleEndian, uint16(numChannels))
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))

	byteRate := uint32(sampleRate * numChannels * bitsPerSample / 8)
	binary.Write(buf, binary.LittleEndian, byteRate)

	blockAlign := uint16(numChannels * bitsPerSample / 8)
	binary.Write(buf, binary.LittleEndian, blockAlign)

	binary.Write(buf, binary.LittleEndian, uint16(bitsPerSample))

	// data chunk
	buf.Write([]byte("data"))
	subchunk2Size := uint32(len(pcmData))
	binary.Write(buf, binary.LittleEndian, subchunk2Size)

	// Actual PCM data
	buf.Write(pcmData)

	return buf.Bytes()
}
