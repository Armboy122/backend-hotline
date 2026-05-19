package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	dto "backend-hotlines3/internal/dto"
	teamplandto "backend-hotlines3/internal/feature/teamplan/dto"
	"backend-hotlines3/internal/feature/teamplan/entity"
	"backend-hotlines3/internal/feature/teamplan/service"

	"github.com/gin-gonic/gin"
)

type TeamPlanService interface {
	List(ctx context.Context, actor entity.Actor, input service.ListInput) (*service.ListOutput, error)
	GetByID(ctx context.Context, actor entity.Actor, id int64) (*entity.TeamPlan, error)
	Create(ctx context.Context, actor entity.Actor, input service.CreateInput) (*entity.TeamPlan, error)
	Update(ctx context.Context, actor entity.Actor, input service.UpdateInput) (*entity.TeamPlan, error)
	Delete(ctx context.Context, actor entity.Actor, id int64) error
}

type Controller struct {
	service TeamPlanService
}

func NewController(s TeamPlanService) *Controller { return &Controller{service: s} }

func actorFromContext(ctx *gin.Context) (entity.Actor, bool) {
	uidVal, ok := ctx.Get("user_id")
	if !ok {
		return entity.Actor{}, false
	}
	roleVal, _ := ctx.Get("role")
	uid, ok := uidVal.(uint)
	if !ok {
		return entity.Actor{}, false
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
	return entity.Actor{UserID: int64(uid), Role: role, TeamID: teamID}, true
}

func errResp(code, message string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: message}
}

func mapErr(err error) int {
	switch {
	case errors.Is(err, entity.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, entity.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, entity.ErrInvalidID), errors.Is(err, entity.ErrInvalidInput), errors.Is(err, entity.ErrInvalidStatus), errors.Is(err, entity.ErrInvalidRange):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func parseDate(raw string) (time.Time, error) { return time.Parse("2006-01-02", raw) }

func parseOptionalDate(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	parsed, err := parseDate(*raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (c *Controller) List(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}
	var from, to time.Time
	var err error
	if raw := strings.TrimSpace(ctx.Query("from")); raw != "" {
		from, err = parseDate(raw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_DATE", "Invalid from date format. Use YYYY-MM-DD")})
			return
		}
	}
	if raw := strings.TrimSpace(ctx.Query("to")); raw != "" {
		to, err = parseDate(raw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_DATE", "Invalid to date format. Use YYYY-MM-DD")})
			return
		}
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	var teamID *int64
	if raw := strings.TrimSpace(ctx.Query("teamId")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_TEAM_ID", "Invalid teamId")})
			return
		}
		teamID = &parsed
	}
	var statuses []string
	if raw := strings.TrimSpace(ctx.Query("status")); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			if v := strings.TrimSpace(part); v != "" {
				statuses = append(statuses, v)
			}
		}
	}
	out, err := c.service.List(ctx.Request.Context(), actor, service.ListInput{From: from, To: to, TeamID: teamID, Status: statuses, Page: page, Limit: limit})
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	resp := make([]teamplandto.TeamPlanResponse, 0, len(out.Items))
	for i := range out.Items {
		resp = append(resp, toResponse(out.Items[i], actor))
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: resp, Meta: &dto.Meta{Page: out.Page, Limit: out.Limit, Total: out.Total}})
}

func (c *Controller) GetByID(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_ID", "Invalid team plan ID")})
		return
	}
	item, err := c.service.GetByID(ctx.Request.Context(), actor, id)
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toResponse(*item, actor)})
}

func (c *Controller) Create(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}
	var req teamplandto.CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("VALIDATION_ERROR", err.Error())})
		return
	}
	var startDate *time.Time
	if req.StartDate != nil && strings.TrimSpace(*req.StartDate) != "" {
		parsed, err := parseDate(strings.TrimSpace(*req.StartDate))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_DATE", "Invalid start date format. Use YYYY-MM-DD")})
			return
		}
		startDate = &parsed
	}
	endDate, err := parseOptionalDate(req.EndDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_DATE", "Invalid end date format. Use YYYY-MM-DD")})
		return
	}
	item, err := c.service.Create(ctx.Request.Context(), actor, service.CreateInput{
		TeamID:            req.TeamID,
		Title:             req.Title,
		WorkType:          req.WorkType,
		StartDate:         startDate,
		EndDate:           endDate,
		WorkTime:          req.WorkTime,
		LocationText:      req.LocationText,
		PEAID:             req.PEAID,
		OperationCenterID: req.OperationCenterID,
		FeederID:          req.FeederID,
		StationID:         req.StationID,
		Notes:             req.Notes,
	})
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	ctx.JSON(http.StatusCreated, dto.StandardResponse{Success: true, Data: toResponse(*item, actor)})
}

func (c *Controller) Update(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_ID", "Invalid team plan ID")})
		return
	}
	var req teamplandto.UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("VALIDATION_ERROR", err.Error())})
		return
	}
	if req.Status != nil {
		status := strings.TrimSpace(*req.Status)
		if !entity.IsValidStatus(status) {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_STATUS", "Invalid team plan status")})
			return
		}
	}
	var startDate *time.Time
	if req.StartDate != nil && strings.TrimSpace(*req.StartDate) != "" {
		parsed, err := parseDate(strings.TrimSpace(*req.StartDate))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_DATE", "Invalid start date format. Use YYYY-MM-DD")})
			return
		}
		startDate = &parsed
	}
	endDate, err := parseOptionalDate(req.EndDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_DATE", "Invalid end date format. Use YYYY-MM-DD")})
		return
	}
	item, err := c.service.Update(ctx.Request.Context(), actor, service.UpdateInput{ID: id, TeamID: req.TeamID, Title: req.Title, WorkType: req.WorkType, StartDate: startDate, EndDate: endDate, WorkTime: req.WorkTime, LocationText: req.LocationText, PEAID: req.PEAID, OperationCenterID: req.OperationCenterID, FeederID: req.FeederID, StationID: req.StationID, Notes: req.Notes, Status: req.Status})
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toResponse(*item, actor)})
}

func (c *Controller) Delete(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("INVALID_ID", "Invalid team plan ID")})
		return
	}
	if err := c.service.Delete(ctx.Request.Context(), actor, id); err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func toResponse(plan entity.TeamPlan, actor entity.Actor) teamplandto.TeamPlanResponse {
	resp := teamplandto.TeamPlanResponse{
		ID:                plan.ID,
		TeamID:            plan.TeamID,
		CreatedByUserID:   plan.CreatedByUserID,
		Title:             plan.Title,
		WorkType:          plan.WorkType,
		WorkTime:          plan.WorkTime,
		LocationText:      plan.LocationText,
		PEAID:             plan.PEAID,
		OperationCenterID: plan.OperationCenterID,
		FeederID:          plan.FeederID,
		StationID:         plan.StationID,
		Notes:             plan.Notes,
		Status:            plan.Status,
		DailyTaskID:       plan.DailyTaskID,
		CreatedAt:         plan.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         plan.UpdatedAt.UTC().Format(time.RFC3339),
		Actions:           teamplandto.TeamPlanActionRef{CanEdit: actor.CanUpdateTeamPlan(plan), CanDelete: actor.CanDeleteTeamPlan(plan)},
	}
	if plan.StartDate != nil {
		start := plan.StartDate.UTC().Format("2006-01-02")
		resp.StartDate = &start
	}
	if plan.EndDate != nil {
		end := plan.EndDate.UTC().Format("2006-01-02")
		resp.EndDate = &end
	}
	if plan.DeletedAt != nil {
		deleted := plan.DeletedAt.UTC().Format(time.RFC3339)
		resp.DeletedAt = &deleted
	}
	if plan.Team != nil {
		resp.Team = &teamplandto.TeamSummaryRef{ID: plan.Team.ID, Name: plan.Team.Name}
	}
	if plan.CreatedBy != nil {
		resp.CreatedBy = &teamplandto.UserSummaryRef{ID: plan.CreatedBy.ID, Username: plan.CreatedBy.Username, DisplayName: plan.CreatedBy.DisplayName}
	}
	return resp
}
