package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	service "backend-hotlines3/internal/feature/dailyreportdraft/service"
	taskentity "backend-hotlines3/internal/feature/task/entity"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	capturedListInput service.ListSourcesInput
	capturedFromInput service.FromPlanInput
	capturedActor     taskentity.Actor
	listResult        *service.ListSourcesOutput
	fromResult        *service.FromPlanOutput
	listErr           error
	fromErr           error
}

func (f *fakeService) ListSources(_ context.Context, actor taskentity.Actor, input service.ListSourcesInput) (*service.ListSourcesOutput, error) {
	f.capturedActor = actor
	f.capturedListInput = input
	return f.listResult, f.listErr
}

func (f *fakeService) FromPlan(_ context.Context, actor taskentity.Actor, input service.FromPlanInput) (*service.FromPlanOutput, error) {
	f.capturedActor = actor
	f.capturedFromInput = input
	return f.fromResult, f.fromErr
}

func TestListSourcesReturnsEnvelopeAndScopesTeam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	teamID := int64(7)
	selected := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	fake := &fakeService{listResult: &service.ListSourcesOutput{Items: []service.SourceCandidate{{SourceRef: service.SourceRef{SourceType: service.SourceTypeTeamPlan, SourceID: 101, Title: "Patrol feeder A"}}}}}
	c := NewController(fake)
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(5))
		ctx.Set("role", "user")
		ctx.Set("team_id", teamID)
		ctx.Next()
	})
	r.GET("/v1/daily-report-drafts/sources", c.ListSources)

	req := httptest.NewRequest(http.MethodGet, "/v1/daily-report-drafts/sources?workDate=2026-06-04", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if fake.capturedListInput.WorkDate == nil || fake.capturedListInput.WorkDate.Format("2006-01-02") != selected.Format("2006-01-02") {
		t.Fatalf("captured workDate = %#v", fake.capturedListInput.WorkDate)
	}
	if fake.capturedListInput.TeamID == nil || *fake.capturedListInput.TeamID != teamID {
		t.Fatalf("captured teamID = %#v", fake.capturedListInput.TeamID)
	}
	var resp struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected success response")
	}
	var out service.ListSourcesOutput
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].SourceID != 101 {
		t.Fatalf("items = %#v", out.Items)
	}
}

func TestFromPlanMapsValidationAndForbiddenErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name       string
		query      string
		fromErr    error
		wantStatus int
	}{
		{name: "bad query", query: "/v1/daily-report-drafts/from-plan?sourceType=team_plan&sourceId=oops", wantStatus: http.StatusBadRequest},
		{name: "forbidden", query: "/v1/daily-report-drafts/from-plan?sourceType=team_plan&sourceId=101", fromErr: service.ErrForbidden, wantStatus: http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeService{fromErr: tc.fromErr}
			c := NewController(fake)
			r := gin.New()
			r.Use(func(ctx *gin.Context) {
				ctx.Set("user_id", uint(5))
				ctx.Set("role", "user")
				ctx.Set("team_id", int64(7))
				ctx.Next()
			})
			r.GET("/v1/daily-report-drafts/from-plan", c.FromPlan)
			req := httptest.NewRequest(http.MethodGet, tc.query, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

func TestFromPlanReturnsPrefillEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	teamID := int64(7)
	selected := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	fake := &fakeService{fromResult: &service.FromPlanOutput{Source: service.SourceRef{SourceType: service.SourceTypeTeamPlan, SourceID: 101, Title: "Patrol feeder A"}, Prefill: service.Prefill{WorkDate: "2026-06-04", TeamID: &teamID}, Warnings: []string{"jobTypeId and jobDetailId must be selected before saving"}}}
	c := NewController(fake)
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(5))
		ctx.Set("role", "user")
		ctx.Set("team_id", teamID)
		ctx.Next()
	})
	r.GET("/v1/daily-report-drafts/from-plan", c.FromPlan)

	req := httptest.NewRequest(http.MethodGet, "/v1/daily-report-drafts/from-plan?sourceType=team_plan&sourceId=101&workDate=2026-06-04", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if fake.capturedFromInput.SourceType != service.SourceTypeTeamPlan || fake.capturedFromInput.SourceID != 101 {
		t.Fatalf("captured input = %#v", fake.capturedFromInput)
	}
	if fake.capturedFromInput.WorkDate == nil || fake.capturedFromInput.WorkDate.Format("2006-01-02") != selected.Format("2006-01-02") {
		t.Fatalf("captured workDate = %#v", fake.capturedFromInput.WorkDate)
	}
	var resp struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.Success {
		t.Fatal("expected success response")
	}
	var out service.FromPlanOutput
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if out.Source.SourceID != 101 || out.Source.SourceType != service.SourceTypeTeamPlan {
		t.Fatalf("data = %#v", out.Source)
	}
}
