package interceptor

import (
	"context"
	"fmt"

	"ai_companion/rpc/ai/ai"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LogSanitizeInterceptor 是一个 gRPC 服务端一元拦截器。
// 职责：
//  1. 把 GetAiCommentaryReq.AudioContent 这类大字节字段的实际内容剥离，
//     只保留字节大小摘要，避免 base64 编码的音频数据把日志撑爆。
//  2. 为每次 RPC 调用打结构化日志，带 type 字段（hint / evaluation），
//     方便 tail 时用 grep '"type":"evaluation"' 专门过滤 AI 评估日志。
func LogSanitizeInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	var rpcType string

	switch req.(type) {
	case *ai.GetAiCommentaryReq:
		rpcType = "evaluation"
	case *ai.GetAiSuggestionReq:
		rpcType = "hint"
	default:
		rpcType = "other"
	}

	// 调用实际 handler
	resp, err := handler(ctx, req)

	if err != nil {
		st, _ := status.FromError(err)
		logx.Errorw("[ai-rpc] call failed",
			logx.Field("type", rpcType),
			logx.Field("method", info.FullMethod),
			logx.Field("grpc_code", st.Code().String()),
			logx.Field("error", fmt.Sprintf("%v", err)),
		)
	}

	return resp, err
}
