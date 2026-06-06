package main

import (
	"flag"
	"fmt"

	_ "github.com/joho/godotenv/autoload"

	"ai_companion/rpc/ai/ai"

	"ai_companion/rpc/ai/internal/config"
	"ai_companion/rpc/ai/internal/interceptor"
	"ai_companion/rpc/ai/internal/server"
	"ai_companion/rpc/ai/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/ai.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		ai.RegisterAiServer(grpcServer, server.NewAiServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(interceptor.LogSanitizeInterceptor)
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
