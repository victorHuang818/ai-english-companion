package config

import "github.com/zeromicro/go-zero/zrpc"

type ProviderConfig struct {
	Vendor  string
	ApiKey  string
	BaseUrl string
	Model   string
}

type Config struct {
	zrpc.RpcServerConf
	AiCommentator ProviderConfig
	AiSuggestion  ProviderConfig
}
