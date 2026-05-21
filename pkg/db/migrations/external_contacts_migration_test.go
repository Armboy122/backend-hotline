package migrations_test

import (
	"strings"
	"testing"
)

func TestContactDirectoryExternalContactsMigrationExists(t *testing.T) {
	migration := readSingleMigration(t, "*_create_external_contacts.sql")

	requiredFragments := []string{
		"-- +goose Up",
		"CREATE TABLE IF NOT EXISTS external_contacts",
		"type TEXT NOT NULL",
		"display_name TEXT NOT NULL",
		"phone_number TEXT NOT NULL",
		"team_id BIGINT REFERENCES \"Team\"(id)",
		"is_active BOOLEAN NOT NULL DEFAULT TRUE",
		"deleted_at TIMESTAMPTZ(6)",
		"CREATE INDEX IF NOT EXISTS idx_external_contacts_type",
		"CREATE INDEX IF NOT EXISTS idx_external_contacts_team",
		"-- +goose Down",
		"DROP TABLE IF EXISTS external_contacts",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(migration, fragment) {
			t.Fatalf("external contacts migration missing fragment %q\nMigration:\n%s", fragment, migration)
		}
	}
}
