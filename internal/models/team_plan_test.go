package models

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestTeamPlanModelMatchesPlanningPersistenceContract(t *testing.T) {
	modelType := reflect.TypeOf(TeamPlan{})

	assertTeamPlanField(t, modelType, "ID", "id")
	assertTeamPlanField(t, modelType, "TeamID", "team_id")
	assertTeamPlanField(t, modelType, "CreatedByUserID", "created_by_user_id")
	assertTeamPlanField(t, modelType, "Title", "title")
	assertTeamPlanField(t, modelType, "WorkType", "work_type")
	assertTeamPlanField(t, modelType, "StartDate", "start_date")
	assertTeamPlanField(t, modelType, "EndDate", "end_date")
	assertTeamPlanField(t, modelType, "WorkTime", "work_time")
	assertTeamPlanField(t, modelType, "LocationText", "location_text")
	assertTeamPlanField(t, modelType, "PEAID", "pea_id")
	assertTeamPlanField(t, modelType, "OperationCenterID", "operation_center_id")
	assertTeamPlanField(t, modelType, "FeederID", "feeder_id")
	assertTeamPlanField(t, modelType, "StationID", "station_id")
	assertTeamPlanField(t, modelType, "Notes", "notes")
	assertTeamPlanField(t, modelType, "Status", "status")
	assertTeamPlanField(t, modelType, "DailyTaskID", "daily_task_id")
	assertTeamPlanField(t, modelType, "CreatedAt", "created_at")
	assertTeamPlanField(t, modelType, "UpdatedAt", "updated_at")
	assertTeamPlanField(t, modelType, "DeletedAt", "deleted_at")

	if (TeamPlan{}).TableName() != "team_plans" {
		t.Fatalf("TeamPlan table name = %q, want team_plans", (TeamPlan{}).TableName())
	}
}

func TestTeamPlanDateAndOptionalFieldsUseExpectedGoTypes(t *testing.T) {
	modelType := reflect.TypeOf(TeamPlan{})

	assertType(t, modelType, "StartDate", reflect.TypeOf((*time.Time)(nil)))
	assertType(t, modelType, "EndDate", reflect.TypeOf((*time.Time)(nil)))
	assertType(t, modelType, "WorkTime", reflect.TypeOf((*string)(nil)))
	assertType(t, modelType, "WorkType", reflect.TypeOf((*string)(nil)))
	assertType(t, modelType, "PEAID", reflect.TypeOf((*int64)(nil)))
	assertType(t, modelType, "OperationCenterID", reflect.TypeOf((*int64)(nil)))
	assertType(t, modelType, "FeederID", reflect.TypeOf((*int64)(nil)))
	assertType(t, modelType, "StationID", reflect.TypeOf((*int64)(nil)))
	assertType(t, modelType, "DailyTaskID", reflect.TypeOf((*int64)(nil)))
	assertType(t, modelType, "DeletedAt", reflect.TypeOf((*time.Time)(nil)))
}

func TestTeamPlanStatusConstants(t *testing.T) {
	statuses := []string{TeamPlanStatusDraft, TeamPlanStatusPlanned, TeamPlanStatusCancelled, TeamPlanStatusCompleted}
	want := []string{"draft", "planned", "cancelled", "completed"}
	for i := range want {
		if statuses[i] != want[i] {
			t.Fatalf("status[%d] = %q, want %q", i, statuses[i], want[i])
		}
	}
}

func assertTeamPlanField(t *testing.T, modelType reflect.Type, fieldName, columnName string) {
	t.Helper()
	field, ok := modelType.FieldByName(fieldName)
	if !ok {
		t.Fatalf("TeamPlan missing field %s", fieldName)
	}
	gormTag := field.Tag.Get("gorm")
	if !strings.Contains(gormTag, "column:"+columnName) {
		t.Fatalf("TeamPlan.%s gorm tag = %q, want column:%s", fieldName, gormTag, columnName)
	}
}

func assertType(t *testing.T, modelType reflect.Type, fieldName string, want reflect.Type) {
	t.Helper()
	field, ok := modelType.FieldByName(fieldName)
	if !ok {
		t.Fatalf("TeamPlan missing field %s", fieldName)
	}
	if field.Type != want {
		t.Fatalf("TeamPlan.%s type = %s, want %s", fieldName, field.Type, want)
	}
}
