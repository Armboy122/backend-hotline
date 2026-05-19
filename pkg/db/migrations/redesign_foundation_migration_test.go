package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRedesignFoundationMigrationDefinesCapabilities(t *testing.T) {
	migration := readSingleMigration(t, "*_prd_redesign_foundation.sql")

	requiredFragments := []string{
		"-- +goose Up",
		"CREATE TABLE IF NOT EXISTS user_capabilities",
		"user_id BIGINT NOT NULL REFERENCES \"User\"(id)",
		"code TEXT NOT NULL",
		"revoked_at TIMESTAMPTZ(6)",
		"CREATE UNIQUE INDEX IF NOT EXISTS user_capabilities_active_user_code_idx",
		"-- +goose Down",
		"DROP TABLE IF EXISTS user_capabilities",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(migration, fragment) {
			t.Fatalf("redesign foundation migration missing fragment %q\nMigration:\n%s", fragment, migration)
		}
	}
}

func readRedesignMigration(t *testing.T) string {
	t.Helper()
	matches, err := filepath.Glob("*_prd_redesign_foundation.sql")
	if err != nil {
		t.Fatalf("glob redesign migration: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly one redesign migration, got %d: %v", len(matches), matches)
	}
	contents, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read redesign migration: %v", err)
	}
	return string(contents)
}
