package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"backend-hotlines3/internal/feature/dashboard/dto"
	"backend-hotlines3/internal/feature/dashboard/entity"
	"backend-hotlines3/internal/feature/dashboard/mapper"

	"github.com/gin-gonic/gin"
)

func (c *Controller) Summary(ctx *gin.Context) {
	result, err := c.service.Summary(ctx.Request.Context(), entity.SummaryFilter{
		Year:      ctx.Query("year"),
		Month:     ctx.Query("month"),
		TeamID:    ctx.Query("teamId"),
		JobTypeID: ctx.Query("jobTypeId"),
	})
	if err != nil {
		log.Printf("Dashboard summary error: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", "Failed to load dashboard data"),
		})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToSummaryResponse(result)})
}

func (c *Controller) TopJobs(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	result, err := c.service.TopJobs(ctx.Request.Context(), entity.TopFilter{
		Year:      ctx.Query("year"),
		Month:     ctx.Query("month"),
		TeamID:    ctx.Query("teamId"),
		JobTypeID: ctx.Query("jobTypeId"),
		Limit:     limit,
	})
	if err != nil {
		log.Printf("Dashboard top-jobs error: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", "Failed to load top jobs"),
		})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToTopJobResponses(result)})
}

func (c *Controller) TopFeeders(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	result, err := c.service.TopFeeders(ctx.Request.Context(), entity.TopFilter{
		Year:      ctx.Query("year"),
		Month:     ctx.Query("month"),
		TeamID:    ctx.Query("teamId"),
		JobTypeID: ctx.Query("jobTypeId"),
		Limit:     limit,
	})
	if err != nil {
		log.Printf("Dashboard top-feeders error: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", "Failed to load top feeders"),
		})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToTopFeederResponses(result)})
}

func (c *Controller) FeederMatrix(ctx *gin.Context) {
	feederIDStr := ctx.Query("feederId")
	if feederIDStr == "" {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   errResp("VALIDATION_ERROR", "feederId is required"),
		})
		return
	}
	feederID, err := strconv.ParseInt(feederIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   errResp("INVALID_ID", "Invalid feeder ID"),
		})
		return
	}

	result, err := c.service.FeederMatrix(ctx.Request.Context(), entity.MatrixFilter{
		FeederID: feederID,
		Year:     ctx.Query("year"),
		Month:    ctx.Query("month"),
	})
	if err != nil {
		if errors.Is(err, entity.ErrFeederNotFound) {
			ctx.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error:   errResp("NOT_FOUND", "Feeder not found"),
			})
			return
		}
		log.Printf("Dashboard feeder-matrix error: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", "Failed to load feeder matrix"),
		})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToFeederMatrixResponse(result)})
}

func (c *Controller) Stats(ctx *gin.Context) {
	result, err := c.service.Stats(ctx.Request.Context(), entity.StatsFilter{
		StartDate: ctx.Query("startDate"),
		EndDate:   ctx.Query("endDate"),
		TeamID:    ctx.Query("teamId"),
		FeederID:  ctx.Query("feederId"),
	})
	if err != nil {
		log.Printf("Dashboard stats error: %v", err)
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   errResp("INTERNAL_ERROR", "Failed to load dashboard stats"),
		})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToStatsResponse(result)})
}

func errResp(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}
