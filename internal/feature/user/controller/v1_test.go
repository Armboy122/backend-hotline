package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/user/entity"
	"backend-hotlines3/internal/feature/user/repository"
	"backend-hotlines3/internal/feature/user/service"

	"github.com/gin-gonic/gin"
)

func TestCreateForbidsLegacyAdminCreatingUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &controllerFakeRepo{activeSuperAdmins: 1}
	r := gin.New()
	ctrl := NewController(service.NewService(repo))
	r.POST("/users", withActor(10, policy.RoleAdmin), ctrl.Create)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"username":"123456","password":"secret1","role":"user"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s, want 403", rec.Code, rec.Body.String())
	}
	if repo.created.Role != "" {
		t.Fatalf("repository create should not be called, got role %q", repo.created.Role)
	}
}

func TestCreateRejectsLegacyAdminRolePayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &controllerFakeRepo{activeSuperAdmins: 1}
	r := gin.New()
	ctrl := NewController(service.NewService(repo))
	r.POST("/users", withActor(1, policy.RoleSuperAdmin), ctrl.Create)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"username":"123456","password":"secret1","role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s, want 400", rec.Code, rec.Body.String())
	}
	if repo.created.Role != "" {
		t.Fatalf("repository create should not be called for invalid role, got role %q", repo.created.Role)
	}
}

func TestResetPasswordForbidsAdminAndAllowsSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &controllerFakeRepo{users: map[uint]entity.UserInfo{2: {ID: 2, Role: policy.RoleUser, IsActive: true}}}
	ctrl := NewController(service.NewService(repo))
	r := gin.New()
	r.PUT("/users/:id/reset-password", withActor(10, policy.RoleAdmin), ctrl.ResetPassword)

	body := `{"newPassword":"newsecret"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/users/2/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("admin status = %d body=%s, want 403", rec.Code, rec.Body.String())
	}

	r = gin.New()
	r.PUT("/users/:id/reset-password", withActor(1, policy.RoleSuperAdmin), ctrl.ResetPassword)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/users/2/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("super_admin status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}
	if repo.updatedPasswordID != 2 || repo.updatedPasswordHash == "" {
		t.Fatalf("password reset not persisted: id=%d hash=%q", repo.updatedPasswordID, repo.updatedPasswordHash)
	}
}

func TestUserResponsesNeverExposePasswordHash(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &controllerFakeRepo{users: map[uint]entity.UserInfo{2: {ID: 2, Username: "123456", Role: policy.RoleUser, IsActive: true, CreatedAt: "2026-01-01T00:00:00Z"}}}
	ctrl := NewController(service.NewService(repo))
	r := gin.New()
	r.GET("/users/:id", withActor(1, policy.RoleSuperAdmin), ctrl.GetByID)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/2", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	data := payload["data"].(map[string]any)
	if _, exists := data["password"]; exists {
		t.Fatalf("response exposes password field: %s", rec.Body.String())
	}
}

func TestListContactsSearchReturnsContactFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	displayName := "Somchai"
	position := "Line technician"
	phone := "0812345678"
	repo := &controllerFakeRepo{users: map[uint]entity.UserInfo{2: {ID: 2, Username: "123456", Role: policy.RoleUser, DisplayName: &displayName, Position: &position, PhoneNumber: &phone, IsActive: true}}}
	ctrl := NewController(service.NewService(repo))
	r := gin.New()
	r.GET("/contact-directory", withActor(1, policy.RoleUser), ctrl.ListContacts)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/contact-directory?query=somchai", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}
	if repo.capturedContactQuery.Query != "somchai" {
		t.Fatalf("search query = %q, want somchai", repo.capturedContactQuery.Query)
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	data := payload["data"].([]any)[0].(map[string]any)
	if data["displayName"] != displayName || data["position"] != position || data["phoneNumber"] != phone {
		t.Fatalf("contact data missing fields: %s", rec.Body.String())
	}
}

func TestUpdateOwnContactAllowsAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &controllerFakeRepo{users: map[uint]entity.UserInfo{7: {ID: 7, Username: "777777", Role: policy.RoleUser, IsActive: true}}}
	ctrl := NewController(service.NewService(repo))
	r := gin.New()
	r.PATCH("/users/me/contact", withActor(7, policy.RoleUser), ctrl.UpdateOwnContact)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/users/me/contact", strings.NewReader(`{"displayName":"Somchai","position":"Line technician","phoneNumber":"0812345678"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}
	if repo.updatedContactID != 7 || repo.updatedContact.DisplayName == nil || *repo.updatedContact.DisplayName != "Somchai" {
		t.Fatalf("contact update not captured: id=%d input=%+v", repo.updatedContactID, repo.updatedContact)
	}
}

func TestUpdateContactForbidsOtherUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &controllerFakeRepo{users: map[uint]entity.UserInfo{8: {ID: 8, Username: "888888", Role: policy.RoleUser, IsActive: true}}}
	ctrl := NewController(service.NewService(repo))
	r := gin.New()
	r.PATCH("/users/:id/contact", withActor(7, policy.RoleUser), ctrl.UpdateContact)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/users/8/contact", strings.NewReader(`{"displayName":"Other"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s, want 403", rec.Code, rec.Body.String())
	}
	if repo.updatedContactID != 0 {
		t.Fatalf("repo update should not be called, got id %d", repo.updatedContactID)
	}
}

func withActor(id uint, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", id)
		c.Set("role", role)
		c.Next()
	}
}

type controllerFakeRepo struct {
	users                map[uint]entity.UserInfo
	activeSuperAdmins    int64
	created              entity.CreateInput
	capturedContactQuery entity.ContactListQuery
	updatedContactID     uint
	updatedContact       entity.UpdateContactInput
	updatedPasswordID    uint
	updatedPasswordHash  string
}

var _ repository.Repository = (*controllerFakeRepo)(nil)

func (r *controllerFakeRepo) List(context.Context, int, int) ([]entity.UserInfo, int64, error) {
	items := make([]entity.UserInfo, 0, len(r.users))
	for _, user := range r.users {
		items = append(items, user)
	}
	return items, int64(len(items)), nil
}

func (r *controllerFakeRepo) GetByID(_ context.Context, id uint) (entity.UserInfo, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return entity.UserInfo{}, entity.ErrNotFound
}

func (r *controllerFakeRepo) Create(_ context.Context, in entity.CreateInput) (entity.UserInfo, error) {
	r.created = in
	return entity.UserInfo{ID: 100, Username: in.Username, Role: in.Role, TeamID: in.TeamID, IsActive: in.IsActive}, nil
}

func (r *controllerFakeRepo) Update(_ context.Context, id uint, in entity.UpdateInput) (entity.UserInfo, error) {
	user, ok := r.users[id]
	if !ok {
		return entity.UserInfo{}, entity.ErrNotFound
	}
	if in.Role != nil {
		user.Role = *in.Role
	}
	return user, nil
}

func (r *controllerFakeRepo) Delete(context.Context, uint) error { return nil }
func (r *controllerFakeRepo) GetPasswordHash(context.Context, uint) (string, error) {
	return "$2a$12$BL9bVdgl.6mnHeYvuewIEeJcDObTvWNiyzxkSvjKw8aUP3XvR2lRm", nil
}
func (r *controllerFakeRepo) UpdatePassword(_ context.Context, id uint, hashed string) error {
	r.updatedPasswordID = id
	r.updatedPasswordHash = hashed
	return nil
}
func (r *controllerFakeRepo) CountActiveSuperAdmins(context.Context, *uint) (int64, error) {
	return r.activeSuperAdmins, nil
}

func (r *controllerFakeRepo) ListContacts(_ context.Context, q entity.ContactListQuery) ([]entity.ContactInfo, int64, error) {
	r.capturedContactQuery = q
	items := []entity.ContactInfo{}
	for _, user := range r.users {
		items = append(items, entity.ContactInfo{ID: user.ID, Username: user.Username, Role: user.Role, TeamID: user.TeamID, DisplayName: user.DisplayName, Position: user.Position, PhoneNumber: user.PhoneNumber, IsActive: user.IsActive})
	}
	return items, int64(len(items)), nil
}

func (r *controllerFakeRepo) GetContactByID(_ context.Context, id uint) (entity.ContactInfo, error) {
	if user, ok := r.users[id]; ok {
		return entity.ContactInfo{ID: user.ID, Username: user.Username, Role: user.Role, TeamID: user.TeamID, DisplayName: user.DisplayName, Position: user.Position, PhoneNumber: user.PhoneNumber, IsActive: user.IsActive}, nil
	}
	return entity.ContactInfo{}, entity.ErrNotFound
}

func (r *controllerFakeRepo) UpdateContact(_ context.Context, id uint, in entity.UpdateContactInput) (entity.ContactInfo, error) {
	r.updatedContactID = id
	r.updatedContact = in
	user, ok := r.users[id]
	if !ok {
		return entity.ContactInfo{}, entity.ErrNotFound
	}
	if in.DisplayName != nil {
		user.DisplayName = in.DisplayName
	}
	if in.Position != nil {
		user.Position = in.Position
	}
	if in.PhoneNumber != nil {
		user.PhoneNumber = in.PhoneNumber
	}
	return entity.ContactInfo{ID: user.ID, Username: user.Username, Role: user.Role, TeamID: user.TeamID, DisplayName: user.DisplayName, Position: user.Position, PhoneNumber: user.PhoneNumber, IsActive: user.IsActive}, nil
}
