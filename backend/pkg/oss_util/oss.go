package oss_util

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config holds the OSS configuration credentials.
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

// OSSClient wraps the underlying MinIO client to provide common OSS functionalities.
type OSSClient struct {
	client     *minio.Client
	bucketName string
}

// NewOSSClient initializes a new OSSClient with the provided config.
func NewOSSClient(cfg Config) (*OSSClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &OSSClient{
		client:     client,
		bucketName: cfg.BucketName,
	}, nil
}

// GeneratePresignedDownloadURL generates a pre-signed GET URL for downloading files.
func (c *OSSClient) GeneratePresignedDownloadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedGetObject(ctx, c.bucketName, objectKey, expiry, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// GeneratePresignedUploadURL generates a pre-signed PUT URL for uploading files.
func (c *OSSClient) GeneratePresignedUploadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedPutObject(ctx, c.bucketName, objectKey, expiry)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// FileExists checks whether the specified object key exists in the bucket.
func (c *OSSClient) FileExists(ctx context.Context, objectKey string) (bool, error) {
	_, err := c.client.StatObject(ctx, c.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetMinioClient returns the underlying minio.Client for any advanced usages.
func (c *OSSClient) GetMinioClient() *minio.Client {
	return c.client
}

// GetBucketName returns the configured bucket name.
func (c *OSSClient) GetBucketName() string {
	return c.bucketName
}

// DownloadFile downloads the object content from the bucket as a byte slice.
func (c *OSSClient) DownloadFile(ctx context.Context, objectKey string) ([]byte, error) {
	obj, err := c.client.GetObject(ctx, c.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
