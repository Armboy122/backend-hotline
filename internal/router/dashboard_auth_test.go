package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/middleware"
	"backend-hotlines3/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func TestDashboardRoutesRequireBearerAuth(t *testing.T) {
	r := SetupRouter(testRouterConfig(), nil, jwt.NewJWTManager("secret", time.Hour, 24*time.Hour))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/dashboard/summary", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s, want %d", rec.Code, rec.Body.String(), http.StatusUnauthorized)
	}
}

func TestDashboardRoutesRejectNonDashboardReadRoles(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	r := SetupRouter(testRouterConfig(), nil, jwtManager)
	token := mustAccessToken(t, jwtManager, "user")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/dashboard/summary", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s, want %d", rec.Code, rec.Body.String(), http.StatusForbidden)
	}
}

func TestDashboardRoleGateAllowsApprovedDashboardReadRoles(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	authMw := middleware.NewAuthMiddleware(jwtManager)

	for _, role := range []string{"super_admin", "viewer"} {
		t.Run(role, func(t *testing.T) {
			r := gin.New()
			r.GET("/v1/dashboard/summary", authMw.RequireAuth(), authMw.RequireRole(dashboardReadRoles()...), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/dashboard/summary", nil)
			req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager, role))
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d body=%s, want role %s allowed", rec.Code, rec.Body.String(), role)
			}
		})
	}
}

func TestDashboardSummaryCanUsePrivateCacheForRepeatedReads(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	authMw := middleware.NewAuthMiddleware(jwtManager)
	handlerCalls := 0

	r := gin.New()
	r.GET("/v1/dashboard/summary", authMw.RequireAuth(), authMw.RequireRole(dashboardReadRoles()...), middleware.CachePrivate(60), func(c *gin.Context) {
		handlerCalls++
		c.JSON(http.StatusOK, gin.H{"calls": handlerCalls})
	})

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/dashboard/summary?year=2026", nil)
		req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager, "super_admin"))
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d body=%s, want 200", i+1, rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Cache-Control"); got != "private, max-age=60" {
			t.Fatalf("request %d Cache-Control = %q, want private, max-age=60", i+1, got)
		}
	}
	if handlerCalls != 1 {
		t.Fatalf("handler calls = %d, want dashboard cache to serve repeated read", handlerCalls)
	}
}

func mustAccessToken(t *testing.T, manager *jwt.JWTManager, role string) string {
	t.Helper()
	teamID := int64(1)
	accessToken, _, err := manager.GenerateTokenPair(1, "tester", role, &teamID)
	if err != nil {
		t.Fatalf("GenerateTokenPair(%s): %v", role, err)
	}
	return accessToken
}

func testRouterConfig() *config.Config {
	return &config.Config{
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Authorization", "Content-Type"},
		},
	}
}
