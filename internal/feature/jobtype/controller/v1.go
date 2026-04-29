package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"backend-hotlines3/internal/feature/jobtype/dto"
	"backend-hotlines3/internal/feature/jobtype/entity"
	"backend-hotlines3/internal/feature/jobtype/mapper"

	"github.com/gin-gonic/gin"
)

func (c *Controller) List(ctx *gin.Context) {
	items, err := c.service.List(ctx.Request.Context())
	if err != nil {
		log.Printf("Failed to fetch job types: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", "An error occurred while fetching job types"),
		})
		return
	}

	response := make([]dto.JobTypeResponse, 0, len(items))
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
		writeError(ctx, err, "An error occurred while fetching the job type")
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToResponse(*item)})
}

func (c *Controller) Create(ctx *gin.Context) {
	var req dto.UpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   errResp("VALIDATION_ERROR", err.Error()),
		})
		return
	}

	item, err := c.service.Create(ctx.Request.Context(), req.Name)
	if err != nil {
		log.Printf("Failed to create job type: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", "An error occurred while creating the job type"),
		})
		return
	}
	ctx.JSON(http.StatusCreated, dto.StandardResponse{Success: true, Data: mapper.ToResponse(*item)})
}

func (c *Controller) Update(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}

	var req dto.UpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   errResp("VALIDATION_ERROR", err.Error()),
		})
		return
	}

	item, err := c.service.Update(ctx.Request.Context(), id, req.Name)
	if err != nil {
		writeError(ctx, err, "An error occurred while updating the job type")
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
		writeError(ctx, err, "An error occurred while deleting the job type")
		return
	}
	ctx.Status(http.StatusNoContent)
}

func parseID(ctx *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   errResp("INVALID_ID", "Invalid job type ID"),
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
			Error:   errResp("INVALID_ID", "Invalid job type ID"),
		})
	case errors.Is(err, entity.ErrNotFound):
		ctx.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error:   errResp("NOT_FOUND", "Job type not found"),
		})
	default:
		log.Printf("Job type request failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", internalMessage),
		})
	}
}

func errResp(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}
