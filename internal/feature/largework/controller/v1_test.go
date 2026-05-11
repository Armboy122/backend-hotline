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
	createResult       *entity.LargeWorkItem
	createErr          error
	getResult          *entity.LargeWorkItem
	getErr             error
	listResult         *service.ListOutput
	listErr            error
	updateResult       *entity.LargeWorkItem
	updateErr          error
	cancelResult       *entity.LargeWorkItem
	cancelErr          error
	overviewResult     *service.OverviewOutput
	overviewErr        error
	replaceTasksResult []entity.LargeWorkTask
	replaceTasksErr    error
	listTasksResult    []entity.LargeWorkTask
	listTasksErr       error
	myTodosResult      *service.MyTodosOutput
	myTodosErr         error
	taskResult         *entity.LargeWorkTask
	taskErr            error
	lastCreate         service.CreateInput
	lastUpdate         service.UpdateInput
	lastList           service.ListInput
	lastReplaceTasks   service.ReplaceTasksInput
	lastComplete       service.CompleteTaskInput
	lastPhotoKind      string
	lastPhotoURL       string
	lastBlockReason    string
	createCalls        int
	updateCalls        int
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
func (f *fakeService) GetOverview(_ context.Context, _ entity.Actor, _ int64) (*service.OverviewOutput, error) {
	return f.overviewResult, f.overviewErr
}
func (f *fakeService) ReplaceTasks(_ context.Context, _ entity.Actor, _ int64, input service.ReplaceTasksInput) ([]entity.LargeWorkTask, error) {
	f.lastReplaceTasks = input
	return f.replaceTasksResult, f.replaceTasksErr
}
func (f *fakeService) ListTasks(_ context.Context, _ entity.Actor, _ int64) ([]entity.LargeWorkTask, error) {
	return f.listTasksResult, f.listTasksErr
}
func (f *fakeService) MyTodos(_ context.Context, _ entity.Actor, _ service.MyTodosInput) (*service.MyTodosOutput, error) {
	return f.myTodosResult, f.myTodosErr
}
func (f *fakeService) StartTask(_ context.Context, _ entity.Actor, _ int64) (*entity.LargeWorkTask, error) {
	return f.taskResult, f.taskErr
}
func (f *fakeService) AttachTaskPhoto(_ context.Context, _ entity.Actor, _ int64, kind string, url string) (*entity.LargeWorkTask, error) {
	f.lastPhotoKind = kind
	f.lastPhotoURL = url
	return f.taskResult, f.taskErr
}
func (f *fakeService) CompleteTask(_ context.Context, _ entity.Actor, _ int64, input service.CompleteTaskInput) (*entity.LargeWorkTask, error) {
	f.lastComplete = input
	return f.taskResult, f.taskErr
}
func (f *fakeService) BlockTask(_ context.Context, _ entity.Actor, _ int64, reason string) (*entity.LargeWorkTask, error) {
	f.lastBlockReason = reason
	return f.taskResult, f.taskErr
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

func TestOverviewEndpointReturnsLowerCamelCaseContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	teamID := int64(99)
	svc := &fakeService{overviewResult: &service.OverviewOutput{
		Plan:     entity.LargeWorkItem{ID: 501, OwnerTeamID: 7, Title: "งานระดมทีม"},
		Progress: service.ProgressSummary{Total: 2, Todo: 1, Done: 1},
		TeamProgress: []service.TeamProgressSummary{
			{AssignedTeamID: 8, Total: 2, Todo: 1, Done: 1},
		},
	}}
	c := NewController(svc)
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(1))
		ctx.Set("role", policy.RoleUser)
		ctx.Set("team_id", teamID)
		ctx.Next()
	})
	r.GET("/v1/large-work-items/:id/overview", c.GetOverview)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/large-work-items/501/overview", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d want=200 body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"Plan"`) || !strings.Contains(rec.Body.String(), `"plan"`) || !strings.Contains(rec.Body.String(), `"teamProgress"`) {
		t.Fatalf("overview body uses wrong JSON keys: %s", rec.Body.String())
	}
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Plan struct {
				ID int64 `json:"id"`
			} `json:"plan"`
			Progress struct {
				Total int `json:"total"`
				Done  int `json:"done"`
			} `json:"progress"`
			TeamProgress []struct {
				AssignedTeamID int64 `json:"assignedTeamId"`
				Total          int   `json:"total"`
			} `json:"teamProgress"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.Success || body.Data.Plan.ID != 501 || body.Data.Progress.Total != 2 || body.Data.Progress.Done != 1 || len(body.Data.TeamProgress) != 1 || body.Data.TeamProgress[0].AssignedTeamID != 8 {
		t.Fatalf("unexpected overview contract: %+v body=%s", body, rec.Body.String())
	}
}

func TestExecutionEndpointsReturnOverviewTasksTodosAndActions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	teamID := int64(8)
	task := entity.LargeWorkTask{ID: 11, LargeWorkItemID: 501, AssignedTeamID: teamID, Status: entity.LargeWorkTaskStatusTodo, PointLabel: "P1", WorkType: "tree"}
	svc := &fakeService{
		overviewResult:     &service.OverviewOutput{Plan: entity.LargeWorkItem{ID: 501, OwnerTeamID: 7}, Progress: service.ProgressSummary{Total: 1, Todo: 1}},
		replaceTasksResult: []entity.LargeWorkTask{task},
		listTasksResult:    []entity.LargeWorkTask{task},
		myTodosResult:      &service.MyTodosOutput{Items: []entity.LargeWorkTask{task}, Page: 1, Limit: 50, Total: 1},
		taskResult:         &task,
	}
	c := NewController(svc)
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(1))
		ctx.Set("role", policy.RoleUser)
		ctx.Set("team_id", teamID)
		ctx.Next()
	})
	r.GET("/v1/large-work-items/:id/overview", c.GetOverview)
	r.POST("/v1/large-work-items/:id/tasks", c.ReplaceTasks)
	r.GET("/v1/large-work-items/:id/tasks", c.ListTasks)
	r.GET("/v1/large-work-tasks/my-todos", c.MyTodos)
	r.PATCH("/v1/large-work-tasks/:taskId/start", c.StartTask)
	r.POST("/v1/large-work-tasks/:taskId/photos", c.AttachTaskPhoto)
	r.PATCH("/v1/large-work-tasks/:taskId/complete", c.CompleteTask)
	r.PATCH("/v1/large-work-tasks/:taskId/block", c.BlockTask)

	cases := []struct {
		method, path, body string
		want               int
	}{
		{http.MethodGet, "/v1/large-work-items/501/overview", "", http.StatusOK},
		{http.MethodPost, "/v1/large-work-items/501/tasks", `{"tasks":[{"assignedTeamId":8,"sequence":1,"pointLabel":"P1","workType":"tree"}]}`, http.StatusOK},
		{http.MethodGet, "/v1/large-work-items/501/tasks", "", http.StatusOK},
		{http.MethodGet, "/v1/large-work-tasks/my-todos?page=0&limit=200", "", http.StatusOK},
		{http.MethodPatch, "/v1/large-work-tasks/11/start", "", http.StatusOK},
		{http.MethodPost, "/v1/large-work-tasks/11/photos", `{"kind":"before","url":"https://cdn.example/before.jpg"}`, http.StatusOK},
		{http.MethodPatch, "/v1/large-work-tasks/11/complete", `{"afterPhotoUrls":["https://cdn.example/after.jpg"],"completionNote":"done"}`, http.StatusOK},
		{http.MethodPatch, "/v1/large-work-tasks/11/block", `{"reason":"access denied"}`, http.StatusOK},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Fatalf("%s %s status=%d want=%d body=%s", tc.method, tc.path, rec.Code, tc.want, rec.Body.String())
		}
	}
	if len(svc.lastReplaceTasks.Tasks) != 1 || svc.lastReplaceTasks.Tasks[0].AssignedTeamID != teamID {
		t.Fatalf("replace tasks input = %#v", svc.lastReplaceTasks)
	}
	if svc.lastPhotoKind != service.PhotoKindBefore || svc.lastPhotoURL == "" {
		t.Fatalf("photo capture kind=%q url=%q", svc.lastPhotoKind, svc.lastPhotoURL)
	}
	if svc.lastComplete.CompletionNote != "done" || len(svc.lastComplete.AfterPhotoURLs) != 1 {
		t.Fatalf("complete input = %#v", svc.lastComplete)
	}
	if svc.lastBlockReason != "access denied" {
		t.Fatalf("block reason = %q", svc.lastBlockReason)
	}
}

func TestExecutionEndpointsMapValidationAndForbiddenErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeService{taskErr: service.ErrPhotoRequired, replaceTasksErr: service.ErrForbidden}
	c := NewController(svc)
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint(1))
		ctx.Set("role", policy.RoleUser)
		ctx.Set("team_id", int64(8))
		ctx.Next()
	})
	r.PATCH("/v1/large-work-tasks/:taskId/complete", c.CompleteTask)
	r.POST("/v1/large-work-items/:id/tasks", c.ReplaceTasks)

	req := httptest.NewRequest(http.MethodPatch, "/v1/large-work-tasks/11/complete", strings.NewReader(`{"completionNote":"done"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("complete status=%d want 400 body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/large-work-items/501/tasks", strings.NewReader(`{"tasks":[{"assignedTeamId":8,"pointLabel":"P1","workType":"tree"}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("replace status=%d want 403 body=%s", rec.Code, rec.Body.String())
	}
}
