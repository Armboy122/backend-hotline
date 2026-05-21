package migrations_test

import (
	"strings"
	"testing"
)

func TestUserPasswordStatusMigrationAddsMustChangePassword(t *testing.T) {
	migration := readSingleMigration(t, "*_add_user_must_change_password.sql")

	requiredFragments := []string{
		"-- +goose Up",
		"ALTER TABLE \"User\"",
		"ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT false",
		"-- +goose Down",
		"DROP COLUMN IF EXISTS must_change_password",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(migration, fragment) {
			t.Fatalf("user password status migration missing fragment %q\nMigration:\n%s", fragment, migration)
		}
	}
}
