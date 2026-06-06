package svc

import (
	"context"

	"ai_companion/pkg/oss_util"
	"ai_companion/rpc/ai/ai"

	"ai_companion/rpc/core/internal/config"
	"ai_companion/rpc/core/model"
	
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config               config.Config
	MinioClient          *minio.Client
	OssClient            *oss_util.OSSClient
	UserProfileModel     model.UserProfilesModel
	ScenarioModel        model.ScenariosModel
	PracticeSessionModel model.PracticeSessionsModel
	DialogueModel        model.DialoguesModel
	AiRpc                ai.AiClient
	SqlConn              sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MinIO
	minioClient, err := minio.New(c.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.Minio.AccessKeyID, c.Minio.SecretAccessKey, ""),
		Secure: c.Minio.UseSSL,
	})
	if err != nil {
		logx.Must(err)
	}

	// 自动检查并创建存储桶
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, c.Minio.BucketName)
	if err != nil {
		logx.Errorf("Failed to check if bucket %s exists: %v", c.Minio.BucketName, err)
	} else if !exists {
		err = minioClient.MakeBucket(ctx, c.Minio.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			logx.Errorf("Failed to create bucket %s: %v", c.Minio.BucketName, err)
		} else {
			logx.Infof("Successfully created bucket: %s", c.Minio.BucketName)
		}
	}

	// 使用通用 OSS 工具包初始化 OssClient
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

	// 初始化数据库连接
	sqlConn := sqlx.NewMysql(c.DataSource)

	return &ServiceContext{
		Config:               c,
		MinioClient:          minioClient,
		OssClient:            ossClient,
		UserProfileModel:     model.NewUserProfilesModel(sqlConn, c.CacheRedis),
		ScenarioModel:        model.NewScenariosModel(sqlConn, c.CacheRedis),
		PracticeSessionModel: model.NewPracticeSessionsModel(sqlConn, c.CacheRedis),
		DialogueModel:        model.NewDialoguesModel(sqlConn, c.CacheRedis),
		AiRpc:                ai.NewAiClient(zrpc.MustNewClient(c.AiRpc).Conn()),
		SqlConn:              sqlConn,
	}
}
