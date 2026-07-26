package router

import (
	"os"
	"strings"
	"testing"
)

func TestMonthlyScheduleAndClinicIntegrationRoutesAreRegistered(t *testing.T) {
	source, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}
	body := string(source)
	expected := []string{
		`monthlySchedulesV1.GET("/:year/:month/schedule"`,
		`monthlySchedulesV1.PUT("/:year/:month/schedule/draft"`,
		`monthlySchedulesV1.POST("/:year/:month/schedule/publish"`,
		`integrationsV1.GET("/monthly-plans/:year/:month"`,
	}
	for _, route := range expected {
		if !strings.Contains(body, route) {
			t.Fatalf("router missing %s", route)
		}
	}
}
