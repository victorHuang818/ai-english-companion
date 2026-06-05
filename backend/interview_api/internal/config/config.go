// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	CoreRpc zrpc.RpcClientConf
	UserRpc zrpc.RpcClientConf
	AiRpc   zrpc.RpcClientConf
	Mock    bool `json:",optional"`

	Redis redis.RedisConf

	Minio struct {
		Endpoint        string
		AccessKeyID     string
		SecretAccessKey string
		UseSSL          bool
		BucketName      string
	}

	RealtimeInterviewer struct {
		Provider string // "Gemini" or "Qwen"
		Gemini   struct {
			ApiKey string
			Model  string
		}
		Qwen struct {
			ApiKey string
			Model  string
		}
	}
	Auth struct {
		AccessSecret string
	}
}


