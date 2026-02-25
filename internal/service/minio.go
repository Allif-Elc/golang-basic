package service

import (
	"context"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioService struct {
	client     *minio.Client
	bucketName string
}

func NewMinioService(client *minio.Client, bucketName string) *MinioService {
	return &MinioService{client: client, bucketName: bucketName}
}

func (s *MinioService) GeneratePresignedUploadURL(objectName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u, err := s.client.PresignedPutObject(ctx, s.bucketName, objectName, 5*time.Minute)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *MinioService) MultipartUpload(objectName string, reader io.Reader, size int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, size, minio.PutObjectOptions{
		PartSize:     10 * 1024 * 1024,
		NumThreads:   4,
		CacheControl: "public, max-age=31536000, immutable",
	})
	return err
}

func (s *MinioService) UploadFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	objectName := filepath.Join("uploads", time.Now().Format("20060102"), strings.TrimSuffix(file.Filename, ext)+"_"+time.Now().Format("150405")+ext)

	if file.Size > 5*1024*1024 {
		err = s.MultipartUpload(objectName, src, file.Size)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		_, err = s.client.PutObject(ctx, s.bucketName, objectName, src, file.Size, minio.PutObjectOptions{
			ContentType:  getContentType(file.Filename),
			CacheControl: "public, max-age=31536000, immutable",
		})
	}
	return objectName, err
}

func (s *MinioService) GeneratePresignedDownloadURL(objectName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename=\""+filepath.Base(objectName)+"\"")
	u, err := s.client.PresignedGetObject(ctx, s.bucketName, objectName, 24*time.Hour, reqParams)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".json":
		return "application/json"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	default:
		return "application/octet-stream"
	}
}
