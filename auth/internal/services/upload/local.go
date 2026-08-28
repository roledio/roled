package upload

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/models"
	"github.com/roledio/roled/pkg/errors"
)

type localUploadService struct {
	UploadPath  string
	BaseURL     string
	MaxFileSize int64
}

func newLocalUploadService(defaultConfig *configs.DefaultConfig) *localUploadService {
	return &localUploadService{
		UploadPath:  defaultConfig.Upload.Local.UploadPath,
		BaseURL:     defaultConfig.BaseURL + "/uploads",
		MaxFileSize: int64(defaultConfig.Upload.MaxFileSizeMB * 1024 * 1024),
	}
}

func (f *localUploadService) Upload(ctx context.Context, req *models.UploadFileRequest) (*models.UploadFileResponse, error) {
	filename, _, err := validateContent(ctx, req, f.MaxFileSize)
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

	// Store in tmp folder
	relativePath := filepath.Join("tmp", req.Type, filename)
	fullPath := filepath.Join(f.UploadPath, relativePath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, errors.ErrSystemError.WithError(err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, errors.ErrSystemError.WithError(err)
	}
	defer func() {
		err := dst.Close()
		if err != nil {
			log.Errorw("Failed to close destination file", "error", err)
		}
	}()

	_, err = io.Copy(dst, file)
	if err != nil {
		return nil, errors.ErrSystemError.WithError(err)
	}

	// The configured UploadPath will be accessed from /uploads URL path defined in main.go
	// So the file URL should not contains the UploadPath prefix
	fileURL := fmt.Sprintf("%s/%s", f.BaseURL, relativePath)

	return &models.UploadFileResponse{
		URL: fileURL,
	}, nil
}

func (f *localUploadService) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(f.UploadPath, path)
	err := os.Remove(fullPath)
	if err != nil {
		return errors.ErrSystemError.WithError(err)
	}
	return nil
}

func (f *localUploadService) Move(ctx context.Context, from string, to string) error {
	src := filepath.Join(f.UploadPath, from)
	dst := filepath.Join(f.UploadPath, to)

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		log.WithContext(ctx).Errorw("Failed to prepare target file directories", "error", err, "to", dst)
		return errors.ErrSystemError.WithError(err)
	}

	err := os.Rename(src, dst)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to move file", "error", err, "from", src, "to", dst)
		return errors.ErrSystemError.WithError(err)
	}
	return nil
}
