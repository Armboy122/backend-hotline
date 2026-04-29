package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"backend-hotlines3/internal/feature/station/dto"
	"backend-hotlines3/internal/feature/station/entity"
	"backend-hotlines3/internal/feature/station/mapper"

	"github.com/gin-gonic/gin"
)

func (c *Controller) List(ctx *gin.Context) {
	items, err := c.service.List(ctx.Request.Context())
	if err != nil {
		log.Printf("Failed to fetch stations: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{Success: false, Error: errResp("INTERNAL_ERROR", "An error occurred while fetching stations")})
		return
	}
	response := make([]dto.StationResponse, 0, len(items))
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
		writeError(ctx, err, "An error occurred while fetching the station")
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToResponse(*item)})
}

func (c *Controller) Create(ctx *gin.Context) {
	var req dto.CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("VALIDATION_ERROR", err.Error())})
		return
	}
	item, err := c.service.Create(ctx.Request.Context(), entity.CreateInput{
		Name:        req.Name,
		CodeName:    req.CodeName,
		OperationID: req.OperationID,
	})
	if err != nil {
		log.Printf("Failed to create station: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{Success: false, Error: errResp("INTERNAL_ERROR", "An error occurred while creating the station")})
		return
	}
	ctx.JSON(http.StatusCreated, dto.StandardResponse{Success: true, Data: mapper.ToResponse(*item)})
}

func (c *Controller) Update(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}
	var req dto.UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("VALIDATION_ERROR", err.Error())})
		return
	}
	item, err := c.service.Update(ctx.Request.Context(), entity.UpdateInput{
		ID:          id,
		Name:        req.Name,
		CodeName:    req.CodeName,
		OperationID: req.OperationID,
	})
	if err != nil {
		writeError(ctx, err, "An error occurred while updating the station")
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
		writeError(ctx, err, "An error occurred while deleting the station")
		return
	}
	ctx.Status(http.StatusNoContent)
}

func parseID(ctx *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_ID", "Invalid station ID")})
		return 0, false
	}
	return id, true
}

func writeError(ctx *gin.Context, err error, internalMessage string) {
	switch {
	case errors.Is(err, entity.ErrInvalidID):
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_ID", "Invalid station ID")})
	case errors.Is(err, entity.ErrNotFound):
		ctx.JSON(http.StatusNotFound, dto.StandardResponse{Success: false, Error: errResp("NOT_FOUND", "Station not found")})
	default:
		log.Printf("Station request failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{Success: false, Error: errResp("INTERNAL_ERROR", internalMessage)})
	}
}

func errResp(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}
