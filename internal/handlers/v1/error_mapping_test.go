package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	taskusecase "backend-hotlines3/internal/app/task/usecase"
	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/dto"
	"github.com/gin-gonic/gin"
)

type fakeListTasksUseCase struct {
	captured taskusecase.ListTasksInput
	output   *taskusecase.ListTasksOutput
	err      error
}

func (f *fakeListTasksUseCase) Execute(ctx context.Context, input taskusecase.ListTasksInput) (*taskusecase.ListTasksOutput, error) {
	f.captured = input
	return f.output, f.err
}

func TestTaskHandlerListReturnsStandardEnvelopeAndNestedTaskData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workDate := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 28, 12, 30, 0, 0, time.UTC)
	feederID := int64(7)
	operationCenterID := int64(55)
	task := taskdomain.Entity{
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
	usecase := &fakeListTasksUseCase{
		output: &taskusecase.ListTasksOutput{
			Tasks: []taskdomain.Entity{task},
			Total: 1,
			Page:  1,
			Limit: 50,
		},
	}
	h := &TaskHandler{listUC: usecase}

	r := gin.New()
	r.GET("/v1/tasks", h.List)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks?page=0&limit=200&teamId=11&jobTypeId=22&feederId=7&workDate=2026-04-28", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	// Pagination normalization is tested at the usecase level (list_tasks_test.go).
	// The handler simply passes through query params to the usecase.
	if usecase.captured.Page != 0 || usecase.captured.Limit != 200 {
		t.Fatalf("captured pagination = (%d,%d), want (0,200) — normalization is usecase's job", usecase.captured.Page, usecase.captured.Limit)
	}
	if usecase.captured.Filter.TeamID == nil || *usecase.captured.Filter.TeamID != 11 {
		t.Fatalf("captured TeamID = %#v, want 11", usecase.captured.Filter.TeamID)
	}

	var resp struct {
		Success bool               `json:"success"`
		Data    []dto.TaskResponse `json:"data"`
		Meta    *dto.Meta          `json:"meta"`
		Error   *dto.ErrorInfo     `json:"error"`
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

func TestTaskHandlerListMapsUseCaseErrorToStandardError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &TaskHandler{listUC: &fakeListTasksUseCase{err: errors.New("db failure")}}
	r := gin.New()
	r.GET("/v1/tasks", h.List)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
		Meta    *dto.Meta       `json:"meta"`
		Error   *dto.ErrorInfo  `json:"error"`
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
