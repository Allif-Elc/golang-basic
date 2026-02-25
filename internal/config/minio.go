package config

import (
	"context"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
}

var MinioClient *minio.Client

func InitMinio() error {
	cfg := MinioConfig{
		Endpoint:   getEnv("MINIO_ENDPOINT", "minio:9000"),
		AccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
		BucketName: getEnv("MINIO_BUCKET_NAME", "golang-basic"),
		UseSSL:     getEnvBool("MINIO_USE_SSL", false),
	}

	var err error
	MinioClient, err = minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := MinioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return err
	}
	if !exists {
		err = MinioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{Region: "us-east-1"})
		if err != nil {
			return err
		}
	}

	log.Printf("MinIO connected: %s (bucket: %s, workers: %d)", cfg.Endpoint, cfg.BucketName, runtime.NumCPU()*4)
	return nil
}

func getEnvBool(key string, defaultValue bool) bool {
	return os.Getenv(key) == "true"
}
