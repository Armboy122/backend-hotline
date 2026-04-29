package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	taskdto "backend-hotlines3/internal/feature/task/dto"
	"backend-hotlines3/internal/feature/task/entity"
	taskrepository "backend-hotlines3/internal/feature/task/repository"
	taskservice "backend-hotlines3/internal/feature/task/service"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	capturedList taskservice.ListTasksInput
	listOutput   *taskservice.ListTasksOutput
	listErr      error
}

func (f *fakeService) List(_ context.Context, input taskservice.ListTasksInput) (*taskservice.ListTasksOutput, error) {
	f.capturedList = input
	return f.listOutput, f.listErr
}
func (f *fakeService) GetByID(context.Context, int64) (*entity.Task, error) { return nil, nil }
func (f *fakeService) Create(context.Context, taskservice.CreateTaskInput) (*entity.Task, error) {
	return nil, nil
}
func (f *fakeService) Update(context.Context, taskservice.UpdateTaskInput) (*entity.Task, error) {
	return nil, nil
}
func (f *fakeService) Delete(context.Context, int64) error { return nil }
func (f *fakeService) ListByTeam(context.Context, taskservice.ListTasksByTeamInput) (*taskservice.ListTasksByTeamOutput, error) {
	return nil, nil
}
func (f *fakeService) ListByFilter(context.Context, taskservice.ListTasksByFilterInput) (*taskservice.ListTasksByFilterOutput, error) {
	return nil, nil
}

func TestListReturnsStandardEnvelopeAndNestedTaskData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workDate := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 28, 12, 30, 0, 0, time.UTC)
	feederID := int64(7)
	operationCenterID := int64(55)
	task := entity.Task{
		ID:                  101,
		WorkDate:            workDate,
		TeamID:              11,
		JobTypeID:           22,
		JobDetailID:         33,
		FeederID:            &feederID,
		URLsBefore:          []string{"before-1"},
		URLsAfter:           []string{"after-1"},
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
			Tasks: []entity.Task{task},
			Total: 1,
			Page:  1,
			Limit: 50,
		},
	}
	c := &controller{service: service}

	r := gin.New()
	r.GET("/v1/tasks", c.List)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks?page=0&limit=200&teamId=11&jobTypeId=22&feederId=7&workDate=2026-04-28", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
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
}

func TestListMapsServiceErrorToStandardError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c := &controller{service: &fakeService{listErr: errors.New("db failure")}}
	r := gin.New()
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

func stringPtr(v string) *string { return &v }

var _ taskrepository.Repository
