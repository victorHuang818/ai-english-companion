// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"ai_interview/interview_api/internal/config"
	"ai_interview/pkg/oss_util"
	"ai_interview/rpc/ai/ai"

	"ai_interview/rpc/core/core"
	"ai_interview/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	CoreRpc     core.CoreClient
	UserRpc     user.UserClient
	AiRpc       ai.AiClient
	OssClient   *oss_util.OSSClient
	RedisClient *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	ossClient, err := oss_util.NewOSSClient(oss_util.Config{
		Endpoint:        c.Minio.Endpoint,
		AccessKeyID:     c.Minio.AccessKeyID,
		SecretAccessKey: c.Minio.SecretAccessKey,
		UseSSL:          c.Minio.UseSSL,
		BucketName:      c.Minio.BucketName,
	})
	if err != nil {
		logx.Must(err)
	}

	redisClient := redis.MustNewRedis(c.Redis)

	return &ServiceContext{
		Config:      c,
		CoreRpc:     core.NewCoreClient(zrpc.MustNewClient(c.CoreRpc).Conn()),
		UserRpc:     user.NewUserClient(zrpc.MustNewClient(c.UserRpc).Conn()),
		AiRpc:       ai.NewAiClient(zrpc.MustNewClient(c.AiRpc).Conn()),
		OssClient:   ossClient,
		RedisClient: redisClient,
	}
}

