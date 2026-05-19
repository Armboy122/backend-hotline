package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	taskdto "backend-hotlines3/internal/feature/task/dto"
	taskentity "backend-hotlines3/internal/feature/task/entity"
	taskrepository "backend-hotlines3/internal/feature/task/repository"
	taskservice "backend-hotlines3/internal/feature/task/service"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	capturedActor      taskentity.Actor
	capturedList       taskservice.ListTasksInput
	capturedByFilter   taskservice.ListTasksByFilterInput
	listOutput         *taskservice.ListTasksOutput
	listErr            error
	listByFilterOutput *taskservice.ListTasksByFilterOutput
	listByFilterErr    error
	createResult       *taskentity.Task
	createErr          error
}

func (f *fakeService) List(_ context.Context, actor taskentity.Actor, input taskservice.ListTasksInput) (*taskservice.ListTasksOutput, error) {
	f.capturedActor = actor
	f.capturedList = input
	return f.listOutput, f.listErr
}
func (f *fakeService) GetByID(context.Context, taskentity.Actor, int64) (*taskentity.Task, error) {
	return nil, nil
}
func (f *fakeService) Create(context.Context, taskentity.Actor, taskservice.CreateTaskInput) (*taskentity.Task, error) {
	return f.createResult, f.createErr
}
func (f *fakeService) Update(context.Context, taskentity.Actor, taskservice.UpdateTaskInput) (*taskentity.Task, error) {
	return nil, nil
}
func (f *fakeService) Delete(context.Context, taskentity.Actor, int64) error { return nil }
func (f *fakeService) ListByTeam(context.Context, taskentity.Actor, taskservice.ListTasksByTeamInput) (*taskservice.ListTasksByTeamOutput, error) {
	return nil, nil
}
func (f *fakeService) ListByFilter(_ context.Context, _ taskentity.Actor, input taskservice.ListTasksByFilterInput) (*taskservice.ListTasksByFilterOutput, error) {
	f.capturedByFilter = input
	return f.listByFilterOutput, f.listByFilterErr
}

func TestListReturnsStandardEnvelopeAndNestedTaskData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workDate := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 28, 12, 30, 0, 0, time.UTC)
	feederID := int64(7)
	operationCenterID := int64(55)
	sourceType := "large_work"
	sourceID := int64(501)
	largeWorkTaskID := int64(9001)
	task := taskentity.Task{
		ID:                  101,
		WorkDate:            workDate,
		TeamID:              11,
		JobTypeID:           22,
		JobDetailID:         33,
		FeederID:            &feederID,
		URLsBefore:          []string{"before-1"},
		URLsAfter:           []string{"after-1"},
		SourceType:          &sourceType,
		SourceID:            &sourceID,
		LargeWorkTaskID:     &largeWorkTaskID,
		CreatedAt:           now,
		UpdatedAt:           now,
		TeamName:            stringPtr("Team Alpha"),
		JobTypeName:         stringPtr("Inspection"),
		JobDetailName:       stringPtr("Fuse Check"),
		FeederCode:          stringPtr("F-007"),
		StationName:         stringPtr("Station North"),
		OperationCenterID:   &operationCenterID,
		OperationCenterName: stringPtr("Center A"),
	}
	service := &fakeService{
		listOutput: &taskservice.ListTasksOutput{
			Tasks: []taskentity.Task{task},
			Total: 1,
			Page:  1,
			Limit: 50,
		},
	}
	c := &controller{service: service}

	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(5))
		ctx.Set("role", "user")
		ctx.Set("team_id", int64(11))
		ctx.Next()
	})
	r.GET("/v1/tasks", c.List)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks?page=0&limit=200&teamId=11&jobTypeId=22&feederId=7&workDate=2026-04-28", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if service.capturedActor.Role != "user" || service.capturedActor.TeamID == nil || *service.capturedActor.TeamID != 11 {
		t.Fatalf("captured actor = %#v, want user team 11", service.capturedActor)
	}
	if service.capturedList.Page != 0 || service.capturedList.Limit != 200 {
		t.Fatalf("captured pagination = (%d,%d), want (0,200)", service.capturedList.Page, service.capturedList.Limit)
	}
	if service.capturedList.Filter.TeamID == nil || *service.capturedList.Filter.TeamID != 11 {
		t.Fatalf("captured TeamID = %#v, want 11", service.capturedList.Filter.TeamID)
	}
	if service.capturedList.Filter.JobTypeID == nil || *service.capturedList.Filter.JobTypeID != 22 {
		t.Fatalf("captured JobTypeID = %#v, want 22", service.capturedList.Filter.JobTypeID)
	}
	if service.capturedList.Filter.FeederID == nil || *service.capturedList.Filter.FeederID != 7 {
		t.Fatalf("captured FeederID = %#v, want 7", service.capturedList.Filter.FeederID)
	}

	var resp struct {
		Success bool                   `json:"success"`
		Data    []taskdto.TaskResponse `json:"data"`
		Meta    *taskdto.Meta          `json:"meta"`
		Error   *taskdto.ErrorInfo     `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.Success {
		t.Fatal("Success = false, want true")
	}
	if resp.Error != nil {
		t.Fatalf("Error = %#v, want nil", resp.Error)
	}
	if resp.Meta == nil || resp.Meta.Page != 1 || resp.Meta.Limit != 50 || resp.Meta.Total != 1 {
		t.Fatalf("Meta = %#v, want page=1 limit=50 total=1", resp.Meta)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("data length = %d, want 1", len(resp.Data))
	}
	got := resp.Data[0]
	if got.Team == nil || got.Team.Name != "Team Alpha" {
		t.Fatalf("team = %#v, want Team Alpha", got.Team)
	}
	if got.Feeder == nil || got.Feeder.Station == nil || got.Feeder.Station.OperationCenter == nil {
		t.Fatalf("feeder graph = %#v, want nested station and operation center", got.Feeder)
	}
	if got.Feeder.Station.OperationCenter.Name != "Center A" {
		t.Fatalf("operation center = %#v, want Center A", got.Feeder.Station.OperationCenter)
	}
	if got.SourceType == nil || *got.SourceType != sourceType || got.SourceID == nil || *got.SourceID != sourceID || got.LargeWorkTaskID == nil || *got.LargeWorkTaskID != largeWorkTaskID {
		t.Fatalf("source fields = type:%#v source:%#v largeTask:%#v", got.SourceType, got.SourceID, got.LargeWorkTaskID)
	}
}

func TestListMapsServiceErrorToStandardError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c := &controller{service: &fakeService{listErr: errors.New("db failure")}}
	r := gin.New()
	r.Use(func(ctx *gin.Context) { ctx.Set("user_id", uint(1)); ctx.Set("role", "admin"); ctx.Next() })
	r.GET("/v1/tasks", c.List)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp struct {
		Success bool               `json:"success"`
		Data    json.RawMessage    `json:"data"`
		Meta    *taskdto.Meta      `json:"meta"`
		Error   *taskdto.ErrorInfo `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Success {
		t.Fatal("Success = true, want false")
	}
	if resp.Error == nil || resp.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("Error = %#v, want INTERNAL_ERROR", resp.Error)
	}
	if resp.Error.Message != "db failure" {
		t.Fatalf("Error.Message = %q, want db failure", resp.Error.Message)
	}
}

func TestCreateMapsForbiddenToForbiddenResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeService{createErr: taskservice.ErrForbidden}
	c := &controller{service: service}
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(5))
		ctx.Set("role", "user")
		ctx.Set("team_id", int64(11))
		ctx.Next()
	})
	r.POST("/v1/tasks", c.Create)

	req := httptest.NewRequest(http.MethodPost, "/v1/tasks", strings.NewReader(`{"workDate":"2026-04-28","teamId":99,"jobTypeId":1,"jobDetailId":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestListByFilterPassesTeamIDQueryToService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeService{listByFilterOutput: &taskservice.ListTasksByFilterOutput{}}
	c := &controller{service: service}
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(1))
		ctx.Set("role", "super_admin")
		ctx.Next()
	})
	r.GET("/v1/tasks/by-filter", c.ListByFilter)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/by-filter?year=2026&month=5&teamId=42", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want %d", rec.Code, rec.Body.String(), http.StatusOK)
	}
	if service.capturedByFilter.Year != "2026" || service.capturedByFilter.Month != "5" {
		t.Fatalf("captured year/month = (%q,%q), want (2026,5)", service.capturedByFilter.Year, service.capturedByFilter.Month)
	}
	if service.capturedByFilter.TeamID == nil || *service.capturedByFilter.TeamID != 42 {
		t.Fatalf("captured TeamID = %#v, want 42", service.capturedByFilter.TeamID)
	}
}

func TestListRequiresAuthenticatedActor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c := &controller{service: &fakeService{}}
	r := gin.New()
	r.GET("/v1/tasks", c.List)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func stringPtr(v string) *string { return &v }

var _ taskrepository.Repository
