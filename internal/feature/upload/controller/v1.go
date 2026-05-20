package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/feature/auth/policy"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

func (h *Controller) GetPresignedURL(c *gin.Context) {
	if role, _ := c.Get("role"); role != policy.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "FORBIDDEN", Message: "Insufficient permissions"}})
		return
	}

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

	ext := filepath.Ext(req.FileName)
	if ext == "" {
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

func (h *Controller) DeleteFile(c *gin.Context) {
	if role, _ := c.Get("role"); role != policy.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "FORBIDDEN", Message: "Insufficient permissions"}})
		return
	}

	fileKey := c.Param("key")
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
