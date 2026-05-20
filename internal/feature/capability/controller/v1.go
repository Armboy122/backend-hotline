package controller

import (
	"errors"
	"net/http"
	"strconv"

	"backend-hotlines3/internal/feature/capability/dto"
	"backend-hotlines3/internal/feature/capability/entity"

	"github.com/gin-gonic/gin"
)

func (c *Controller) List(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: c.service.ListAvailable()})
}

func (c *Controller) ListForUser(ctx *gin.Context) {
	userID, ok := parseUserID(ctx)
	if !ok {
		return
	}
	out, err := c.service.ListForUser(ctx.Request.Context(), userID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: out})
}

func (c *Controller) ReplaceForUser(ctx *gin.Context) {
	userID, ok := parseUserID(ctx)
	if !ok {
		return
	}
	var req dto.ReplaceCapabilitiesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("VALIDATION_ERROR", err.Error())})
		return
	}
	actorID, _ := ctx.Get("user_id")
	out, err := c.service.ReplaceForUser(ctx.Request.Context(), actorID.(uint), userID, req.Capabilities)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: out})
}

func parseUserID(ctx *gin.Context) (uint, bool) {
	raw := ctx.Param("id")
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || parsed == 0 {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_ID", "invalid user id")})
		return 0, false
	}
	return uint(parsed), true
}

func writeError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, entity.ErrInvalidUserID):
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_ID", "invalid user id")})
	case errors.Is(err, entity.ErrInvalidCapability):
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_CAPABILITY", "invalid capability")})
	case errors.Is(err, entity.ErrTargetUserNotFound):
		ctx.JSON(http.StatusNotFound, dto.StandardResponse{Success: false, Error: errResp("TARGET_USER_NOT_FOUND", "target user not found")})
	case errors.Is(err, entity.ErrInactiveTargetUser):
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INACTIVE_TARGET_USER", "target user is inactive")})
	case errors.Is(err, entity.ErrViewerTargetForbidden):
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("VIEWER_TARGET_FORBIDDEN", "viewer cannot receive capabilities")})
	default:
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{Success: false, Error: errResp("INTERNAL_ERROR", "capability operation failed")})
	}
}

func errResp(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}
