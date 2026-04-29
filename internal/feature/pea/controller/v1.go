package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"backend-hotlines3/internal/feature/pea/dto"
	"backend-hotlines3/internal/feature/pea/entity"
	"backend-hotlines3/internal/feature/pea/mapper"

	"github.com/gin-gonic/gin"
)

func (c *Controller) List(ctx *gin.Context) {
	items, err := c.service.List(ctx.Request.Context())
	if err != nil {
		log.Printf("Failed to fetch PEAs: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", "An error occurred while fetching PEAs"),
		})
		return
	}

	response := make([]dto.PEAResponse, 0, len(items))
	for _, item := range items {
		response = append(response, mapper.ToResponse(item))
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: response})
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}

	item, err := c.service.GetByID(ctx.Request.Context(), id)
	if err != nil {
		writeError(ctx, err, "An error occurred while fetching the PEA")
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToResponse(*item)})
}

func (c *Controller) Create(ctx *gin.Context) {
	var req dto.CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   errResp("VALIDATION_ERROR", err.Error()),
		})
		return
	}

	item, err := c.service.Create(ctx.Request.Context(), req.Shortname, req.Fullname, req.OperationID)
	if err != nil {
		log.Printf("Failed to create PEA: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", "An error occurred while creating the PEA"),
		})
		return
	}
	ctx.JSON(http.StatusCreated, dto.StandardResponse{Success: true, Data: mapper.ToResponse(*item)})
}

func (c *Controller) BulkCreate(ctx *gin.Context) {
	var req []dto.CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   errResp("VALIDATION_ERROR", err.Error()),
		})
		return
	}

	if len(req) == 0 {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   errResp("VALIDATION_ERROR", "At least one PEA is required"),
		})
		return
	}

	items := make([]entity.Entity, 0, len(req))
	for _, r := range req {
		items = append(items, entity.Entity{
			Shortname:   r.Shortname,
			Fullname:    r.Fullname,
			OperationID: r.OperationID,
		})
	}

	created, err := c.service.BulkCreate(ctx.Request.Context(), items)
	if err != nil {
		log.Printf("Failed to bulk create PEAs: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", "An error occurred while creating PEAs"),
		})
		return
	}

	response := make([]dto.PEAResponse, 0, len(created))
	for _, item := range created {
		response = append(response, mapper.ToResponse(item))
	}
	ctx.JSON(http.StatusCreated, dto.StandardResponse{Success: true, Data: response})
}

func (c *Controller) Update(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}

	var req dto.UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   errResp("VALIDATION_ERROR", err.Error()),
		})
		return
	}

	item, err := c.service.Update(ctx.Request.Context(), id, req.Shortname, req.Fullname, req.OperationID)
	if err != nil {
		writeError(ctx, err, "An error occurred while updating the PEA")
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToResponse(*item)})
}

func (c *Controller) Delete(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}

	if err := c.service.Delete(ctx.Request.Context(), id); err != nil {
		writeError(ctx, err, "An error occurred while deleting the PEA")
		return
	}
	ctx.Status(http.StatusNoContent)
}

func parseID(ctx *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   errResp("INVALID_ID", "Invalid PEA ID"),
		})
		return 0, false
	}
	return id, true
}

func writeError(ctx *gin.Context, err error, internalMessage string) {
	switch {
	case errors.Is(err, entity.ErrInvalidID):
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   errResp("INVALID_ID", "Invalid PEA ID"),
		})
	case errors.Is(err, entity.ErrNotFound):
		ctx.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error:   errResp("NOT_FOUND", "PEA not found"),
		})
	default:
		log.Printf("PEA request failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", internalMessage),
		})
	}
}

func errResp(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}
