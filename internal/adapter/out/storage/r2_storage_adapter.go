package storageadapter

import (
	"context"
	"fmt"
	"time"

	"backend-hotlines3/internal/port/outbound/storage"
	"backend-hotlines3/pkg/s3"
)

// R2StorageAdapter adapts pkg/s3.R2Client to the storage.ObjectStorage port.
type R2StorageAdapter struct {
	client *s3.R2Client
}

func NewR2StorageAdapter(client *s3.R2Client) *R2StorageAdapter {
	return &R2StorageAdapter{client: client}
}

func (a *R2StorageAdapter) PresignUpload(ctx context.Context, fileKey, contentType string, expiration time.Duration) (*storage.PresignResult, error) {
	result, err := a.client.GeneratePresignedURL(ctx, fileKey, contentType, expiration)
	if err != nil {
		return nil, fmt.Errorf("presign upload: %w", err)
	}
	return &storage.PresignResult{
		UploadURL: result.UploadURL,
		FileURL:   result.FileURL,
		FileKey:   result.FileKey,
	}, nil
}

func (a *R2StorageAdapter) PresignDownload(ctx context.Context, fileKey string, expiration time.Duration) (string, error) {
	url, err := a.client.GeneratePresignedGetURL(ctx, fileKey, expiration)
	if err != nil {
		return "", fmt.Errorf("presign download: %w", err)
	}
	return url, nil
}

func (a *R2StorageAdapter) DeleteObject(ctx context.Context, fileKey string) error {
	if err := a.client.DeleteObject(ctx, fileKey); err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}

func (a *R2StorageAdapter) PublicURL(fileKey string) string {
	return a.client.GetPublicURL(fileKey)
}
