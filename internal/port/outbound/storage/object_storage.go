package storage

import (
	"context"
	"time"
)

// PresignResult holds the presigned upload URL and derived public URL.
type PresignResult struct {
	UploadURL string
	FileURL   string
	FileKey   string
}

// ObjectStorage abstracts R2/S3 operations for the monthly plan module.
type ObjectStorage interface {
	// PresignUpload generates a presigned PUT URL for direct client upload.
	PresignUpload(ctx context.Context, fileKey, contentType string, expiration time.Duration) (*PresignResult, error)

	// PresignDownload generates a presigned GET URL for downloading.
	PresignDownload(ctx context.Context, fileKey string, expiration time.Duration) (string, error)

	// DeleteObject removes an object from storage. Best-effort; errors are logged, not fatal.
	DeleteObject(ctx context.Context, fileKey string) error

	// PublicURL returns the public URL for a given file key.
	PublicURL(fileKey string) string
}
