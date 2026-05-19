package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/teamplan/entity"
	"backend-hotlines3/internal/feature/teamplan/service"

	"github.com/gin-gonic/gin"
)

func TestListRequiresAuthAndReturnsPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &fakeTeamPlanService{listResult: &service.ListOutput{
		Items: []entity.TeamPlan{{ID: 10, TeamID: 7, CreatedByUserID: 91, Title: "Patrol", LocationText: "Bang Khen", Status: entity.StatusPlanned}},
		Total: 1,
		Page:  2,
		Limit: 25,
	}}
	ctrl := NewController(fake)

	r := gin.New()
	r.GET("/team-plans", ctrl.List)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/team-plans?from=2026-06-01&to=2026-06-30&page=2&limit=25&status=planned", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want 401 body=%s", rec.Code, rec.Body.String())
	}

	r = gin.New()
	r.Use(withActor(3, policy.RoleUser, int64Ptr(7)))
	r.GET("/team-plans", ctrl.List)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/team-plans?from=2026-06-01&to=2026-06-30&page=2&limit=25&status=planned", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200 body=%s", rec.Code, rec.Body.String())
	}
	if fake.listInput.Page != 2 || fake.listInput.Limit != 25 {
		t.Fatalf("list pagination not forwarded: %+v", fake.listInput)
	}
	if got := fake.listInput.Status; len(got) != 1 || got[0] != "planned" {
		t.Fatalf("list status filter = %+v", got)
	}
	var body struct {
		Success bool `json:"success"`
		Data    []struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
		} `json:"data"`
		Meta struct {
			Page  int   `json:"page"`
			Limit int   `json:"limit"`
			Total int64 `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || body.Meta.Page != 2 || body.Meta.Limit != 25 || body.Meta.Total != 1 || len(body.Data) != 1 {
		t.Fatalf("unexpected list body: %+v", body)
	}
}

func TestCreateUpdateGetAndDeleteMapAuthAndDomainErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	teamID := int64(7)
	creatorID := int64(91)
	plan := entity.TeamPlan{ID: 10, TeamID: teamID, CreatedByUserID: creatorID, Title: "Patrol", LocationText: "Bang Khen", Status: entity.StatusPlanned}
	fake := &fakeTeamPlanService{getResult: &plan, createResult: &plan, updateResult: &plan}
	ctrl := NewController(fake)

	r := gin.New()
	r.Use(withActor(uint(creatorID), policy.RoleUser, &teamID))
	r.GET("/team-plans/:id", ctrl.GetByID)
	r.POST("/team-plans", ctrl.Create)
	r.PUT("/team-plans/:id", ctrl.Update)
	r.DELETE("/team-plans/:id", ctrl.Delete)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/team-plans/10", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200 body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/team-plans", strings.NewReader(`{"teamId":7,"title":"Patrol","startDate":"2026-06-03","locationText":"Bang Khen"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201 body=%s", rec.Code, rec.Body.String())
	}
	if fake.createInput.TeamID != teamID || fake.createActor.UserID != creatorID {
		t.Fatalf("create input not forwarded: actor=%+v input=%+v", fake.createActor, fake.createInput)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/team-plans/10", strings.NewReader(`{"title":"Patrol updated"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want 200 body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/team-plans/10", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204 body=%s", rec.Code, rec.Body.String())
	}

	fake.getErr = entity.ErrNotFound
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/team-plans/999", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("not found status = %d, want 404 body=%s", rec.Code, rec.Body.String())
	}

	fake.createErr = entity.ErrForbidden
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/team-plans", strings.NewReader(`{"teamId":8,"title":"Other team","startDate":"2026-06-03","locationText":"Bang Khen"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("forbidden create status = %d, want 403 body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/team-plans/10", strings.NewReader(`{"status":"bogus"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid status update = %d, want 400 body=%s", rec.Code, rec.Body.String())
	}

	fake.deleteErr = entity.ErrForbidden
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/team-plans/10", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("forbidden delete status = %d, want 403 body=%s", rec.Code, rec.Body.String())
	}
}

type fakeTeamPlanService struct {
	listResult   *service.ListOutput
	listErr      error
	getResult    *entity.TeamPlan
	getErr       error
	createResult *entity.TeamPlan
	createErr    error
	updateResult *entity.TeamPlan
	updateErr    error
	deleteErr    error
	listInput    service.ListInput
	getID        int64
	createActor  entity.Actor
	createInput  service.CreateInput
	updateInput  service.UpdateInput
	deleteID     int64
}

func (f *fakeTeamPlanService) List(_ context.Context, _ entity.Actor, input service.ListInput) (*service.ListOutput, error) {
	f.listInput = input
	return f.listResult, f.listErr
}

func (f *fakeTeamPlanService) GetByID(_ context.Context, _ entity.Actor, id int64) (*entity.TeamPlan, error) {
	f.getID = id
	return f.getResult, f.getErr
}

func (f *fakeTeamPlanService) Create(_ context.Context, actor entity.Actor, input service.CreateInput) (*entity.TeamPlan, error) {
	f.createActor = actor
	f.createInput = input
	return f.createResult, f.createErr
}

func (f *fakeTeamPlanService) Update(_ context.Context, _ entity.Actor, input service.UpdateInput) (*entity.TeamPlan, error) {
	f.updateInput = input
	return f.updateResult, f.updateErr
}

func (f *fakeTeamPlanService) Delete(_ context.Context, _ entity.Actor, id int64) error {
	f.deleteID = id
	return f.deleteErr
}

func withActor(userID uint, role string, teamID *int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		if teamID != nil {
			c.Set("team_id", *teamID)
		}
		c.Next()
	}
}

func int64Ptr(v int64) *int64 { return &v }
