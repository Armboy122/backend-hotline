package db

import (
	"testing"

	"backend-hotlines3/internal/models"
)

func TestMigrationModelsIncludeLargeWorkFoundation(t *testing.T) {
	migrationModels := MigrationModels()

	assertMigrationModelRegistered[models.LargeWorkItem](t, migrationModels)
	assertMigrationModelRegistered[models.LargeWorkItemTeam](t, migrationModels)
	assertMigrationModelRegistered[models.LargeWorkTask](t, migrationModels)
}

func TestMigrationModelsIncludeContactDirectoryExternalContacts(t *testing.T) {
	migrationModels := MigrationModels()

	assertMigrationModelRegistered[models.ExternalContact](t, migrationModels)
}

func assertMigrationModelRegistered[T any](t *testing.T, migrationModels []any) {
	t.Helper()
	for _, model := range migrationModels {
		if _, ok := model.(*T); ok {
			return
		}
	}
	var target T
	t.Fatalf("migration model %T is not registered in MigrationModels", &target)
}
