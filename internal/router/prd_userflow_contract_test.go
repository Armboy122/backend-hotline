package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"backend-hotlines3/pkg/jwt"
)

func TestPRDProtectedRoutesRejectUnauthenticatedRequestsBeforeControllers(t *testing.T) {
	r := SetupRouter(testRouterConfig(), nil, jwt.NewJWTManager("secret", time.Hour, 24*time.Hour))

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/v1/team-plans"},
		{http.MethodPost, "/v1/team-plans"},
		{http.MethodGet, "/v1/planning/calendar/2026/5"},
		{http.MethodGet, "/v1/large-work-items"},
		{http.MethodGet, "/v1/large-work-tasks/my-todos"},
		{http.MethodGet, "/v1/tasks"},
		{http.MethodPost, "/v1/tasks"},
		{http.MethodGet, "/v1/work-reports?year=2026&month=5"},
		{http.MethodGet, "/v1/daily-report-drafts/sources"},
		{http.MethodGet, "/v1/contact-directory"},
		{http.MethodPatch, "/v1/users/me/contact"},
		{http.MethodPost, "/v1/auth/register"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d body=%s, want 401 before controller", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestPRDRouterSourceKeepsWriteAndFileRoutesBehindAuth(t *testing.T) {
	source, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}
	text := string(source)

	requiredFragments := []string{
		`authGroup.POST("/register", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), aCtrl.Register)`,
		`tasksV1.Use(authMw.RequireAuth())`,
		`teamPlansV1.Use(authMw.RequireAuth())`,
		`planningV1.Use(authMw.RequireAuth())`,
		`largeWorksV1.Use(authMw.RequireAuth())`,
		`largeWorksTasksV1.Use(authMw.RequireAuth())`,
		`workReportsV1.Use(authMw.RequireAuth())`,
		`dailyDraftV1.Use(authMw.RequireAuth())`,
		`contactsV1.Use(authMw.RequireAuth())`,
		`usersV1.Use(authMw.RequireAuth())`,
		`uploadV1.Use(authMw.RequireAuth())`,
		`monthlyPlansV1.Use(authMw.RequireAuth()`,
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(text, fragment) {
			t.Fatalf("router.go missing auth fragment %q", fragment)
		}
	}
}
