package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	userentity "backend-hotlines3/internal/feature/user/entity"
	"backend-hotlines3/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type fakePasswordStatusStore struct {
	mustChange bool
	err        error
}

func (s fakePasswordStatusStore) MustChangePassword(ctx context.Context, userID uint) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.mustChange, nil
}

func TestRequireAuthBlocksProtectedRouteWhilePasswordMustChange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	access, _, err := jwtManager.GenerateTokenPair(7, "700001", "user", nil)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	authMw := NewAuthMiddleware(jwtManager)
	authMw.SetPasswordStatusStore(fakePasswordStatusStore{mustChange: true})

	r := gin.New()
	r.GET("/v1/planning/calendar/2026/5", authMw.RequireAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/planning/calendar/2026/5", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s, want 403", rec.Code, rec.Body.String())
	}
}

func TestRequireAuthReturnsUnauthorizedWhenSessionUserIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	access, _, err := jwtManager.GenerateTokenPair(7, "700001", "user", nil)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	authMw := NewAuthMiddleware(jwtManager)
	authMw.SetPasswordStatusStore(fakePasswordStatusStore{err: userentity.ErrNotFound})

	r := gin.New()
	r.GET("/v1/planning/calendar/2026/5", authMw.RequireAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/planning/calendar/2026/5", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s, want 401", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"SESSION_USER_NOT_FOUND"`) {
		t.Fatalf("body=%s, want SESSION_USER_NOT_FOUND", rec.Body.String())
	}
}

func TestRequireAuthReturnsPasswordStatusErrorForStoreFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	access, _, err := jwtManager.GenerateTokenPair(7, "700001", "user", nil)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	authMw := NewAuthMiddleware(jwtManager)
	authMw.SetPasswordStatusStore(fakePasswordStatusStore{err: errors.New("db unavailable")})

	r := gin.New()
	r.GET("/v1/planning/calendar/2026/5", authMw.RequireAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/planning/calendar/2026/5", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d body=%s, want 500", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"PASSWORD_STATUS_ERROR"`) {
		t.Fatalf("body=%s, want PASSWORD_STATUS_ERROR", rec.Body.String())
	}
}

func TestRequireAuthAllowsPasswordChangeAndMeWhilePasswordMustChange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	access, _, err := jwtManager.GenerateTokenPair(7, "700001", "user", nil)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	authMw := NewAuthMiddleware(jwtManager)
	authMw.SetPasswordStatusStore(fakePasswordStatusStore{mustChange: true})

	r := gin.New()
	r.PUT("/v1/users/7/password", authMw.RequireAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/v1/auth/me", authMw.RequireAuth(), func(c *gin.Context) { c.Status(http.StatusOK) })

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPut, path: "/v1/users/7/password"},
		{method: http.MethodGet, path: "/v1/auth/me"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+access)
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
			}
		})
	}
}
