package controller

import (
	"errors"
	"net/http"
	"strconv"

	"backend-hotlines3/internal/feature/user/dto"
	"backend-hotlines3/internal/feature/user/entity"
	"backend-hotlines3/internal/feature/user/mapper"
	"backend-hotlines3/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (c *Controller) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	result, err := c.service.List(ctx.Request.Context(), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false, Error: errResp("INTERNAL_ERROR", "Failed to list users"),
		})
		return
	}
	responses := make([]dto.UserResponse, 0, len(result.Items))
	for _, info := range result.Items {
		responses = append(responses, mapper.ToResponse(info))
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    responses,
		Meta:    &dto.Meta{Page: result.Page, Limit: result.Limit, Total: result.Total},
	})
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false, Error: errResp("INVALID_ID", "Invalid user ID"),
		})
		return
	}
	info, err := c.service.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			ctx.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false, Error: errResp("NOT_FOUND", "User not found"),
			})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
				Success: false, Error: errResp("INTERNAL_ERROR", "Failed to get user"),
			})
		}
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToResponse(info)})
}

func (c *Controller) Create(ctx *gin.Context) {
	var req dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleValidationError(ctx, err)
		return
	}
	info, err := c.service.Create(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false, Error: errResp("INTERNAL_ERROR", "Failed to create user"),
		})
		return
	}
	ctx.JSON(http.StatusCreated, dto.StandardResponse{Success: true, Data: mapper.ToResponse(info)})
}

func (c *Controller) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false, Error: errResp("INVALID_ID", "Invalid user ID"),
		})
		return
	}
	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleValidationError(ctx, err)
		return
	}
	info, err := c.service.Update(ctx.Request.Context(), uint(id), req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			ctx.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false, Error: errResp("NOT_FOUND", "User not found"),
			})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
				Success: false, Error: errResp("INTERNAL_ERROR", "Failed to update user"),
			})
		}
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToResponse(info)})
}

func (c *Controller) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false, Error: errResp("INVALID_ID", "Invalid user ID"),
		})
		return
	}
	var callerID uint
	if v, exists := ctx.Get("user_id"); exists {
		switch val := v.(type) {
		case uint:
			callerID = val
		case float64:
			callerID = uint(val)
		}
	}
	if err := c.service.Delete(ctx.Request.Context(), uint(id), callerID); err != nil {
		switch {
		case errors.Is(err, entity.ErrCannotDeleteSelf):
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
				Success: false, Error: errResp("CANNOT_DELETE_SELF", "Cannot delete your own account"),
			})
		case errors.Is(err, entity.ErrNotFound):
			ctx.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false, Error: errResp("NOT_FOUND", "User not found"),
			})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
				Success: false, Error: errResp("INTERNAL_ERROR", "Failed to delete user"),
			})
		}
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) ChangePassword(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false, Error: errResp("INVALID_ID", "Invalid user ID"),
		})
		return
	}
	callerIDRaw, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusForbidden, dto.StandardResponse{
			Success: false, Error: errResp("FORBIDDEN", "Can only change your own password"),
		})
		return
	}
	var callerID uint
	switch val := callerIDRaw.(type) {
	case uint:
		callerID = val
	case float64:
		callerID = uint(val)
	}
	if callerID != uint(id) {
		ctx.JSON(http.StatusForbidden, dto.StandardResponse{
			Success: false, Error: errResp("FORBIDDEN", "Can only change your own password"),
		})
		return
	}
	var req dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleValidationError(ctx, err)
		return
	}
	if err := c.service.ChangePassword(ctx.Request.Context(), uint(id), req); err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			ctx.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false, Error: errResp("NOT_FOUND", "User not found"),
			})
		case errors.Is(err, entity.ErrInvalidPassword):
			ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{
				Success: false, Error: errResp("INVALID_PASSWORD", "Old password is incorrect"),
			})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
				Success: false, Error: errResp("INTERNAL_ERROR", "Failed to change password"),
			})
		}
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    gin.H{"message": "Password changed successfully"},
	})
}

func errResp(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}
