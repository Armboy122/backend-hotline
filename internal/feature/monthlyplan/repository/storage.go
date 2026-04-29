package repository

import (
	"context"
	"fmt"
	"time"

	"backend-hotlines3/pkg/s3"
)

type r2Storage struct {
	client *s3.R2Client
}

func (s *r2Storage) PresignUpload(ctx context.Context, fileKey, contentType string, expiration time.Duration) (*PresignResult, error) {
	result, err := s.client.GeneratePresignedURL(ctx, fileKey, contentType, expiration)
	if err != nil {
		return nil, fmt.Errorf("presign upload: %w", err)
	}
	return &PresignResult{UploadURL: result.UploadURL, FileURL: result.FileURL, FileKey: result.FileKey}, nil
}

func (s *r2Storage) PresignDownload(ctx context.Context, fileKey string, expiration time.Duration) (string, error) {
	url, err := s.client.GeneratePresignedGetURL(ctx, fileKey, expiration)
	if err != nil {
		return "", fmt.Errorf("presign download: %w", err)
	}
	return url, nil
}

func (s *r2Storage) DeleteObject(ctx context.Context, fileKey string) error {
	if err := s.client.DeleteObject(ctx, fileKey); err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}

func (s *r2Storage) PublicURL(fileKey string) string {
	return s.client.GetPublicURL(fileKey)
}
