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

func (c *Controller) ListContacts(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	var teamID *int64
	if raw := ctx.Query("teamId"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid teamId")})
			return
		}
		teamID = &parsed
	}
	includeInactive := ctx.Query("includeInactive") == "true"
	result, err := c.service.ListContacts(ctx.Request.Context(), actorFromContext(ctx), entity.ContactListQuery{
		Query:           ctx.Query("query"),
		TeamID:          teamID,
		Role:            ctx.Query("role"),
		IncludeInactive: includeInactive,
		Page:            page,
		Limit:           limit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{Success: false, Error: errResp("ERROR", "Failed to list contacts")})
		return
	}
	responses := make([]dto.ContactDirectoryResponse, 0, len(result.Items))
	actor := actorFromContext(ctx)
	for _, info := range result.Items {
		responses = append(responses, mapper.ToContactResponse(info, actor))
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: responses, Meta: &dto.Meta{Page: result.Page, Limit: result.Limit, Total: result.Total}})
}

func (c *Controller) GetContactByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("userId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid user ID")})
		return
	}
	info, err := c.service.GetContactByID(ctx.Request.Context(), uint(id))
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			ctx.JSON(http.StatusNotFound, dto.StandardResponse{Success: false, Error: errResp("NOT_FOUND", "User not found")})
		case errors.Is(err, entity.ErrInvalidID):
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid user ID")})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{Success: false, Error: errResp("ERROR", "Failed to get contact")})
		}
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToContactResponse(info, actorFromContext(ctx))})
}

func (c *Controller) UpdateOwnContact(ctx *gin.Context) {
	actor := actorFromContext(ctx)
	if actor.ID == 0 {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Missing user context")})
		return
	}
	c.updateContact(ctx, actor.ID)
}

func (c *Controller) UpdateContact(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid user ID")})
		return
	}
	c.updateContact(ctx, uint(id))
}

func (c *Controller) updateContact(ctx *gin.Context, id uint) {
	var req dto.UpdateContactRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleValidationError(ctx, err)
		return
	}
	info, err := c.service.UpdateContact(ctx.Request.Context(), actorFromContext(ctx), id, req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			ctx.JSON(http.StatusForbidden, dto.StandardResponse{Success: false, Error: errResp("FORBIDDEN", "Insufficient permissions")})
		case errors.Is(err, entity.ErrNotFound):
			ctx.JSON(http.StatusNotFound, dto.StandardResponse{Success: false, Error: errResp("NOT_FOUND", "User not found")})
		case errors.Is(err, entity.ErrInvalidID), errors.Is(err, entity.ErrNoContactFields), errors.Is(err, entity.ErrInvalidContactField):
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", err.Error())})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{Success: false, Error: errResp("ERROR", "Failed to update contact")})
		}
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToContactResponse(info, actorFromContext(ctx))})
}

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
	info, err := c.service.Create(ctx.Request.Context(), actorFromContext(ctx), req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			ctx.JSON(http.StatusForbidden, dto.StandardResponse{Success: false, Error: errResp("FORBIDDEN", "Insufficient permissions")})
		case errors.Is(err, entity.ErrSuperAdminAlreadyExists):
			ctx.JSON(http.StatusConflict, dto.StandardResponse{Success: false, Error: errResp("SUPER_ADMIN_EXISTS", "Active super admin already exists")})
		case errors.Is(err, entity.ErrInvalidRole):
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_ROLE", "Invalid role")})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
				Success: false, Error: errResp("INTERNAL_ERROR", "Failed to create user"),
			})
		}
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
	info, err := c.service.Update(ctx.Request.Context(), actorFromContext(ctx), uint(id), req)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			ctx.JSON(http.StatusForbidden, dto.StandardResponse{
				Success: false, Error: errResp("FORBIDDEN", "Insufficient permissions"),
			})
		case errors.Is(err, entity.ErrSuperAdminAlreadyExists):
			ctx.JSON(http.StatusConflict, dto.StandardResponse{
				Success: false, Error: errResp("SUPER_ADMIN_EXISTS", "Active super admin already exists"),
			})
		case errors.Is(err, entity.ErrCannotRemoveOnlySuperAdmin):
			ctx.JSON(http.StatusConflict, dto.StandardResponse{
				Success: false, Error: errResp("ONLY_SUPER_ADMIN", "Cannot remove only active super admin"),
			})
		case errors.Is(err, entity.ErrInvalidRole):
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
				Success: false, Error: errResp("INVALID_ROLE", "Invalid role"),
			})
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
	if err := c.service.Delete(ctx.Request.Context(), actorFromContext(ctx), uint(id)); err != nil {
		switch {
		case errors.Is(err, entity.ErrCannotDeleteSelf):
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
				Success: false, Error: errResp("CANNOT_DELETE_SELF", "Cannot delete your own account"),
			})
		case errors.Is(err, entity.ErrForbidden):
			ctx.JSON(http.StatusForbidden, dto.StandardResponse{
				Success: false, Error: errResp("FORBIDDEN", "Insufficient permissions"),
			})
		case errors.Is(err, entity.ErrCannotRemoveOnlySuperAdmin):
			ctx.JSON(http.StatusConflict, dto.StandardResponse{
				Success: false, Error: errResp("ONLY_SUPER_ADMIN", "Cannot remove only active super admin"),
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

func (c *Controller) ResetPassword(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_ID", "Invalid user ID")})
		return
	}
	var req dto.ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.HandleValidationError(ctx, err)
		return
	}
	if err := c.service.ResetPassword(ctx.Request.Context(), actorFromContext(ctx), uint(id), req); err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			ctx.JSON(http.StatusForbidden, dto.StandardResponse{Success: false, Error: errResp("FORBIDDEN", "Only super admin can reset passwords")})
		case errors.Is(err, entity.ErrNotFound):
			ctx.JSON(http.StatusNotFound, dto.StandardResponse{Success: false, Error: errResp("NOT_FOUND", "User not found")})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{Success: false, Error: errResp("INTERNAL_ERROR", "Failed to reset password")})
		}
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: gin.H{"message": "Password reset successfully"}})
}

func actorFromContext(ctx *gin.Context) entity.Actor {
	actor := entity.Actor{}
	if v, exists := ctx.Get("user_id"); exists {
		switch val := v.(type) {
		case uint:
			actor.ID = val
		case float64:
			actor.ID = uint(val)
		case int:
			actor.ID = uint(val)
		case int64:
			actor.ID = uint(val)
		}
	}
	if v, exists := ctx.Get("role"); exists {
		if role, ok := v.(string); ok {
			actor.Role = role
		}
	}
	return actor
}

func errResp(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}
