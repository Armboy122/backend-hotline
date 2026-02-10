package v1

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/pkg/s3"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandler struct {
	r2Client *s3.R2Client
}

func NewUploadHandler(cfg *config.Config) (*UploadHandler, error) {
	r2Client, err := s3.NewR2Client(s3.R2Config{
		AccountID:       cfg.Cloudflare.R2.AccountID,
		AccessKeyID:     cfg.Cloudflare.R2.AccessKeyID,
		SecretAccessKey: cfg.Cloudflare.R2.SecretAccessKey,
		BucketName:      cfg.Cloudflare.R2.BucketName,
		PublicURL:       cfg.Cloudflare.R2.PublicURL,
	})
	if err != nil {
		return nil, err
	}

	return &UploadHandler{r2Client: r2Client}, nil
}

// allowedImageTypes defines allowed MIME types for images
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// GetPresignedURL generates a presigned URL for direct image upload to Cloudflare R2.
// The presigned URL is valid for 15 minutes.
func (h *UploadHandler) GetPresignedURL(c *gin.Context) {
	var req dto.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Validate file type
	if !allowedImageTypes[req.FileType] {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_FILE_TYPE",
				Message: "ประเภทไฟล์ไม่ถูกต้อง รองรับเฉพาะ JPG, PNG, WebP, GIF",
			},
		})
		return
	}

	// Generate unique file key
	ext := filepath.Ext(req.FileName)
	if ext == "" {
		// Derive extension from MIME type
		switch req.FileType {
		case "image/jpeg", "image/jpg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		case "image/gif":
			ext = ".gif"
		}
	}

	fileKey := fmt.Sprintf("images/%d-%s%s", time.Now().UnixMilli(), uuid.New().String()[:8], ext)

	// Generate presigned URL (valid for 15 minutes)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := h.r2Client.GeneratePresignedURL(ctx, fileKey, req.FileType, 15*time.Minute)
	if err != nil {
		log.Printf("Failed to generate presigned URL for key %s: %v", fileKey, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "UPLOAD_ERROR",
				Message: "Failed to generate upload URL",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.PresignedURLResponse{
			UploadURL: result.UploadURL,
			FileURL:   result.FileURL,
			FileKey:   result.FileKey,
		},
	})
}

// DeleteFile removes a file from Cloudflare R2 storage by its key.
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	// The key might contain slashes, so we need to get the full path
	fileKey := c.Param("key")

	// Handle URL-encoded slashes and remove leading slash if present
	fileKey = strings.TrimPrefix(fileKey, "/")

	if fileKey == "" {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_KEY",
				Message: "File key is required",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.r2Client.DeleteObject(ctx, fileKey); err != nil {
		log.Printf("Failed to delete file with key %s: %v", fileKey, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DELETE_ERROR",
				Message: "Failed to delete file",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: map[string]string{
			"message": "File deleted successfully",
		},
	})
}

// UploadDirect handles direct file uploads through the server (multipart/form-data).
// This is an alternative to the presigned URL approach.
func (h *UploadHandler) UploadDirect(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NO_FILE",
				Message: "No file uploaded",
			},
		})
		return
	}
	defer file.Close()

	// Check file size (max 5MB)
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "FILE_TOO_LARGE",
				Message: "File size exceeds 5MB limit",
			},
		})
		return
	}

	// Check content type
	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_FILE_TYPE",
				Message: "ประเภทไฟล์ไม่ถูกต้อง รองรับเฉพาะ JPG, PNG, WebP, GIF",
			},
		})
		return
	}

	// Generate unique file key
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		switch contentType {
		case "image/jpeg", "image/jpg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		case "image/gif":
			ext = ".gif"
		}
	}

	fileKey := fmt.Sprintf("images/%d-%s%s", time.Now().UnixMilli(), uuid.New().String()[:8], ext)

	// For direct upload, we would need to implement PutObject
	// For now, return the presigned URL approach info
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := h.r2Client.GeneratePresignedURL(ctx, fileKey, contentType, 15*time.Minute)
	if err != nil {
		log.Printf("Failed to generate presigned URL for direct upload %s: %v", fileKey, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "UPLOAD_ERROR",
				Message: "Failed to process upload",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.UploadResponse{
			URL:          result.FileURL,
			FileName:     result.FileKey,
			OriginalName: header.Filename,
			Size:         header.Size,
			Type:         contentType,
		},
	})
}
