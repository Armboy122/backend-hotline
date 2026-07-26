package db

import (
	"reflect"
	"testing"

	"backend-hotlines3/internal/models"
)

func TestMigrationModelsIncludesMonthlyPlanScheduleInDependencyOrder(t *testing.T) {
	all := MigrationModels()
	index := func(want any) int {
		wantType := reflect.TypeOf(want)
		for i, model := range all {
			if reflect.TypeOf(model) == wantType {
				return i
			}
		}
		return -1
	}

	plan := index(&models.MonthlyPlan{})
	revision := index(&models.MonthlyPlanScheduleRevision{})
	assignment := index(&models.MonthlyPlanTeamAssignment{})
	if plan < 0 || revision < 0 || assignment < 0 {
		t.Fatalf("migration models missing schedule dependency: plan=%d revision=%d assignment=%d", plan, revision, assignment)
	}
	if !(plan < revision && revision < assignment) {
		t.Fatalf("migration order must be plan < revision < assignment, got %d < %d < %d", plan, revision, assignment)
	}
}
