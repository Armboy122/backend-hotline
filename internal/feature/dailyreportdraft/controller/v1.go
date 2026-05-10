package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	dto "backend-hotlines3/internal/dto"
	service "backend-hotlines3/internal/feature/dailyreportdraft/service"
	taskentity "backend-hotlines3/internal/feature/task/entity"

	"github.com/gin-gonic/gin"
)

type Service interface {
	ListSources(ctx context.Context, actor taskentity.Actor, input service.ListSourcesInput) (*service.ListSourcesOutput, error)
	FromPlan(ctx context.Context, actor taskentity.Actor, input service.FromPlanInput) (*service.FromPlanOutput, error)
}

type Controller struct {
	service Service
}

func NewController(s Service) *Controller { return &Controller{service: s} }

func actorFromContext(ctx *gin.Context) (taskentity.Actor, bool) {
	uidVal, ok := ctx.Get("user_id")
	if !ok {
		return taskentity.Actor{}, false
	}
	roleVal, _ := ctx.Get("role")
	uid, ok := uidVal.(uint)
	if !ok {
		return taskentity.Actor{}, false
	}
	role, _ := roleVal.(string)
	var teamID *int64
	if rawTeamID, exists := ctx.Get("team_id"); exists {
		switch v := rawTeamID.(type) {
		case int64:
			teamID = &v
		case int:
			id := int64(v)
			teamID = &id
		case uint:
			id := int64(v)
			teamID = &id
		case float64:
			id := int64(v)
			teamID = &id
		}
	}
	return taskentity.Actor{UserID: int64(uid), Role: role, TeamID: teamID}, true
}

func errResp(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}

func mapErr(err error) int {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, service.ErrWorkDateRequired), errors.Is(err, service.ErrTeamIDRequired), errors.Is(err, service.ErrUnsupportedSourceType):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func parseDate(raw string) (time.Time, error) { return time.Parse("2006-01-02", raw) }

func (c *Controller) ListSources(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}
	workDate, err := parseDate(strings.TrimSpace(ctx.Query("workDate")))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_DATE", "Invalid work date format. Use YYYY-MM-DD")})
		return
	}
	var teamID *int64
	if raw := strings.TrimSpace(ctx.Query("teamId")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_TEAM_ID", "Invalid teamId")})
			return
		}
		teamID = &parsed
	}
	if teamID == nil {
		teamID = actor.TeamID
	}
	out, err := c.service.ListSources(ctx.Request.Context(), actor, service.ListSourcesInput{WorkDate: &workDate, TeamID: teamID})
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: out})
}

func (c *Controller) FromPlan(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}
	var sourceID int64
	if raw := strings.TrimSpace(ctx.Query("sourceId")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_SOURCE_ID", "Invalid sourceId")})
			return
		}
		sourceID = parsed
	}
	var workDate *time.Time
	if raw := strings.TrimSpace(ctx.Query("workDate")); raw != "" {
		parsed, err := parseDate(raw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_DATE", "Invalid work date format. Use YYYY-MM-DD")})
			return
		}
		workDate = &parsed
	}
	out, err := c.service.FromPlan(ctx.Request.Context(), actor, service.FromPlanInput{SourceType: strings.TrimSpace(ctx.Query("sourceType")), SourceID: sourceID, WorkDate: workDate})
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: out})
}
