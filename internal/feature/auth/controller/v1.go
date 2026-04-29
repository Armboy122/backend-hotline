package controller

import (
	"errors"
	"log"
	"net/http"

	"backend-hotlines3/internal/feature/auth/dto"
	"backend-hotlines3/internal/feature/auth/entity"
	"backend-hotlines3/internal/feature/auth/mapper"
	"backend-hotlines3/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (c *Controller) Register(ctx *gin.Context) {
	var req dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleValidationError(ctx, err)
		return
	}

	info, err := c.service.Register(ctx.Request.Context(), entity.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
		TeamID:   req.TeamID,
		IsActive: req.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrUserExists):
			ctx.JSON(http.StatusConflict, dto.StandardResponse{
				Success: false, Error: errResp("USER_EXISTS", "Username already taken"),
			})
		case errors.Is(err, entity.ErrTeamRequired):
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
				Success: false, Error: errResp("TEAM_ID_REQUIRED", "TeamID is required for role 'user'"),
			})
		default:
			log.Printf("Register error: %v", err)
			ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
				Success: false, Error: errResp("INTERNAL_ERROR", "Failed to create user"),
			})
		}
		return
	}
	ctx.JSON(http.StatusCreated, dto.StandardResponse{Success: true, Data: mapper.InfoToResponse(info)})
}

func (c *Controller) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleValidationError(ctx, err)
		return
	}

	result, err := c.service.Login(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidCredentials):
			ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{
				Success: false, Error: errResp("INVALID_CREDENTIALS", "Invalid username or password"),
			})
		case errors.Is(err, entity.ErrAccountDisabled):
			ctx.JSON(http.StatusForbidden, dto.StandardResponse{
				Success: false, Error: errResp("ACCOUNT_DISABLED", "Account is disabled"),
			})
		default:
			log.Printf("Login error: %v", err)
			ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
				Success: false, Error: errResp("INTERNAL_ERROR", "Failed to generate tokens"),
			})
		}
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.LoginResultToResponse(result)})
}

func (c *Controller) Logout(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    gin.H{"message": "Logged out successfully"},
	})
}

func (c *Controller) RefreshToken(ctx *gin.Context) {
	var req dto.RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleValidationError(ctx, err)
		return
	}

	pair, err := c.service.RefreshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrInvalidToken):
			ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{
				Success: false, Error: errResp("INVALID_TOKEN", "Invalid or expired refresh token"),
			})
		case errors.Is(err, entity.ErrUserNotFound):
			ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{
				Success: false, Error: errResp("USER_NOT_FOUND", "User not found"),
			})
		case errors.Is(err, entity.ErrAccountDisabled):
			ctx.JSON(http.StatusForbidden, dto.StandardResponse{
				Success: false, Error: errResp("ACCOUNT_DISABLED", "Account is disabled"),
			})
		default:
			log.Printf("RefreshToken error: %v", err)
			ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
				Success: false, Error: errResp("TOKEN_GENERATION_ERROR", "Failed to generate tokens"),
			})
		}
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.TokenPairToRefreshResponse(pair)})
}

func (c *Controller) Me(ctx *gin.Context) {
	userIDRaw, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false, Error: errResp("UNAUTHORIZED", "User not authenticated"),
		})
		return
	}

	var userID uint
	switch v := userIDRaw.(type) {
	case uint:
		userID = v
	case float64:
		userID = uint(v)
	default:
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false, Error: errResp("UNAUTHORIZED", "User not authenticated"),
		})
		return
	}

	info, err := c.service.Me(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false, Error: errResp("USER_NOT_FOUND", "User not found"),
		})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.InfoToResponse(info)})
}

func errResp(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}
