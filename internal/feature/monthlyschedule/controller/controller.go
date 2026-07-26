package controller

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend-hotlines3/internal/feature/monthlyschedule/dto"
	"backend-hotlines3/internal/feature/monthlyschedule/entity"
	"backend-hotlines3/internal/feature/monthlyschedule/mapper"
	"backend-hotlines3/internal/feature/monthlyschedule/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	service        *service.Service
	integrationKey string
}

func New(service *service.Service, integrationKey string) *Controller {
	return &Controller{service: service, integrationKey: strings.TrimSpace(integrationKey)}
}

func (c *Controller) GetWorkspace(ctx *gin.Context) {
	year, month, ok := parsePeriod(ctx)
	if !ok {
		return
	}
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errorInfo("UNAUTHORIZED", "Unauthorized")})
		return
	}
	workspace, err := c.service.GetWorkspace(ctx.Request.Context(), year, month)
	if err != nil {
		writeError(ctx, err)
		return
	}
	if actor.Role != "super_admin" {
		workspace.Draft = nil
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.WorkspaceToResponse(workspace)})
}

func (c *Controller) SaveDraft(ctx *gin.Context) {
	year, month, ok := parsePeriod(ctx)
	if !ok {
		return
	}
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errorInfo("UNAUTHORIZED", "Unauthorized")})
		return
	}
	var request dto.SaveDraftRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errorInfo("VALIDATION_ERROR", err.Error())})
		return
	}
	assignments := make([]entity.Assignment, 0, len(request.Assignments))
	for index, input := range request.Assignments {
		start, err := time.Parse(time.DateOnly, input.StartDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errorInfo("VALIDATION_ERROR", "Invalid startDate at assignment "+strconv.Itoa(index+1))})
			return
		}
		end, err := time.Parse(time.DateOnly, input.EndDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errorInfo("VALIDATION_ERROR", "Invalid endDate at assignment "+strconv.Itoa(index+1))})
			return
		}
		assignments = append(assignments, entity.Assignment{
			TeamID:         input.TeamID,
			AssignmentType: input.AssignmentType,
			StartDate:      start,
			EndDate:        end,
			Destination:    input.Destination,
			Note:           input.Note,
			SourceType:     input.SourceType,
			SourceID:       input.SourceID,
		})
	}
	workspace, err := c.service.SaveDraft(ctx.Request.Context(), actor, year, month, request.ExpectedRevisionNo, assignments)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.WorkspaceToResponse(workspace)})
}

func (c *Controller) Publish(ctx *gin.Context) {
	year, month, ok := parsePeriod(ctx)
	if !ok {
		return
	}
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errorInfo("UNAUTHORIZED", "Unauthorized")})
		return
	}
	var request dto.PublishRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errorInfo("VALIDATION_ERROR", err.Error())})
		return
	}
	projection, err := c.service.Publish(ctx.Request.Context(), actor, year, month, request.ExpectedRevisionNo)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ProjectionToResponse(projection)})
}

func (c *Controller) GetPublishedForClinicTool(ctx *gin.Context) {
	requestID := strings.TrimSpace(ctx.GetHeader("X-Request-ID"))
	if requestID == "" {
		requestID = uuid.NewString()
	}
	ctx.Header("X-Request-ID", requestID)

	if c.integrationKey == "" {
		ctx.JSON(http.StatusServiceUnavailable, dto.StandardResponse{Success: false, Error: errorInfo("INTEGRATION_NOT_CONFIGURED", "Clinic Tool integration is not configured")})
		return
	}
	provided := strings.TrimSpace(ctx.GetHeader("X-Integration-Key"))
	if subtle.ConstantTimeCompare([]byte(provided), []byte(c.integrationKey)) != 1 {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errorInfo("UNAUTHORIZED", "Invalid integration key")})
		return
	}
	year, month, ok := parsePeriod(ctx)
	if !ok {
		return
	}
	projection, err := c.service.GetPublished(ctx.Request.Context(), year, month)
	if err != nil {
		writeError(ctx, err)
		return
	}
	etag := `"` + projection.Checksum + `"`
	ctx.Header("ETag", etag)
	ctx.Header("Cache-Control", "private, max-age=0, must-revalidate")
	if ctx.GetHeader("If-None-Match") == etag {
		ctx.Status(http.StatusNotModified)
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ProjectionToResponse(projection)})
}

func actorFromContext(ctx *gin.Context) (entity.Actor, bool) {
	userValue, ok := ctx.Get("user_id")
	if !ok {
		return entity.Actor{}, false
	}
	var userID int64
	switch value := userValue.(type) {
	case uint:
		userID = int64(value)
	case int64:
		userID = value
	case int:
		userID = int64(value)
	default:
		return entity.Actor{}, false
	}
	role, _ := ctx.Get("role")
	roleString, _ := role.(string)
	return entity.Actor{UserID: userID, Role: roleString}, userID > 0
}

func parsePeriod(ctx *gin.Context) (int, int, bool) {
	year, yearErr := strconv.Atoi(ctx.Param("year"))
	month, monthErr := strconv.Atoi(ctx.Param("month"))
	if yearErr != nil || monthErr != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errorInfo("INVALID_PERIOD", "Invalid year or month")})
		return 0, 0, false
	}
	return year, month, true
}

func writeError(ctx *gin.Context, err error) {
	status, code := http.StatusInternalServerError, "INTERNAL_ERROR"
	switch {
	case errors.Is(err, entity.ErrInvalidPeriod),
		errors.Is(err, entity.ErrInvalidAssignment),
		errors.Is(err, entity.ErrOverlappingAssignment),
		errors.Is(err, entity.ErrTeamMetadataMissing):
		status, code = http.StatusBadRequest, "VALIDATION_ERROR"
	case errors.Is(err, entity.ErrForbidden):
		status, code = http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, entity.ErrDraftNotFound),
		errors.Is(err, entity.ErrPublishedNotFound):
		status, code = http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, entity.ErrRevisionConflict):
		status, code = http.StatusConflict, "REVISION_CONFLICT"
	}
	message := err.Error()
	if status == http.StatusInternalServerError {
		message = "An error occurred while processing the monthly schedule"
	}
	ctx.JSON(status, dto.StandardResponse{Success: false, Error: errorInfo(code, message)})
}

func errorInfo(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}
