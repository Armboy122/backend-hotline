package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-hotlines3/internal/feature/auth/policy"
	taskentity "backend-hotlines3/internal/feature/task/entity"
	service "backend-hotlines3/internal/feature/workreport/service"

	"github.com/gin-gonic/gin"
)

type fakeWorkReportService struct {
	capturedActor taskentity.Actor
	capturedInput service.GetReportInput
	result        *service.GetReportOutput
	err           error
}

func (f *fakeWorkReportService) GetReport(_ context.Context, actor taskentity.Actor, input service.GetReportInput) (*service.GetReportOutput, error) {
	f.capturedActor = actor
	f.capturedInput = input
	return f.result, f.err
}

func workReportTestRouter(ctrl *Controller, role string, teamID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(5))
		ctx.Set("role", role)
		ctx.Set("team_id", teamID)
		ctx.Next()
	})
	r.GET("/v1/work-reports", ctrl.GetReport)
	return r
}

func TestGetReportPassesYearMonthTeamAndActorToService(t *testing.T) {
	requestedTeamID := int64(9)
	fake := &fakeWorkReportService{result: &service.GetReportOutput{Summary: service.Summary{TotalTasks: 1}, Items: []service.ReportItem{{ID: 101, WorkDate: "2026-05-03", TeamID: requestedTeamID}}}}
	ctrl := NewController(fake)
	r := workReportTestRouter(ctrl, policy.RoleAdmin, 7)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/work-reports?year=2026&month=5&teamId=9", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}
	if fake.capturedActor.Role != policy.RoleAdmin || fake.capturedActor.TeamID == nil || *fake.capturedActor.TeamID != 7 {
		t.Fatalf("actor = %#v", fake.capturedActor)
	}
	if fake.capturedInput.Year != "2026" || fake.capturedInput.Month != "5" {
		t.Fatalf("captured input = %#v", fake.capturedInput)
	}
	if fake.capturedInput.TeamID == nil || *fake.capturedInput.TeamID != requestedTeamID {
		t.Fatalf("captured teamID = %#v, want %d", fake.capturedInput.TeamID, requestedTeamID)
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
	var out service.GetReportOutput
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if out.Summary.TotalTasks != 1 || len(out.Items) != 1 || out.Items[0].ID != 101 {
		t.Fatalf("data = %#v", out)
	}
}

func TestGetReportAcceptsMonthYYYYMMQuery(t *testing.T) {
	fake := &fakeWorkReportService{result: &service.GetReportOutput{}}
	ctrl := NewController(fake)
	r := workReportTestRouter(ctrl, policy.RoleViewer, 7)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/work-reports?month=2026-05", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}
	if fake.capturedInput.Year != "2026" || fake.capturedInput.Month != "5" {
		t.Fatalf("captured input = %#v, want year 2026 month 5", fake.capturedInput)
	}
}

func TestGetReportRejectsInvalidQueryAndMapsServiceErrors(t *testing.T) {
	for _, tc := range []struct {
		name       string
		query      string
		err        error
		wantStatus int
	}{
		{name: "missing year", query: "/v1/work-reports?month=5", wantStatus: http.StatusBadRequest},
		{name: "invalid month", query: "/v1/work-reports?year=2026&month=13", wantStatus: http.StatusBadRequest},
		{name: "invalid team", query: "/v1/work-reports?year=2026&month=5&teamId=x", wantStatus: http.StatusBadRequest},
		{name: "forbidden", query: "/v1/work-reports?year=2026&month=5", err: service.ErrForbidden, wantStatus: http.StatusForbidden},
		{name: "service validation", query: "/v1/work-reports?year=2026&month=5", err: service.ErrYearMonthRequired, wantStatus: http.StatusBadRequest},
		{name: "internal", query: "/v1/work-reports?year=2026&month=5", err: errors.New("db failure"), wantStatus: http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeWorkReportService{result: &service.GetReportOutput{}, err: tc.err}
			ctrl := NewController(fake)
			r := workReportTestRouter(ctrl, policy.RoleUser, 7)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.query, nil)
			r.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d body=%s, want %d", rec.Code, rec.Body.String(), tc.wantStatus)
			}
		})
	}
}

func TestGetReportRequiresAuthenticatedActorContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewController(&fakeWorkReportService{})
	r := gin.New()
	r.GET("/v1/work-reports", ctrl.GetReport)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/work-reports?year=2026&month=5", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s, want 401", rec.Code, rec.Body.String())
	}
}
