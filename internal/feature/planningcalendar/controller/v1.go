package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	dto "backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/feature/planningcalendar/service"

	"github.com/gin-gonic/gin"
)

type Actor = service.Actor

type MonthProjection = service.MonthProjection

type CalendarDay = service.CalendarDay

type CalendarItem = service.CalendarItem

type Controller struct {
	service PlanningService
}

type PlanningService interface {
	GetMonth(ctx context.Context, actor service.Actor, year, month int) (*service.MonthProjection, error)
}

func NewController(s PlanningService) *Controller { return &Controller{service: s} }

func actorFromContext(ctx *gin.Context) (service.Actor, bool) {
	uidVal, ok := ctx.Get("user_id")
	if !ok {
		return service.Actor{}, false
	}
	roleVal, _ := ctx.Get("role")
	uid, ok := uidVal.(uint)
	if !ok {
		return service.Actor{}, false
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
	return service.Actor{UserID: int64(uid), Role: role, TeamID: teamID}, true
}

func errResp(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}

func (c *Controller) GetMonth(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}
	year, err := strconv.Atoi(ctx.Param("year"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_YEAR", "Invalid year")})
		return
	}
	month, err := strconv.Atoi(ctx.Param("month"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_MONTH", "Invalid month")})
		return
	}
	result, err := c.service.GetMonth(ctx.Request.Context(), actor, year, month)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidPeriod) {
			status = http.StatusBadRequest
		}
		ctx.JSON(status, dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: result})
}
