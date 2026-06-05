package handler

import (
	"encoding/json"
	"testing"
)

// 1. 测试 JSON 反序列化的内存消耗（对比有无 sync.Pool）

// 优化前：每次反序列化都在堆上分配新的 RealtimeResponse 结构体
func BenchmarkUnmarshalWithoutPool(b *testing.B) {
	data := []byte(`{
		"serverContent": {
			"turnComplete": true,
			"inputTranscription": {"text": "hello"},
			"outputTranscription": {"text": "hi"}
		},
		"usageMetadata": {
			"totalTokenCount": 150
		}
	}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var gResp RealtimeResponse
		_ = json.Unmarshal(data, &gResp)
	}
}

// 优化后：通过 sync.Pool 复用 RealtimeResponse 结构体
func BenchmarkUnmarshalWithPool(b *testing.B) {
	data := []byte(`{
		"serverContent": {
			"turnComplete": true,
			"inputTranscription": {"text": "hello"},
			"outputTranscription": {"text": "hi"}
		},
		"usageMetadata": {
			"totalTokenCount": 150
		}
	}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gResp := realtimeResponsePool.Get().(*RealtimeResponse)
		*gResp = RealtimeResponse{} // 重置字段
		_ = json.Unmarshal(data, gResp)
		realtimeResponsePool.Put(gResp)
	}
}

// 2. 测试音频切片动态扩容的内存消耗（对比有无 sync.Pool + 预分配）

// 优化前：从空切片开始 append，触发多次扩容和堆内存分配
func BenchmarkAudioWithoutPoolAndPrealloc(b *testing.B) {
	chunk := make([]byte, 320) // 320 字节假音频片
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf []byte
		// 模拟持续接收并拼接 50 包音频
		for j := 0; j < 50; j++ {
			buf = append(buf, chunk...)
		}
	}
}

// 优化后：从对象池中借用 512KB 预分配切片进行 append，零扩容，零堆分配
func BenchmarkAudioWithPoolAndPrealloc(b *testing.B) {
	chunk := make([]byte, 320)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bufRaw := audioBufferPool.Get().([]byte)
		buf := bufRaw[:0] // 重置长度，保留容量
		for j := 0; j < 50; j++ {
			buf = append(buf, chunk...)
		}
		// 归还到池中
		audioBufferPool.Put(buf)
	}
}
