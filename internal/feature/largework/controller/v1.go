package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	dto "backend-hotlines3/internal/feature/largework/dto"
	"backend-hotlines3/internal/feature/largework/entity"
	"backend-hotlines3/internal/feature/largework/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service service.Service
}

func NewController(s service.Service) *Controller { return &Controller{service: s} }

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

func parseOptionalDate(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", *raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseDate(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", raw)
}

func toStringPtr(v **string) *string {
	if v == nil {
		return nil
	}
	return *v
}

func toInt64Ptr(v **int64) *int64 {
	if v == nil {
		return nil
	}
	return *v
}

func toResponse(item *entity.LargeWorkItem, actor entity.Actor) dto.LargeWorkResponse {
	resp := dto.LargeWorkResponse{
		ID:                item.ID,
		OwnerTeamID:       item.OwnerTeamID,
		CreatedByUserID:   item.CreatedByUserID,
		Title:             item.Title,
		WorkType:          item.WorkType,
		StartDate:         item.StartDate.Format("2006-01-02"),
		WorkTime:          item.WorkTime,
		LocationText:      item.LocationText,
		PEAID:             item.PEAID,
		OperationCenterID: item.OperationCenterID,
		FeederID:          item.FeederID,
		StationID:         item.StationID,
		Notes:             item.Notes,
		Status:            item.Status,
		CreatedAt:         item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         item.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt:         nil,
	}
	if item.EndDate != nil {
		end := item.EndDate.Format("2006-01-02")
		resp.EndDate = &end
	}
	if item.DeletedAt != nil {
		deleted := item.DeletedAt.UTC().Format(time.RFC3339)
		resp.DeletedAt = &deleted
	}
	teams := make([]dto.TeamRef, 0, len(item.Teams))
	for _, team := range item.Teams {
		teams = append(teams, dto.TeamRef{ID: team.ID, Name: team.Name, Role: team.Role})
	}
	resp.Teams = teams
	resp.Actions.CanEdit = actor.CanUpdateLargeWork(item)
	resp.Actions.CanCancel = actor.CanCancelLargeWork(item)
	resp.Actions.CanAssignTasks = actor.CanAssignLargeWorkTasks(item, 0)
	resp.Actions.CanStartDailyReport = false
	return resp
}

func toTaskResponse(task entity.LargeWorkTask) dto.LargeWorkTaskResponse {
	resp := dto.LargeWorkTaskResponse{
		ID: task.ID, LargeWorkItemID: task.LargeWorkItemID, AssignedTeamID: task.AssignedTeamID, Sequence: task.Sequence,
		PointLabel: task.PointLabel, Latitude: task.Latitude, Longitude: task.Longitude, WorkType: task.WorkType,
		WorkDetail: task.WorkDetail, PointCount: task.PointCount, TreeCount: task.TreeCount, ItemCount: task.ItemCount,
		Notes: task.Notes, Status: task.Status, BeforePhotoURLs: task.BeforePhotoURLs, AfterPhotoURLs: task.AfterPhotoURLs,
		CompletionNote: task.CompletionNote, StartedByUserID: task.StartedByUserID, CompletedByUserID: task.CompletedByUserID, Metadata: task.Metadata,
	}
	if resp.BeforePhotoURLs == nil {
		resp.BeforePhotoURLs = []string{}
	}
	if resp.AfterPhotoURLs == nil {
		resp.AfterPhotoURLs = []string{}
	}
	if task.StartedAt != nil {
		v := task.StartedAt.UTC().Format(time.RFC3339)
		resp.StartedAt = &v
	}
	if task.CompletedAt != nil {
		v := task.CompletedAt.UTC().Format(time.RFC3339)
		resp.CompletedAt = &v
	}
	return resp
}

func toTaskResponses(tasks []entity.LargeWorkTask) []dto.LargeWorkTaskResponse {
	out := make([]dto.LargeWorkTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, toTaskResponse(task))
	}
	return out
}

func toOverviewResponse(out *service.OverviewOutput, actor entity.Actor) dto.OverviewResponse {
	teamProgress := make([]dto.TeamProgressResponse, 0, len(out.TeamProgress))
	for _, team := range out.TeamProgress {
		teamProgress = append(teamProgress, dto.TeamProgressResponse{
			AssignedTeamID: team.AssignedTeamID,
			Total:          team.Total,
			Todo:           team.Todo,
			InProgress:     team.InProgress,
			Done:           team.Done,
			Blocked:        team.Blocked,
			Cancelled:      team.Cancelled,
		})
	}
	return dto.OverviewResponse{
		Plan: toResponse(&out.Plan, actor),
		Progress: dto.OverviewProgressResponse{
			Total:      out.Progress.Total,
			Todo:       out.Progress.Todo,
			InProgress: out.Progress.InProgress,
			Done:       out.Progress.Done,
			Blocked:    out.Progress.Blocked,
			Cancelled:  out.Progress.Cancelled,
		},
		TeamProgress: teamProgress,
	}
}

func parseIDParam(ctx *gin.Context, name string, message string) (int64, bool) {
	id, err := strconv.ParseInt(ctx.Param(name), 10, 64)
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_ID", Message: message}})
		return 0, false
	}
	return id, true
}

func mapErr(err error) int {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, service.ErrTaskSchemaUnavailable):
		return http.StatusServiceUnavailable
	case errors.Is(err, service.ErrInvalidID), errors.Is(err, service.ErrInvalidOwnerTeam), errors.Is(err, service.ErrInvalidTitle), errors.Is(err, service.ErrInvalidLocation), errors.Is(err, service.ErrInvalidStartDate), errors.Is(err, service.ErrInvalidParticipantTeams), errors.Is(err, service.ErrInvalidStatus), errors.Is(err, service.ErrInvalidDateRange), errors.Is(err, service.ErrInvalidStateTransition), errors.Is(err, service.ErrInvalidTask), errors.Is(err, service.ErrPhotoRequired):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (c *Controller) Create(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	var req dto.CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()}})
		return
	}
	startDate, err := parseDate(req.StartDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_DATE", Message: "Invalid start date format. Use YYYY-MM-DD"}})
		return
	}
	endDate, err := parseOptionalDate(req.EndDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_DATE", Message: "Invalid end date format. Use YYYY-MM-DD"}})
		return
	}
	if endDate != nil && endDate.Before(startDate) {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_DATE_RANGE", Message: service.ErrInvalidDateRange.Error()}})
		return
	}
	item, err := c.service.Create(ctx.Request.Context(), actor, service.CreateInput{
		OwnerTeamID:        req.OwnerTeamID,
		ParticipantTeamIDs: req.ParticipantTeamIDs,
		Title:              req.Title,
		WorkType:           req.WorkType,
		StartDate:          startDate,
		EndDate:            endDate,
		WorkTime:           req.WorkTime,
		LocationText:       req.LocationText,
		PEAID:              req.PEAID,
		OperationCenterID:  req.OperationCenterID,
		FeederID:           req.FeederID,
		StationID:          req.StationID,
		Notes:              req.Notes,
	})
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusCreated, dto.StandardResponse{Success: true, Data: toResponse(item, actor)})
}

func (c *Controller) GetByID(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid large work item ID"}})
		return
	}
	item, err := c.service.GetByID(ctx.Request.Context(), actor, id)
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toResponse(item, actor)})
}

func (c *Controller) List(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	var from *time.Time
	if raw := ctx.Query("from"); raw != "" {
		parsed, err := parseDate(raw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_DATE", Message: "Invalid from date format. Use YYYY-MM-DD"}})
			return
		}
		from = &parsed
	}
	var to *time.Time
	if raw := ctx.Query("to"); raw != "" {
		parsed, err := parseDate(raw)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_DATE", Message: "Invalid to date format. Use YYYY-MM-DD"}})
			return
		}
		to = &parsed
	}
	var teamID *int64
	if raw := ctx.Query("teamId"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_TEAM_ID", Message: "Invalid teamId"}})
			return
		}
		teamID = &id
	}
	statusCSV := ctx.Query("status")
	statuses := []string{}
	if statusCSV != "" {
		for _, part := range strings.Split(statusCSV, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				statuses = append(statuses, trimmed)
			}
		}
	}
	out, err := c.service.List(ctx.Request.Context(), actor, service.ListInput{Page: page, Limit: limit, From: from, To: to, TeamID: teamID, Statuses: statuses})
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	resp := make([]dto.LargeWorkResponse, 0, len(out.Items))
	for i := range out.Items {
		resp = append(resp, toResponse(&out.Items[i], actor))
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: resp, Meta: &dto.Meta{Page: out.Page, Limit: out.Limit, Total: out.Total}})
}

func (c *Controller) Update(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid large work item ID"}})
		return
	}
	var req dto.UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()}})
		return
	}
	var startDate *time.Time
	if req.StartDate != nil {
		parsed, err := parseDate(*req.StartDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_DATE", Message: "Invalid start date format. Use YYYY-MM-DD"}})
			return
		}
		startDate = &parsed
	}
	var endDate *time.Time
	if req.EndDate != nil {
		endDate, err = parseOptionalDate(req.EndDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_DATE", Message: "Invalid end date format. Use YYYY-MM-DD"}})
			return
		}
	}
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_DATE_RANGE", Message: service.ErrInvalidDateRange.Error()}})
		return
	}
	item, err := c.service.Update(ctx.Request.Context(), actor, service.UpdateInput{ID: id, OwnerTeamID: req.OwnerTeamID, ParticipantTeamIDs: req.ParticipantTeamIDs, Title: req.Title, WorkType: req.WorkType, StartDate: startDate, EndDate: endDate, WorkTime: req.WorkTime, LocationText: req.LocationText, PEAID: req.PEAID, OperationCenterID: req.OperationCenterID, FeederID: req.FeederID, StationID: req.StationID, Notes: req.Notes, Status: req.Status})
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toResponse(item, actor)})
}

func (c *Controller) Cancel(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid large work item ID"}})
		return
	}
	item, err := c.service.Cancel(ctx.Request.Context(), actor, id)
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toResponse(item, actor)})
}

func (c *Controller) Attachments(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "NOT_IMPLEMENTED", Message: "attachments are not ready yet"}})
}

func (c *Controller) GetOverview(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	id, ok := parseIDParam(ctx, "id", "Invalid large work item ID")
	if !ok {
		return
	}
	out, err := c.service.GetOverview(ctx.Request.Context(), actor, id)
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toOverviewResponse(out, actor)})
}

func (c *Controller) ReplaceTasks(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	id, ok := parseIDParam(ctx, "id", "Invalid large work item ID")
	if !ok {
		return
	}
	var req dto.ReplaceTasksRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()}})
		return
	}
	input := service.ReplaceTasksInput{Tasks: make([]service.TaskPointInput, 0, len(req.Tasks))}
	for _, t := range req.Tasks {
		input.Tasks = append(input.Tasks, service.TaskPointInput{AssignedTeamID: t.AssignedTeamID, Sequence: t.Sequence, PointLabel: t.PointLabel, Latitude: t.Latitude, Longitude: t.Longitude, WorkType: t.WorkType, WorkDetail: t.WorkDetail, PointCount: t.PointCount, TreeCount: t.TreeCount, ItemCount: t.ItemCount, Notes: t.Notes, BeforePhotoURLs: t.BeforePhotoURLs, Metadata: t.Metadata})
	}
	tasks, err := c.service.ReplaceTasks(ctx.Request.Context(), actor, id, input)
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toTaskResponses(tasks)})
}

func (c *Controller) ListTasks(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	id, ok := parseIDParam(ctx, "id", "Invalid large work item ID")
	if !ok {
		return
	}
	tasks, err := c.service.ListTasks(ctx.Request.Context(), actor, id)
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toTaskResponses(tasks)})
}

func (c *Controller) MyTodos(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	out, err := c.service.MyTodos(ctx.Request.Context(), actor, service.MyTodosInput{Page: page, Limit: limit})
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toTaskResponses(out.Items), Meta: &dto.Meta{Page: out.Page, Limit: out.Limit, Total: out.Total}})
}

func (c *Controller) StartTask(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	id, ok := parseIDParam(ctx, "taskId", "Invalid large work task ID")
	if !ok {
		return
	}
	task, err := c.service.StartTask(ctx.Request.Context(), actor, id)
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toTaskResponse(*task)})
}

func (c *Controller) AttachTaskPhoto(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	id, ok := parseIDParam(ctx, "taskId", "Invalid large work task ID")
	if !ok {
		return
	}
	var req dto.TaskExecutionPhotoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()}})
		return
	}
	task, err := c.service.AttachTaskPhoto(ctx.Request.Context(), actor, id, req.Kind, req.URL)
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toTaskResponse(*task)})
}

func (c *Controller) CompleteTask(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	id, ok := parseIDParam(ctx, "taskId", "Invalid large work task ID")
	if !ok {
		return
	}
	var req dto.CompleteTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()}})
		return
	}
	task, err := c.service.CompleteTask(ctx.Request.Context(), actor, id, service.CompleteTaskInput{AfterPhotoURLs: req.AfterPhotoURLs, CompletionNote: req.CompletionNote})
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toTaskResponse(*task)})
}

func (c *Controller) BlockTask(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"}})
		return
	}
	id, ok := parseIDParam(ctx, "taskId", "Invalid large work task ID")
	if !ok {
		return
	}
	var req dto.BlockTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()}})
		return
	}
	task, err := c.service.BlockTask(ctx.Request.Context(), actor, id, req.Reason)
	if err != nil {
		ctx.JSON(mapErr(err), dto.StandardResponse{Success: false, Error: &dto.ErrorInfo{Code: "ERROR", Message: err.Error()}})
		return
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toTaskResponse(*task)})
}
