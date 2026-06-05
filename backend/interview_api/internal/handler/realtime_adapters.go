package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// RealtimeInterviewerConn 统一音视频面试官连接的抽象适配器接口
type RealtimeInterviewerConn interface {
	Initialize(systemInstruction string) error
	SendAudioChunk(data []byte) error
	ReadMessage() (messageType int, payload []byte, turnComplete bool, inputTrans string, outputTrans string, totalTokens int64, err error)
	Close() error
}

// ==========================================
// 1. Gemini Live API 适配器实现
// ==========================================
type GeminiRealtimeConn struct {
	ws *websocket.Conn
}

func NewGeminiRealtimeConn(apiKey string, model string) (RealtimeInterviewerConn, error) {
	geminiUrl := fmt.Sprintf("wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1beta.GenerativeService.BidiGenerateContent?key=%s", apiKey)
	ws, _, err := websocket.DefaultDialer.Dial(geminiUrl, nil)
	if err != nil {
		return nil, err
	}
	return &GeminiRealtimeConn{ws: ws}, nil
}

func (c *GeminiRealtimeConn) Initialize(systemInstruction string) error {
	setup := map[string]interface{}{
		"config": map[string]interface{}{
			"systemInstruction": map[string]interface{}{
				"parts": []map[string]interface{}{
					{"text": systemInstruction},
				},
			},
			"responseModalities": []string{"AUDIO", "TEXT"},
		},
	}
	return c.ws.WriteJSON(setup)
}

func (c *GeminiRealtimeConn) SendAudioChunk(data []byte) error {
	geminiInput := map[string]interface{}{
		"realtimeInput": map[string]interface{}{
			"audio": map[string]interface{}{
				"mimeType": "audio/pcm;rate=16000",
				"data":     base64.StdEncoding.EncodeToString(data),
			},
		},
	}
	return c.ws.WriteJSON(geminiInput)
}

func (c *GeminiRealtimeConn) ReadMessage() (messageType int, payload []byte, turnComplete bool, inputTrans string, outputTrans string, totalTokens int64, err error) {
	mt, data, err := c.ws.ReadMessage()
	if err != nil {
		return mt, nil, false, "", "", 0, err
	}

	var gResp RealtimeResponse
	if err := json.Unmarshal(data, &gResp); err == nil && gResp.ServerContent != nil {
		if gResp.ServerContent.InputTranscription != nil {
			inputTrans = gResp.ServerContent.InputTranscription.Text
		}
		if gResp.ServerContent.OutputTranscription != nil {
			outputTrans = gResp.ServerContent.OutputTranscription.Text
		}
		turnComplete = gResp.ServerContent.TurnComplete
		if gResp.UsageMetadata != nil {
			totalTokens = gResp.UsageMetadata.TotalTokenCount
		}
	}

	return mt, data, turnComplete, inputTrans, outputTrans, totalTokens, nil
}

func (c *GeminiRealtimeConn) Close() error {
	return c.ws.Close()
}

// ==========================================
// 2. Qwen-Omni-Realtime (阿里云百炼) 适配器实现
// ==========================================
type QwenRealtimeConn struct {
	ws *websocket.Conn
}

func NewQwenRealtimeConn(apiKey string, model string) (RealtimeInterviewerConn, error) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+apiKey)
	
	// 中国北京接入地址，支持指定模型参数
	qwenUrl := fmt.Sprintf("wss://dashscope.aliyuncs.com/api-ws/v1/realtime?model=%s", model)
	ws, resp, err := websocket.DefaultDialer.Dial(qwenUrl, headers)
	if err != nil {
		if resp != nil {
			defer resp.Body.Close()
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("websocket dial failed: status=%s, body=%s, err=%w", resp.Status, string(bodyBytes), err)
		}
		return nil, err
	}
	return &QwenRealtimeConn{ws: ws}, nil
}

func (c *QwenRealtimeConn) Initialize(systemInstruction string) error {
	setup := map[string]interface{}{
		"type": "session.update",
		"session": map[string]interface{}{
			"modalities": []string{"text", "audio"},
			"instructions": systemInstruction,
			"voice": "Ethan", // 使用官方标准音色 Ethan
			"input_audio_format": "pcm", // 仅支持 pcm
			"output_audio_format": "pcm", // 仅支持 pcm
			"turn_detection": map[string]interface{}{
				"type": "semantic_vad", // 启用语义 VAD 模式，使面试对话判定更自然，避免断句被打断
			},
		},
	}
	return c.ws.WriteJSON(setup)
}

func (c *QwenRealtimeConn) SendAudioChunk(data []byte) error {
	event := map[string]interface{}{
		"type": "input_audio_buffer.append",
		"audio": base64.StdEncoding.EncodeToString(data),
	}
	return c.ws.WriteJSON(event)
}

func (c *QwenRealtimeConn) ReadMessage() (messageType int, payload []byte, turnComplete bool, inputTrans string, outputTrans string, totalTokens int64, err error) {
	mt, data, err := c.ws.ReadMessage()
	if err != nil {
		return mt, nil, false, "", "", 0, err
	}

	var qwenEvent struct {
		Type       string `json:"type"`
		Delta      string `json:"delta"`
		Audio      string `json:"audio"`
		Text       string `json:"text"`
		Transcript string `json:"transcript"`
		Response   *struct {
			Usage *struct {
				TotalTokens int64 `json:"total_tokens"`
			} `json:"usage"`
		} `json:"response"`
	}
	if err := json.Unmarshal(data, &qwenEvent); err != nil {
		return mt, data, false, "", "", 0, nil
	}

	// 🌟 调试日志：打印接收到的所有 Qwen WebSocket 事件类型和原始数据
	if qwenEvent.Type != "" {
		fmt.Printf("[Qwen Event Log] Type: %s, RawData: %s\n", qwenEvent.Type, string(data))
	}

	// 捕获并输出 Qwen 错误事件
	if qwenEvent.Type == "error" {
		fmt.Printf("[Qwen Error Log] Received error event: %s\n", string(data))
	}

	// 将 Qwen-Omni 的协议规范动态桥接并转化为前端和 BFF 所兼容的 Gemini 数据格式，实现完美的透明兼容
	translated := make(map[string]interface{})
	serverContent := make(map[string]interface{})

	switch qwenEvent.Type {
	case "conversation.item.input_audio_transcription.completed":
		// 用户说话完毕，输出完整转录文本，生成单条 clean 消息 bubble
		inputTrans = qwenEvent.Transcript
		serverContent["inputTranscription"] = map[string]string{
			"text": qwenEvent.Transcript,
		}
		translated["serverContent"] = serverContent

	case "response.audio_transcript.done":
		// AI 回答完毕，输出完整转录文本，生成单条 clean 消息 bubble
		outputTrans = qwenEvent.Transcript
		serverContent["outputTranscription"] = map[string]string{
			"text": qwenEvent.Transcript,
		}
		translated["serverContent"] = serverContent

	case "response.text.delta":
		outputTrans = qwenEvent.Delta
		serverContent["outputTranscription"] = map[string]string{
			"text": qwenEvent.Delta,
		}
		translated["serverContent"] = serverContent

	case "response.audio.delta":
		parts := []map[string]interface{}{
			{
				"inlineData": map[string]string{
					"mimeType": "audio/pcm;rate=24000", // Qwen 官方输出 PCM 采样率为 24kHz
					"data":     qwenEvent.Delta,
				},
			},
		}
		serverContent["modelTurn"] = map[string]interface{}{
			"parts": parts,
		}
		translated["serverContent"] = serverContent

	case "response.done":
		turnComplete = true
		serverContent["turnComplete"] = true
		translated["serverContent"] = serverContent

		if qwenEvent.Response != nil && qwenEvent.Response.Usage != nil {
			totalTokens = qwenEvent.Response.Usage.TotalTokens
			translated["usageMetadata"] = map[string]interface{}{
				"totalTokenCount": totalTokens,
			}
		}
	}

	if len(translated) > 0 {
		translatedBytes, _ := json.Marshal(translated)
		return mt, translatedBytes, turnComplete, inputTrans, outputTrans, totalTokens, nil
	}

	return mt, data, false, "", "", 0, nil
}

func (c *QwenRealtimeConn) Close() error {
	return c.ws.Close()
}

// ==========================================
// 3. Mock Realtime Conn 适配器实现 (专供高并发/吞吐量压力测试使用)
// ==========================================
type mockMsg struct {
	payload      []byte
	turnComplete bool
	inputTrans   string
	outputTrans  string
	totalTokens  int64
}

type MockRealtimeConn struct {
	msgChan    chan mockMsg
	audioCount int
	mu         sync.Mutex
	closed     bool
}

func NewMockRealtimeConn() (RealtimeInterviewerConn, error) {
	return &MockRealtimeConn{
		msgChan: make(chan mockMsg, 200),
	}, nil
}

func (c *MockRealtimeConn) Initialize(systemInstruction string) error {
	// 模拟 AI 面试官初始问候
	c.msgChan <- mockMsg{
		payload:     []byte(`{"serverContent": {"outputTranscription": {"text": "你好！欢迎参加面试。请先简要自我介绍一下。"}}}`),
		outputTrans: "你好！欢迎参加面试。请先简要自我介绍一下。",
	}
	// 模拟 AI 发送一小段音频字节
	c.msgChan <- mockMsg{
		payload: []byte(`{"serverContent": {"modelTurn": {"parts": [{"inlineData": {"mimeType": "audio/pcm;rate=24000", "data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}]}}}`),
	}
	// 模拟本回合结束
	c.msgChan <- mockMsg{
		payload:      []byte(`{"serverContent": {"turnComplete": true}, "usageMetadata": {"totalTokenCount": 100}}`),
		turnComplete: true,
		totalTokens:  100,
	}
	return nil
}

func (c *MockRealtimeConn) SendAudioChunk(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return io.ErrClosedPipe
	}

	c.audioCount++
	// 每收到大约 50 个音频小包（约 2 秒音频），模拟用户停止说话并触发 AI 面试官进行流式响应
	if c.audioCount >= 50 {
		c.audioCount = 0
		
		// 1. 模拟用户回答的文本识别结束
		c.msgChan <- mockMsg{
			payload:    []byte(`{"serverContent": {"inputTranscription": {"text": "好的，我是候选人，我来面试这个岗位。"}}}`),
			inputTrans: "好的，我是候选人，我来面试这个岗位。",
		}
		// 2. 模拟 AI 面试官的新提问文本生成完毕
		c.msgChan <- mockMsg{
			payload:     []byte(`{"serverContent": {"outputTranscription": {"text": "非常好。请问你在上一家公司最核心的项目是什么，你担任了什么角色？"}}}`),
			outputTrans: "非常好。请问你在上一家公司最核心的项目是什么，你担任了什么角色？",
		}
		// 3. 模拟 AI 面试官生成的新提问音频流发包
		c.msgChan <- mockMsg{
			payload: []byte(`{"serverContent": {"modelTurn": {"parts": [{"inlineData": {"mimeType": "audio/pcm;rate=24000", "data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}]}}}`),
		}
		// 4. 模拟本回合结束信号并附加计费 Token
		c.msgChan <- mockMsg{
			payload:      []byte(`{"serverContent": {"turnComplete": true}, "usageMetadata": {"totalTokenCount": 150}}`),
			turnComplete: true,
			totalTokens:  150,
		}
	}

	return nil
}

func (c *MockRealtimeConn) ReadMessage() (messageType int, payload []byte, turnComplete bool, inputTrans string, outputTrans string, totalTokens int64, err error) {
	msg, ok := <-c.msgChan
	if !ok {
		return websocket.TextMessage, nil, false, "", "", 0, io.EOF
	}
	return websocket.TextMessage, msg.payload, msg.turnComplete, msg.inputTrans, msg.outputTrans, msg.totalTokens, nil
}

func (c *MockRealtimeConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		close(c.msgChan)
	}
	return nil
}
