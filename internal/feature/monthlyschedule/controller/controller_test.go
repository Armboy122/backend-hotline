package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend-hotlines3/internal/feature/monthlyschedule/entity"
	"backend-hotlines3/internal/feature/monthlyschedule/service"
	"backend-hotlines3/internal/models"

	"github.com/gin-gonic/gin"
)

func TestClinicToolIntegrationRequiresKeyAndReturnsContractWithETag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &controllerFakeRepository{published: publishedSchedule()}
	controller := New(service.New(repo), "integration-secret")
	router := gin.New()
	router.GET("/v1/integrations/clinic-tool/monthly-plans/:year/:month", controller.GetPublishedForClinicTool)

	t.Run("rejects invalid key", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/v1/integrations/clinic-tool/monthly-plans/2026/7", nil)
		request.Header.Set("X-Integration-Key", "wrong")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401; body=%s", response.Code, response.Body.String())
		}
		if response.Header().Get("X-Request-ID") == "" {
			t.Fatal("unauthorized response must include X-Request-ID")
		}
	})

	t.Run("returns stable published payload", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/v1/integrations/clinic-tool/monthly-plans/2026/7", nil)
		request.Header.Set("X-Integration-Key", "integration-secret")
		request.Header.Set("X-Request-ID", "clinic-request-1")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", response.Code, response.Body.String())
		}
		wantETag := `"` + strings.Repeat("a", 64) + `"`
		if response.Header().Get("ETag") != wantETag {
			t.Fatalf("ETag = %q, want %q", response.Header().Get("ETag"), wantETag)
		}
		if response.Header().Get("X-Request-ID") != "clinic-request-1" {
			t.Fatalf("request ID = %q", response.Header().Get("X-Request-ID"))
		}
		var body struct {
			Success bool `json:"success"`
			Data    struct {
				Period struct {
					Year  int `json:"year"`
					Month int `json:"month"`
				} `json:"period"`
				Revision struct {
					No int `json:"no"`
				} `json:"revision"`
				Teams []struct {
					Code     string `json:"code"`
					Segments []struct {
						AssignmentType string `json:"assignmentType"`
						StartDate      string `json:"startDate"`
						EndDate        string `json:"endDate"`
					} `json:"segments"`
				} `json:"teams"`
			} `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if !body.Success || body.Data.Period.Year != 2026 || body.Data.Period.Month != 7 || body.Data.Revision.No != 2 {
			t.Fatalf("unexpected contract header: %+v", body)
		}
		if len(body.Data.Teams) != 1 || body.Data.Teams[0].Code != "T01" {
			t.Fatalf("unexpected team contract: %+v", body.Data.Teams)
		}
		segments := body.Data.Teams[0].Segments
		if len(segments) != 1 || segments[0].AssignmentType != "home" || segments[0].StartDate != "2026-07-01" || segments[0].EndDate != "2026-07-31" {
			t.Fatalf("unexpected projected segments: %+v", segments)
		}
	})

	t.Run("honors If-None-Match", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/v1/integrations/clinic-tool/monthly-plans/2026/7", nil)
		request.Header.Set("X-Integration-Key", "integration-secret")
		request.Header.Set("If-None-Match", `"`+strings.Repeat("a", 64)+`"`)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusNotModified {
			t.Fatalf("status = %d, want 304; body=%s", response.Code, response.Body.String())
		}
		if response.Body.Len() != 0 {
			t.Fatalf("304 body = %q, want empty", response.Body.String())
		}
	})
}

func TestClinicToolIntegrationIsUnavailableWithoutConfiguredSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := New(service.New(&controllerFakeRepository{published: publishedSchedule()}), "")
	router := gin.New()
	router.GET("/v1/integrations/clinic-tool/monthly-plans/:year/:month", controller.GetPublishedForClinicTool)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/integrations/clinic-tool/monthly-plans/2026/7", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body=%s", response.Code, response.Body.String())
	}
}

func TestWorkspaceHidesDraftFromNonSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &controllerFakeRepository{
		draft:     publishedSchedule(),
		published: publishedSchedule(),
	}
	controller := New(service.New(repo), "")
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(42))
		ctx.Set("role", "admin")
		ctx.Next()
	})
	router.GET("/v1/monthly-plans/:year/:month/schedule", controller.GetWorkspace)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/monthly-plans/2026/7/schedule", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Draft     json.RawMessage `json:"draft"`
			Published json.RawMessage `json:"published"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || string(body.Data.Draft) != "null" {
		t.Fatalf("non-super_admin draft = %s, want null", body.Data.Draft)
	}
	if len(body.Data.Published) == 0 || string(body.Data.Published) == "null" {
		t.Fatal("published schedule must remain visible")
	}
}

type controllerFakeRepository struct {
	draft     *entity.Schedule
	published *entity.Schedule
}

func (f *controllerFakeRepository) FindOrCreatePeriod(context.Context, int, int) (entity.Period, error) {
	return entity.Period{ID: 1, Year: 2026, Month: 7}, nil
}

func (f *controllerFakeRepository) FindPeriod(context.Context, int, int) (entity.Period, error) {
	return entity.Period{ID: 1, Year: 2026, Month: 7}, nil
}

func (f *controllerFakeRepository) ListVisibleTeams(context.Context) ([]entity.Team, error) {
	code, base, crew := "T01", "ขอนแก่น", "ฮอทไลน์"
	return []entity.Team{{
		ID:                 1,
		Name:               "ชุด 1",
		Code:               &code,
		BaseArea:           &base,
		CrewType:           &crew,
		DisplayOrder:       1,
		MonthlyPlanVisible: true,
	}}, nil
}

func (f *controllerFakeRepository) GetSchedule(_ context.Context, _ int64, status string) (*entity.Schedule, error) {
	if status == models.MonthlyPlanScheduleStatusDraft && f.draft != nil {
		return f.draft, nil
	}
	if status == models.MonthlyPlanScheduleStatusPublished && f.published != nil {
		return f.published, nil
	}
	return nil, entity.ErrPublishedNotFound
}

func (f *controllerFakeRepository) ReplaceDraft(context.Context, entity.Period, entity.Actor, *int, []entity.Assignment) (*entity.Schedule, error) {
	panic("unexpected ReplaceDraft")
}

func (f *controllerFakeRepository) PublishDraft(context.Context, entity.Period, entity.Actor, int64, string, []byte) (*entity.Schedule, error) {
	panic("unexpected PublishDraft")
}

func publishedSchedule() *entity.Schedule {
	publishedAt := time.Date(2026, 6, 25, 10, 0, 0, 0, time.UTC)
	checksum := strings.Repeat("a", 64)
	return &entity.Schedule{
		Revision: &entity.Revision{
			ID:          20,
			RevisionNo:  2,
			Status:      models.MonthlyPlanScheduleStatusPublished,
			PublishedAt: &publishedAt,
			Checksum:    &checksum,
		},
		Assignments: []entity.Assignment{},
	}
}
