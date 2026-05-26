package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateTeamPlansMigrationDefinesPersistenceFoundation(t *testing.T) {
	migration := readTeamPlanMigration(t)

	requiredFragments := []string{
		"-- +goose Up",
		"CREATE TABLE IF NOT EXISTS team_plans",
		"team_id BIGINT NOT NULL",
		"created_by_user_id BIGINT NOT NULL",
		"start_date DATE NOT NULL",
		"end_date DATE",
		"work_time TEXT",
		"location_text TEXT NOT NULL",
		"pea_id BIGINT",
		"operation_center_id BIGINT",
		"feeder_id BIGINT",
		"station_id BIGINT",
		"status TEXT NOT NULL DEFAULT 'planned'",
		"deleted_at TIMESTAMPTZ",
		"CHECK (status IN ('planned', 'cancelled', 'completed'))",
		"CHECK (end_date IS NULL OR end_date >= start_date)",
		"idx_team_plans_team_start_date",
		"idx_team_plans_date_range",
		"idx_team_plans_created_by",
		"idx_team_plans_active_range",
		"-- +goose Down",
		"DROP TABLE IF EXISTS team_plans",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(migration, fragment) {
			t.Fatalf("team plan migration missing fragment %q\nMigration:\n%s", fragment, migration)
		}
	}
}

func TestTeamPlansMigrationAllowsUnscheduledBacklogItems(t *testing.T) {
	matches, err := filepath.Glob("*_allow_unscheduled_team_plans.sql")
	if err != nil {
		t.Fatalf("glob unscheduled team plan migration: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly one unscheduled team plan migration, got %d: %v", len(matches), matches)
	}
	contents, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read unscheduled team plan migration: %v", err)
	}
	migration := string(contents)

	requiredFragments := []string{
		"-- +goose Up",
		"ALTER COLUMN start_date DROP NOT NULL",
		"DROP CONSTRAINT IF EXISTS team_plans_status_check",
		"CHECK (status IN ('draft', 'planned', 'cancelled', 'completed'))",
		"-- +goose Down",
		"ALTER COLUMN start_date SET NOT NULL",
		"CHECK (status IN ('planned', 'cancelled', 'completed'))",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(migration, fragment) {
			t.Fatalf("unscheduled team plan migration missing fragment %q\nMigration:\n%s", fragment, migration)
		}
	}
}

func readTeamPlanMigration(t *testing.T) string {
	t.Helper()
	matches, err := filepath.Glob("*_create_team_plans.sql")
	if err != nil {
		t.Fatalf("glob team plan migration: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly one team plan migration, got %d: %v", len(matches), matches)
	}
	contents, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read team plan migration: %v", err)
	}
	return string(contents)
}
