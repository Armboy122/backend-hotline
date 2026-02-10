// Package s3 provides access to Cloudflare R2 storage for file uploads.
// It generates presigned URLs for direct client uploads and handles file deletion.
package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Client wraps the S3 client for Cloudflare R2
type R2Client struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucketName string
	publicURL  string
}

// R2Config holds the configuration for R2
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
}

// NewR2Client creates a new R2 client
func NewR2Client(cfg R2Config) (*R2Client, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID),
		}, nil
	})

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	presigner := s3.NewPresignClient(client)

	return &R2Client{
		client:     client,
		presigner:  presigner,
		bucketName: cfg.BucketName,
		publicURL:  cfg.PublicURL,
	}, nil
}

// PresignedURLResult contains the presigned URL and file URL
type PresignedURLResult struct {
	UploadURL string
	FileURL   string
	FileKey   string
}

// GeneratePresignedURL generates a presigned URL for uploading
func (r *R2Client) GeneratePresignedURL(ctx context.Context, fileKey string, contentType string, expiration time.Duration) (*PresignedURLResult, error) {
	presignResult, err := r.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(fileKey),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	fileURL := fmt.Sprintf("%s/%s", r.publicURL, fileKey)

	return &PresignedURLResult{
		UploadURL: presignResult.URL,
		FileURL:   fileURL,
		FileKey:   fileKey,
	}, nil
}

// DeleteObject deletes an object from R2
func (r *R2Client) DeleteObject(ctx context.Context, fileKey string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// GetPublicURL returns the public URL for a file key
func (r *R2Client) GetPublicURL(fileKey string) string {
	return fmt.Sprintf("%s/%s", r.publicURL, fileKey)
}
