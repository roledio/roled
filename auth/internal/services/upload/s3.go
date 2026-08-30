package upload

import (
	"context"
	"fmt"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/internal/configs"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/pkg/errors"
)

type s3UploadService struct {
	S3Client    *s3.Client
	Bucket      string
	BaseURL     string
	MaxFileSize int64
}

func newS3UploadService(defaultConfig *configs.DefaultConfig) *s3UploadService {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			defaultConfig.Upload.S3.AccessKey,
			defaultConfig.Upload.S3.SecretKey,
			"",
		)),
		config.WithRegion(defaultConfig.Upload.S3.Region),
	)
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config: %v", err))
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(defaultConfig.Upload.S3.Endpoint)
		o.UsePathStyle = true // For Backblaze B2 compatibility
	})

	return &s3UploadService{
		S3Client:    client,
		Bucket:      defaultConfig.Upload.S3.Bucket,
		BaseURL:     defaultConfig.Upload.S3.BaseURL,
		MaxFileSize: int64(defaultConfig.Upload.MaxFileSizeMB * 1024 * 1024),
	}
}

func (s *s3UploadService) Upload(ctx context.Context, req *models.UploadFileRequest) (*models.UploadFileResponse, error) {
	filename, contentType, err := validateContent(ctx, req, s.MaxFileSize)
	if err != nil {
		return nil, err
	}

	// Re-open file for upload body
	file, err := req.File.Open()
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to re-open uploaded file for upload", "error", err)
		return nil, errors.ErrSystemError.WithError(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Errorw("Failed to close uploaded file", "error", err)
		}
	}()

	// Construct S3 key
	key := path.Join("tmp", req.Type, filename)

	// Upload to S3
	_, err = s.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to upload file to S3", "error", err, "path", key)
		return nil, errors.ErrSystemError.WithError(err)
	}

	// Construct URL
	fileURL := fmt.Sprintf("%s/%s", s.BaseURL, key)

	return &models.UploadFileResponse{
		URL: fileURL,
	}, nil
}

func (s *s3UploadService) Delete(ctx context.Context, path string) error {
	_, err := s.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to delete file from S3", "error", err, "path", path)
		return errors.ErrSystemError.WithError(err)
	}
	return nil
}

func (s *s3UploadService) Move(ctx context.Context, from string, to string) error {
	// Copy object
	_, err := s.S3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.Bucket),
		CopySource: aws.String(s.Bucket + "/" + from),
		Key:        aws.String(to),
	})
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to copy file in S3", "error", err, "from", from, "to", to)
		return errors.ErrSystemError.WithError(err)
	}

	// Delete original
	err = s.Delete(ctx, from)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to delete old file in s3", "error", err, "path", from)
		return errors.ErrSystemError.WithError(err)
	}
	return nil
}
