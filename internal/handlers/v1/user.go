package v1

import (
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// List - GET /v1/users (admin only)
func (h *UserHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Build query
	query := h.db.WithContext(ctx).Scopes(models.UserNotDeleted).Preload("Team")

	// Get total count
	var total int64
	query.Count(&total)

	// Get users with pagination
	var users []models.User
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	var response []dto.UserResponse
	for _, user := range users {
		lastLoginStr := ""
		if user.LastLogin != nil {
			lastLoginStr = user.LastLogin.Format(time.RFC3339)
		}

		response = append(response, dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Role:      user.Role,
			TeamID:    user.TeamID,
			IsActive:  user.IsActive,
			LastLogin: &lastLoginStr,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
		Meta: &dto.Meta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetByID - GET /v1/users/:id (admin only)
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid user ID",
			},
		})
		return
	}

	var user models.User
	if err := h.db.WithContext(c.Request.Context()).Scopes(models.UserNotDeleted).Preload("Team").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "User not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	lastLoginStr := ""
	if user.LastLogin != nil {
		lastLoginStr = user.LastLogin.Format(time.RFC3339)
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TeamID:    user.TeamID,
		IsActive:  user.IsActive,
		LastLogin: &lastLoginStr,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Create - POST /v1/users (admin only)
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "PASSWORD_HASH_ERROR",
				Message: "Failed to hash password",
			},
		})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     req.Role,
		TeamID:   req.TeamID,
		IsActive: isActive,
	}

	if err := h.db.WithContext(c.Request.Context()).Create(&user).Error; err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	if err := h.db.WithContext(c.Request.Context()).Preload("Team").First(&user, user.ID).Error; err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TeamID:    user.TeamID,
		IsActive:  user.IsActive,
		LastLogin: nil,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Update - PUT /v1/users/:id (admin only)
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid user ID",
			},
		})
		return
	}

	var user models.User
	if err := h.db.WithContext(c.Request.Context()).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "User not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	var req dto.UpdateUserRequest
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

	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.TeamID != nil {
		user.TeamID = req.TeamID
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := h.db.WithContext(c.Request.Context()).Save(&user).Error; err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	lastLoginStr := ""
	if user.LastLogin != nil {
		lastLoginStr = user.LastLogin.Format(time.RFC3339)
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TeamID:    user.TeamID,
		IsActive:  user.IsActive,
		LastLogin: &lastLoginStr,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Delete - DELETE /v1/users/:id (admin only)
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid user ID",
			},
		})
		return
	}

	currentUserID, exists := c.Get("user_id")
	if exists && currentUserID.(uint) == uint(id) {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "CANNOT_DELETE_SELF",
				Message: "Cannot delete your own account",
			},
		})
		return
	}

	var user models.User
	if err := h.db.WithContext(c.Request.Context()).Scopes(models.UserNotDeleted).First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "User not found",
			},
		})
		return
	}

	now := time.Now()
	user.DeletedAt = &now
	if err := h.db.WithContext(c.Request.Context()).Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ChangePassword - PUT /v1/users/:id/password (authenticated user can change own password)
func (h *UserHandler) ChangePassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid user ID",
			},
		})
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists || currentUserID.(uint) != uint(id) {
		c.JSON(http.StatusForbidden, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "FORBIDDEN",
				Message: "Can only change your own password",
			},
		})
		return
	}

	var req dto.ChangePasswordRequest
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

	var user models.User
	if err := h.db.WithContext(c.Request.Context()).First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "User not found",
			},
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_PASSWORD",
				Message: "Old password is incorrect",
			},
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "PASSWORD_HASH_ERROR",
				Message: "Failed to hash password",
			},
		})
		return
	}

	user.Password = string(hashedPassword)
	if err := h.db.WithContext(c.Request.Context()).Save(&user).Error; err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    gin.H{"message": "Password changed successfully"},
	})
}
