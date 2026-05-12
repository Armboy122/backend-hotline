package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLargeWorkTaskSequenceIndexUpgradeMigration(t *testing.T) {
	initial := readSingleMigration(t, "*_create_large_work_items.sql")
	upgrade := readSingleMigration(t, "*_update_large_work_tasks_team_sequence_index.sql")

	if !strings.Contains(initial, "CREATE UNIQUE INDEX IF NOT EXISTS idx_large_work_tasks_plan_sequence ON large_work_tasks (large_work_item_id, sequence)") {
		t.Fatalf("initial large work migration must keep original plan+sequence index for already-versioned goose migration\nMigration:\n%s", initial)
	}
	if strings.Contains(initial, "idx_large_work_tasks_plan_team_sequence") {
		t.Fatalf("initial large work migration must not rely on the new team sequence index because existing DBs will not re-run it")
	}

	requiredUpgradeFragments := []string{
		"-- +goose Up",
		"DROP INDEX IF EXISTS idx_large_work_tasks_plan_sequence",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_large_work_tasks_plan_team_sequence",
		"ON large_work_tasks (large_work_item_id, assigned_team_id, sequence)",
		"-- +goose Down",
		"DROP INDEX IF EXISTS idx_large_work_tasks_plan_team_sequence",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_large_work_tasks_plan_sequence",
		"ON large_work_tasks (large_work_item_id, sequence)",
	}
	for _, fragment := range requiredUpgradeFragments {
		if !strings.Contains(upgrade, fragment) {
			t.Fatalf("large work team sequence upgrade migration missing fragment %q\nMigration:\n%s", fragment, upgrade)
		}
	}
}

func readSingleMigration(t *testing.T, pattern string) string {
	t.Helper()
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob migration %s: %v", pattern, err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly one migration for %s, got %d: %v", pattern, len(matches), matches)
	}
	contents, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read migration %s: %v", matches[0], err)
	}
	return string(contents)
}
