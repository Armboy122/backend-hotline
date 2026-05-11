package router

import (
	"net/http"
	"testing"
	"time"

	"backend-hotlines3/pkg/jwt"
)

func TestLargeWorkRoutesAreRegistered(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	r := SetupRouter(testRouterConfig(), nil, jwtManager)

	routes := map[string]bool{}
	for _, route := range r.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	for _, route := range []string{
		http.MethodGet + " /v1/large-work-items",
		http.MethodGet + " /v1/large-work-items/:id",
		http.MethodPost + " /v1/large-work-items",
		http.MethodPatch + " /v1/large-work-items/:id",
		http.MethodPost + " /v1/large-work-items/:id/cancel",
		http.MethodPost + " /v1/large-work-items/:id/attachments",
		http.MethodGet + " /v1/large-work-items/:id/overview",
		http.MethodPost + " /v1/large-work-items/:id/tasks",
		http.MethodGet + " /v1/large-work-items/:id/tasks",
		http.MethodGet + " /v1/large-work-tasks/my-todos",
		http.MethodPatch + " /v1/large-work-tasks/:taskId/start",
		http.MethodPost + " /v1/large-work-tasks/:taskId/photos",
		http.MethodPatch + " /v1/large-work-tasks/:taskId/complete",
		http.MethodPatch + " /v1/large-work-tasks/:taskId/block",
	} {
		if !routes[route] {
			t.Fatalf("missing route %s", route)
		}
	}
}
