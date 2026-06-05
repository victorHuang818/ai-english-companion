// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"ai_interview/interview_api/internal/config"
	"ai_interview/interview_api/internal/handler"
	"ai_interview/interview_api/internal/logic"
	"ai_interview/interview_api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/interview-api.yaml", "the config file")

func main() {
	flag.Parse()

	// 尝试手动加载 .env 文件，并打印错误以进行诊断
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("[WARN] godotenv.Load() failed: %v. Trying fallback paths...\n", err)
		_ = godotenv.Load("../.env")
		_ = godotenv.Load("backend/.env")
	} else {
		fmt.Println("[INFO] godotenv.Load() successfully loaded .env file")
	}

	// 打印工作目录
	wd, _ := os.Getwd()
	fmt.Printf("[INFO] Current working directory: %s\n", wd)

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	// 检查并打印 Key 的读取情况（脱敏处理，避免日志泄露密钥）
	fmt.Println("\n======================= CONFIG CHECK =======================")
	if c.RealtimeInterviewer.Qwen.ApiKey == "" {
		fmt.Println("[WARN] DASHSCOPE_API_KEY is EMPTY!")
	} else {
		key := c.RealtimeInterviewer.Qwen.ApiKey
		masked := key
		if len(key) > 8 {
			masked = key[:4] + "..." + key[len(key)-4:]
		}
		fmt.Printf("[INFO] DASHSCOPE_API_KEY successfully loaded: %s (Length: %d)\n", masked, len(key))
	}
	if c.RealtimeInterviewer.Gemini.ApiKey == "" {
		fmt.Println("[WARN] GEMINI_API_KEY is EMPTY!")
	} else {
		key := c.RealtimeInterviewer.Gemini.ApiKey
		masked := key
		if len(key) > 8 {
			masked = key[:4] + "..." + key[len(key)-4:]
		}
		fmt.Printf("[INFO] GEMINI_API_KEY successfully loaded: %s (Length: %d)\n", masked, len(key))
	}
	fmt.Println("============================================================\n")

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	// 全局中间件：从网关透传的 Header 中提取 User-ID 并注入 context
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userId := r.Header.Get("X-User-ID")
			if userId != "" {
				newCtx := context.WithValue(r.Context(), "userId", userId)
				next(w, r.WithContext(newCtx))
			} else {
				next(w, r)
			}
		}
	})

	// 注册生成的 HTTP 路由
	handler.RegisterHandlers(server, ctx)

	// 手动注册 WebSocket 路由
	server.AddRoute(rest.Route{
		Method:  http.MethodGet, // WebSocket 握手是 GET 请求
		Path:    "/ws/interview",
		Handler: handler.InterviewWSHandler(ctx),
	})

	// 启动后台 Redis Stream 任务消费者进行异步落库与评估削峰
	logic.StartConsumer(ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

