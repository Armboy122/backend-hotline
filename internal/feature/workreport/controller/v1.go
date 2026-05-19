package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	dto "backend-hotlines3/internal/dto"
	taskentity "backend-hotlines3/internal/feature/task/entity"
	service "backend-hotlines3/internal/feature/workreport/service"

	"github.com/gin-gonic/gin"
)

type Service interface {
	GetReport(ctx context.Context, actor taskentity.Actor, input service.GetReportInput) (*service.GetReportOutput, error)
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
	uid, ok := uidVal.(uint)
	if !ok {
		return taskentity.Actor{}, false
	}
	roleVal, _ := ctx.Get("role")
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
	case errors.Is(err, service.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, service.ErrYearMonthRequired):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func parseReportQuery(ctx *gin.Context) (service.GetReportInput, *dto.ErrorInfo) {
	year := strings.TrimSpace(ctx.Query("year"))
	month := strings.TrimSpace(ctx.Query("month"))
	if year == "" && len(month) == len("2006-01") && month[4] == '-' {
		year = month[:4]
		month = month[5:]
	}
	if len(year) != 4 {
		return service.GetReportInput{}, errResp("INVALID_YEAR_MONTH", "year and month are required")
	}
	yearNum, err := strconv.Atoi(year)
	if err != nil || yearNum < 1900 || yearNum > 2999 {
		return service.GetReportInput{}, errResp("INVALID_YEAR", "Invalid year")
	}
	monthNum, err := strconv.Atoi(month)
	if err != nil || monthNum < 1 || monthNum > 12 {
		return service.GetReportInput{}, errResp("INVALID_MONTH", "Invalid month")
	}
	month = strconv.Itoa(monthNum)

	var teamID *int64
	if raw := strings.TrimSpace(ctx.Query("teamId")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return service.GetReportInput{}, errResp("INVALID_TEAM_ID", "Invalid teamId")
		}
		teamID = &parsed
	}
	return service.GetReportInput{Year: year, Month: month, TeamID: teamID}, nil
}

func (c *Controller) GetReport(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	input, parseErr := parseReportQuery(ctx)
	if parseErr != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: parseErr})
		return
	}

	out, err := c.service.GetReport(ctx.Request.Context(), actor, input)
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: out})
}
