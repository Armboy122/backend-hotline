package v1

import (
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/middleware"
	"backend-hotlines3/internal/models"
	"backend-hotlines3/pkg/jwt"
	"backend-hotlines3/pkg/password"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db         *gorm.DB
	jwtManager *jwt.JWTManager
}

func NewAuthHandler(db *gorm.DB, jwtManager *jwt.JWTManager) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtManager: jwtManager,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleValidationError(c, err)
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "USER_EXISTS",
				Message: "Username already taken",
			},
		})
		return
	}

	// Hash password
	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "HASHING_ERROR",
				Message: "Failed to process password",
			},
		})
		return
	}

	// Create user
	user := models.User{
		Username: req.Username,
		Password: hashedPassword,
		Role:     req.Role,
		TeamID:   req.TeamID,
		IsActive: true,
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: "Failed to create user",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data: dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	})
}

// Login - POST /v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleValidationError(c, err)
		return
	}

	var user models.User
	// Login by username only
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_CREDENTIALS",
				Message: "Invalid username or password",
			},
		})
		return
	}

	if !password.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_CREDENTIALS",
				Message: "Invalid username or password",
			},
		})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "ACCOUNT_DISABLED",
				Message: "Account is disabled",
			},
		})
		return
	}

	now := time.Now()
	user.LastLogin = &now
	h.db.Save(&user)

	accessToken, refreshToken, err := h.jwtManager.GenerateTokenPair(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "Failed to generate tokens",
			},
		})
		return
	}

	lastLoginStr := ""
	if user.LastLogin != nil {
		lastLoginStr = user.LastLogin.Format(time.RFC3339)
	}

	response := dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Role:      user.Role,
			TeamID:    user.TeamID,
			IsActive:  user.IsActive,
			LastLogin: &lastLoginStr,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Logout - POST /v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    gin.H{"message": "Logged out successfully"},
	})
}

// RefreshToken - POST /v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleValidationError(c, err)
		return
	}

	claims, err := h.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_TOKEN",
				Message: "Invalid or expired refresh token",
			},
		})
		return
	}

	var user models.User
	if err := h.db.First(&user, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "USER_NOT_FOUND",
				Message: "User not found",
			},
		})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "ACCOUNT_DISABLED",
				Message: "Account is disabled",
			},
		})
		return
	}

	accessToken, refreshToken, err := h.jwtManager.GenerateTokenPair(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "TOKEN_GENERATION_ERROR",
				Message: "Failed to generate tokens",
			},
		})
		return
	}

	response := dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Me - GET /v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "User not authenticated",
			},
		})
		return
	}

	var user models.User
	if err := h.db.Preload("Team").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "USER_NOT_FOUND",
				Message: "User not found",
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
