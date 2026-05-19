package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/middleware"
	"backend-hotlines3/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func TestUserAdminRoutesRejectAdminBeforeController(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	r := SetupRouter(testRouterConfig(), nil, jwtManager)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager, policy.RoleAdmin))
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s, want admin rejected before user controller", rec.Code, rec.Body.String())
	}
}

func TestMasterDataMutationsRejectAdminBeforeController(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	r := SetupRouter(testRouterConfig(), nil, jwtManager)

	for _, path := range []string{"/v1/teams", "/v1/job-types", "/v1/job-details", "/v1/feeders", "/v1/stations", "/v1/peas", "/v1/operation-centers"} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, path, nil)
			req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager, policy.RoleAdmin))
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d body=%s, want admin rejected before master-data controller", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestDashboardReadRolesRejectAdminAndAllowSuperAdmin(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	authMw := middleware.NewAuthMiddleware(jwtManager)

	for _, tc := range []struct {
		role     string
		wantCode int
	}{
		{role: policy.RoleAdmin, wantCode: http.StatusForbidden},
		{role: policy.RoleSuperAdmin, wantCode: http.StatusNoContent},
		{role: policy.RoleViewer, wantCode: http.StatusNoContent},
	} {
		t.Run(tc.role, func(t *testing.T) {
			r := gin.New()
			r.GET("/v1/dashboard/summary", authMw.RequireAuth(), authMw.RequireRole(dashboardReadRoles()...), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/dashboard/summary", nil)
			req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager, tc.role))
			r.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Fatalf("status = %d body=%s, want %d", rec.Code, rec.Body.String(), tc.wantCode)
			}
		})
	}
}

func TestSuperAdminOnlyRoleGateAllowsOnlySuperAdmin(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	authMw := middleware.NewAuthMiddleware(jwtManager)

	for _, tc := range []struct {
		role     string
		wantCode int
	}{
		{role: policy.RoleAdmin, wantCode: http.StatusForbidden},
		{role: policy.RoleSuperAdmin, wantCode: http.StatusNoContent},
	} {
		t.Run(tc.role, func(t *testing.T) {
			r := gin.New()
			r.POST("/guarded", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/guarded", nil)
			req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager, tc.role))
			r.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Fatalf("status = %d body=%s, want %d", rec.Code, rec.Body.String(), tc.wantCode)
			}
		})
	}
}

func TestMonthlyPlanManagerRoleGateAllowsOnlySuperAdmin(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	authMw := middleware.NewAuthMiddleware(jwtManager)

	for _, tc := range []struct {
		role     string
		wantCode int
	}{
		{role: policy.RoleAdmin, wantCode: http.StatusForbidden},
		{role: policy.RoleSuperAdmin, wantCode: http.StatusNoContent},
		{role: policy.RoleUser, wantCode: http.StatusForbidden},
		{role: policy.RoleViewer, wantCode: http.StatusForbidden},
	} {
		t.Run(tc.role, func(t *testing.T) {
			r := gin.New()
			r.PUT("/v1/monthly-plans/settings", authMw.RequireAuth(), authMw.RequireRole(monthlyPlanManagerRoles()...), func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/v1/monthly-plans/settings", nil)
			req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, jwtManager, tc.role))
			r.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Fatalf("status = %d body=%s, want %d", rec.Code, rec.Body.String(), tc.wantCode)
			}
		})
	}
}
