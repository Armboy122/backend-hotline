package router

import (
	"net/http"
	"testing"
	"time"

	"backend-hotlines3/pkg/jwt"
)

func TestTeamPlanRoutesAreRegistered(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	r := SetupRouter(testRouterConfig(), nil, jwtManager)

	routes := map[string]bool{}
	for _, route := range r.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	for _, route := range []string{
		http.MethodGet + " /v1/team-plans",
		http.MethodGet + " /v1/team-plans/:id",
		http.MethodPost + " /v1/team-plans",
		http.MethodPut + " /v1/team-plans/:id",
		http.MethodDelete + " /v1/team-plans/:id",
	} {
		if !routes[route] {
			t.Fatalf("missing route %s", route)
		}
	}
}
