package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakePlanningService struct {
	called bool
	actor  Actor
	year   int
	month  int
	resp   *MonthProjection
	err    error
}

func (f *fakePlanningService) GetMonth(ctx context.Context, actor Actor, year, month int) (*MonthProjection, error) {
	f.called = true
	f.actor = actor
	f.year = year
	f.month = month
	return f.resp, f.err
}

func TestGetMonthReturnsCalendarProjection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm := int64(77)
	svc := &fakePlanningService{resp: &MonthProjection{
		Year:  2026,
		Month: 5,
		Days:  []CalendarDay{{Date: "2026-05-03", Items: []CalendarItem{{SourceType: "team_plan", SourceID: 1}}}},
	}}
	ctrl := NewController(svc)
	_ = gin.New()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/planning/calendar/2026/5", nil)
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "year", Value: "2026"}, {Key: "month", Value: "5"}}
	ctx.Set("user_id", uint(10))
	ctx.Set("role", "user")
	ctx.Set("team_id", tm)

	ctrl.GetMonth(ctx)

	if !svc.called {
		t.Fatal("expected service to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var out struct {
		Success bool            `json:"success"`
		Data    MonthProjection `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !out.Success || out.Data.Year != 2026 || len(out.Data.Days) != 1 {
		t.Fatalf("unexpected response: %#v", out)
	}
}
