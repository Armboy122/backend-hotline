package router

import (
	"net/http"
	"testing"
	"time"

	"backend-hotlines3/pkg/jwt"
)

func TestPlanningCalendarRoutesAreRegistered(t *testing.T) {
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	r := SetupRouter(testRouterConfig(), nil, jwtManager)

	routes := map[string]bool{}
	for _, route := range r.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	if !routes[http.MethodGet+" /v1/planning/calendar/:year/:month"] {
		t.Fatal("missing planning calendar route")
	}
}
