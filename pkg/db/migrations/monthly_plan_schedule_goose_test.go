package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestMonthlyPlanScheduleMigrationOwnsProductionSchema(t *testing.T) {
	raw, err := os.ReadFile("20260726113000_create_monthly_plan_schedules.sql")
	if err != nil {
		t.Fatalf("read monthly plan schedule migration: %v", err)
	}
	sql := string(raw)
	for _, fragment := range []string{
		`ALTER TABLE "Team" ADD COLUMN IF NOT EXISTS code TEXT`,
		"CREATE TABLE IF NOT EXISTS monthly_plan_schedule_revisions",
		"CREATE TABLE IF NOT EXISTS monthly_plan_team_assignments",
		"monthly_plan_schedule_single_draft_idx",
		"monthly_plan_schedule_single_published_idx",
		"projection JSONB",
		"CHECK (end_date >= start_date)",
		"-- +goose Down",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("monthly plan schedule migration missing %q", fragment)
		}
	}
}
