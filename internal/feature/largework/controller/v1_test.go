package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/largework/entity"
	"backend-hotlines3/internal/feature/largework/service"

	"github.com/gin-gonic/gin"
)

type fakeService struct {
	createResult *entity.LargeWorkItem
	createErr    error
	getResult    *entity.LargeWorkItem
	getErr       error
	listResult   *service.ListOutput
	listErr      error
	updateResult *entity.LargeWorkItem
	updateErr    error
	cancelResult *entity.LargeWorkItem
	cancelErr    error
	lastCreate   service.CreateInput
	lastUpdate   service.UpdateInput
	lastList     service.ListInput
	createCalls  int
	updateCalls  int
}

func (f *fakeService) Create(_ context.Context, _ entity.Actor, input service.CreateInput) (*entity.LargeWorkItem, error) {
	f.createCalls++
	f.lastCreate = input
	return f.createResult, f.createErr
}
func (f *fakeService) GetByID(_ context.Context, _ entity.Actor, id int64) (*entity.LargeWorkItem, error) {
	if f.getResult != nil && f.getResult.ID != id {
		return nil, nil
	}
	return f.getResult, f.getErr
}
func (f *fakeService) List(_ context.Context, _ entity.Actor, input service.ListInput) (*service.ListOutput, error) {
	f.lastList = input
	return f.listResult, f.listErr
}
func (f *fakeService) Update(_ context.Context, _ entity.Actor, input service.UpdateInput) (*entity.LargeWorkItem, error) {
	f.updateCalls++
	f.lastUpdate = input
	return f.updateResult, f.updateErr
}
func (f *fakeService) Cancel(_ context.Context, _ entity.Actor, _ int64) (*entity.LargeWorkItem, error) {
	return f.cancelResult, f.cancelErr
}

func TestCreateReturnsStandardEnvelopeAndParsesTeams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	teamLeadID := int64(7)
	svc := &fakeService{createResult: &entity.LargeWorkItem{ID: 501, OwnerTeamID: 7, Status: entity.LargeWorkStatusPlanned}}
	c := NewController(svc)

	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(1))
		ctx.Set("role", policy.RoleTeamLead)
		ctx.Set("team_id", teamLeadID)
		ctx.Next()
	})
	r.POST("/v1/large-work-items", c.Create)

	req := httptest.NewRequest(http.MethodPost, "/v1/large-work-items", strings.NewReader(`{"ownerTeamId":7,"participantTeamIds":[8,7],"title":"งานระดมทีม","startDate":"2026-06-10","locationText":"Station A"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			ID          int64 `json:"id"`
			OwnerTeamID int64 `json:"ownerTeamId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.Success || body.Data.ID != 501 || body.Data.OwnerTeamID != 7 {
		t.Fatalf("unexpected response: %+v", body)
	}
	if len(svc.lastCreate.ParticipantTeamIDs) != 2 {
		t.Fatalf("captured participants = %#v", svc.lastCreate.ParticipantTeamIDs)
	}
}

func TestListAndGetMapForbiddenAndUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeService{listErr: service.ErrForbidden, getErr: service.ErrForbidden}
	c := NewController(svc)

	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(1))
		ctx.Set("role", policy.RoleUser)
		ctx.Set("team_id", int64(7))
		ctx.Next()
	})
	r.GET("/v1/large-work-items", c.List)
	r.GET("/v1/large-work-items/:id", c.GetByID)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/large-work-items?from=2026-06-01&to=2026-06-30&page=0&limit=200", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("list status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/large-work-items/10", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("get status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestUpdateRejectsBadDateAndDeleteOrCancelUsesStandardResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	teamLeadID := int64(7)
	svc := &fakeService{cancelResult: &entity.LargeWorkItem{ID: 10, OwnerTeamID: 7, Status: entity.LargeWorkStatusCancelled}}
	c := NewController(svc)

	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(1))
		ctx.Set("role", policy.RoleTeamLead)
		ctx.Set("team_id", teamLeadID)
		ctx.Next()
	})
	r.PATCH("/v1/large-work-items/:id", c.Update)
	r.POST("/v1/large-work-items/:id/cancel", c.Cancel)

	req := httptest.NewRequest(http.MethodPatch, "/v1/large-work-items/10", strings.NewReader(`{"startDate":"bad-date"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("update status = %d, want 400 body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/large-work-items/10/cancel", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel status = %d, want 200 body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "cancelled") {
		t.Fatalf("cancel body = %s, want cancelled status", rec.Body.String())
	}
}

func TestCreateAndUpdateRejectEndDateBeforeStartDate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	teamLeadID := int64(7)
	svc := &fakeService{createResult: &entity.LargeWorkItem{ID: 1, OwnerTeamID: 7, Status: entity.LargeWorkStatusPlanned}, updateResult: &entity.LargeWorkItem{ID: 10, OwnerTeamID: 7, Status: entity.LargeWorkStatusPlanned}}
	c := NewController(svc)

	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(1))
		ctx.Set("role", policy.RoleAdmin)
		ctx.Set("team_id", teamLeadID)
		ctx.Next()
	})
	r.POST("/v1/large-work-items", c.Create)
	r.PATCH("/v1/large-work-items/:id", c.Update)

	req := httptest.NewRequest(http.MethodPost, "/v1/large-work-items", strings.NewReader(`{"ownerTeamId":7,"participantTeamIds":[7,8],"title":"งานระดมทีม","startDate":"2026-06-10","endDate":"2026-06-09","locationText":"Station A"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create status = %d, want 400 body=%s", rec.Code, rec.Body.String())
	}
	if svc.createCalls != 0 {
		t.Fatalf("create service calls = %d, want 0", svc.createCalls)
	}

	req = httptest.NewRequest(http.MethodPatch, "/v1/large-work-items/10", strings.NewReader(`{"startDate":"2026-06-10","endDate":"2026-06-09"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("update status = %d, want 400 body=%s", rec.Code, rec.Body.String())
	}
	if svc.updateCalls != 0 {
		t.Fatalf("update service calls = %d, want 0", svc.updateCalls)
	}
}

func TestUpdateAndCancelMapStateRestrictionsToBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeService{updateErr: service.ErrInvalidStateTransition, cancelErr: service.ErrInvalidStateTransition}
	c := NewController(svc)

	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(1))
		ctx.Set("role", policy.RoleAdmin)
		ctx.Next()
	})
	r.PATCH("/v1/large-work-items/:id", c.Update)
	r.POST("/v1/large-work-items/:id/cancel", c.Cancel)

	req := httptest.NewRequest(http.MethodPatch, "/v1/large-work-items/10", strings.NewReader(`{"title":"updated"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("update status = %d, want 400 body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/large-work-items/10/cancel", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("cancel status = %d, want 400 body=%s", rec.Code, rec.Body.String())
	}
}
