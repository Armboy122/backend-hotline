package router

import (
	"net/http"
	"os"
	"strings"
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

func TestMonthlyPlanConvertToPlanningRouteIsRegistered(t *testing.T) {
	source, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}
	if !strings.Contains(string(source), `monthlyPlansV1.POST("/:year/:month/convert-to-planning", monthlyPlanController.ConvertApprovedToPlanning)`) {
		t.Fatal("missing monthly plan convert-to-planning route registration")
	}
}
