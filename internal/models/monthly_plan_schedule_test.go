package models

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestMonthlyPlanScheduleModelsMatchPersistenceContract(t *testing.T) {
	revisionType := reflect.TypeOf(MonthlyPlanScheduleRevision{})
	assertScheduleField(t, revisionType, "MonthlyPlanID", "monthly_plan_id", reflect.TypeOf(int64(0)))
	assertScheduleField(t, revisionType, "RevisionNo", "revision_no", reflect.TypeOf(int(0)))
	assertScheduleField(t, revisionType, "Status", "status", reflect.TypeOf(""))
	assertScheduleField(t, revisionType, "PublishedAt", "published_at", reflect.TypeOf((*time.Time)(nil)))
	assertScheduleField(t, revisionType, "Checksum", "checksum", reflect.TypeOf((*string)(nil)))
	assertScheduleField(t, revisionType, "Projection", "projection", reflect.TypeOf([]byte(nil)))

	assignmentType := reflect.TypeOf(MonthlyPlanTeamAssignment{})
	assertScheduleField(t, assignmentType, "RevisionID", "revision_id", reflect.TypeOf(int64(0)))
	assertScheduleField(t, assignmentType, "TeamID", "team_id", reflect.TypeOf(int64(0)))
	assertScheduleField(t, assignmentType, "AssignmentType", "assignment_type", reflect.TypeOf(""))
	assertScheduleField(t, assignmentType, "StartDate", "start_date", reflect.TypeOf(time.Time{}))
	assertScheduleField(t, assignmentType, "EndDate", "end_date", reflect.TypeOf(time.Time{}))
	assertScheduleField(t, assignmentType, "Destination", "destination", reflect.TypeOf(""))

	if got := (MonthlyPlanScheduleRevision{}).TableName(); got != "monthly_plan_schedule_revisions" {
		t.Fatalf("revision table = %q", got)
	}
	if got := (MonthlyPlanTeamAssignment{}).TableName(); got != "monthly_plan_team_assignments" {
		t.Fatalf("assignment table = %q", got)
	}
}

func TestTeamContainsClinicIntegrationIdentity(t *testing.T) {
	typ := reflect.TypeOf(Team{})
	assertScheduleField(t, typ, "Code", "code", reflect.TypeOf((*string)(nil)))
	assertScheduleField(t, typ, "BaseArea", "base_area", reflect.TypeOf((*string)(nil)))
	assertScheduleField(t, typ, "CrewType", "crew_type", reflect.TypeOf((*string)(nil)))
	assertScheduleField(t, typ, "DisplayOrder", "display_order", reflect.TypeOf(int(0)))
	assertScheduleField(t, typ, "MonthlyPlanVisible", "monthly_plan_visible", reflect.TypeOf(true))
}

func TestMonthlyPlanScheduleConstantsAreStable(t *testing.T) {
	gotStatuses := []string{
		MonthlyPlanScheduleStatusDraft,
		MonthlyPlanScheduleStatusPublished,
		MonthlyPlanScheduleStatusSuperseded,
	}
	if strings.Join(gotStatuses, ",") != "draft,published,superseded" {
		t.Fatalf("unexpected statuses: %v", gotStatuses)
	}
	gotTypes := []string{
		MonthlyPlanAssignmentTypeField,
		MonthlyPlanAssignmentTypeRemote,
		MonthlyPlanAssignmentTypeSupport,
		MonthlyPlanAssignmentTypeSpecial,
	}
	if strings.Join(gotTypes, ",") != "field,remote,support,special" {
		t.Fatalf("unexpected assignment types: %v", gotTypes)
	}
}

func assertScheduleField(t *testing.T, typ reflect.Type, name, column string, wantType reflect.Type) {
	t.Helper()
	field, ok := typ.FieldByName(name)
	if !ok {
		t.Fatalf("%s missing field %s", typ.Name(), name)
	}
	if field.Type != wantType {
		t.Fatalf("%s.%s type = %s, want %s", typ.Name(), name, field.Type, wantType)
	}
	if !strings.Contains(field.Tag.Get("gorm"), "column:"+column) {
		t.Fatalf("%s.%s gorm tag = %q, want column:%s", typ.Name(), name, field.Tag.Get("gorm"), column)
	}
}
