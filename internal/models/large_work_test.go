package models

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLargeWorkItemModelMatchesMobilizationContract(t *testing.T) {
	if got := (LargeWorkItem{}).TableName(); got != "large_work_items" {
		t.Fatalf("LargeWorkItem table name = %q, want large_work_items", got)
	}

	typ := reflect.TypeOf(LargeWorkItem{})
	assertLargeWorkField(t, typ, "ID", "column:id")
	assertLargeWorkField(t, typ, "OwnerTeamID", "column:owner_team_id")
	assertLargeWorkField(t, typ, "CreatedByUserID", "column:created_by_user_id")
	assertLargeWorkField(t, typ, "Title", "column:title")
	assertLargeWorkField(t, typ, "WorkType", "column:work_type")
	assertLargeWorkField(t, typ, "StartDate", "type:date;column:start_date")
	assertLargeWorkField(t, typ, "EndDate", "type:date;column:end_date")
	assertLargeWorkField(t, typ, "WorkTime", "column:work_time")
	assertLargeWorkField(t, typ, "LocationText", "column:location_text")
	assertLargeWorkField(t, typ, "PEAID", "column:pea_id")
	assertLargeWorkField(t, typ, "OperationCenterID", "column:operation_center_id")
	assertLargeWorkField(t, typ, "FeederID", "column:feeder_id")
	assertLargeWorkField(t, typ, "StationID", "column:station_id")
	assertLargeWorkField(t, typ, "Notes", "column:notes")
	assertLargeWorkField(t, typ, "Status", "default:planned;column:status")
	assertLargeWorkField(t, typ, "CreatedAt", "column:created_at")
	assertLargeWorkField(t, typ, "UpdatedAt", "column:updated_at")
	assertLargeWorkField(t, typ, "DeletedAt", "column:deleted_at")

	item := LargeWorkItem{
		OwnerTeamID:     1,
		CreatedByUserID: 9,
		Title:           "งานระดมทีม เปลี่ยนอุปกรณ์หลัก",
		StartDate:       mustDate(t, "2026-06-10"),
		LocationText:    "Station A feeder F01",
		Status:          LargeWorkStatusPlanned,
	}
	if item.Status != "planned" {
		t.Fatalf("planned status constant = %q, want planned", item.Status)
	}
}

func TestLargeWorkParticipantTeamModelMatchesMobilizationContract(t *testing.T) {
	if got := (LargeWorkItemTeam{}).TableName(); got != "large_work_item_teams" {
		t.Fatalf("LargeWorkItemTeam table name = %q, want large_work_item_teams", got)
	}

	typ := reflect.TypeOf(LargeWorkItemTeam{})
	assertLargeWorkField(t, typ, "LargeWorkItemID", "column:large_work_item_id")
	assertLargeWorkField(t, typ, "TeamID", "column:team_id")
	assertLargeWorkField(t, typ, "Role", "column:role")
	assertLargeWorkField(t, typ, "ParticipantStatus", "default:assigned;column:participant_status")
	assertLargeWorkField(t, typ, "CreatedAt", "column:created_at")
	assertLargeWorkField(t, typ, "AcknowledgedByUserID", "column:acknowledged_by_user_id")
	assertLargeWorkField(t, typ, "AcknowledgedAt", "column:acknowledged_at")

	team := LargeWorkItemTeam{
		LargeWorkItemID:      501,
		TeamID:               1,
		Role:                 LargeWorkTeamRoleOwner,
		ParticipantStatus:    LargeWorkParticipantStatusAssigned,
		AcknowledgedAt:       nil,
		AcknowledgedByUserID: nil,
	}
	if team.Role != "owner" {
		t.Fatalf("owner role constant = %q, want owner", team.Role)
	}
	if team.ParticipantStatus != "assigned" {
		t.Fatalf("assigned participant status = %q, want assigned", team.ParticipantStatus)
	}
}

func TestLargeWorkStatusAndParticipantValuesStaySmallForMVP(t *testing.T) {
	statuses := []string{
		LargeWorkStatusDraft,
		LargeWorkStatusPlanned,
		LargeWorkStatusInProgress,
		LargeWorkStatusCompleted,
		LargeWorkStatusCancelled,
	}
	wantStatuses := []string{"draft", "planned", "in_progress", "completed", "cancelled"}
	if !reflect.DeepEqual(statuses, wantStatuses) {
		t.Fatalf("large work statuses = %#v, want %#v", statuses, wantStatuses)
	}

	roles := []string{LargeWorkTeamRoleOwner, LargeWorkTeamRoleParticipant}
	if want := []string{"owner", "participant"}; !reflect.DeepEqual(roles, want) {
		t.Fatalf("large work team roles = %#v, want %#v", roles, want)
	}

	participantStatuses := []string{
		LargeWorkParticipantStatusAssigned,
		LargeWorkParticipantStatusAcknowledged,
	}
	if want := []string{"assigned", "acknowledged"}; !reflect.DeepEqual(participantStatuses, want) {
		t.Fatalf("participant statuses = %#v, want %#v", participantStatuses, want)
	}
}

func assertLargeWorkField(t *testing.T, typ reflect.Type, name string, tagContains string) {
	t.Helper()
	field, ok := typ.FieldByName(name)
	if !ok {
		t.Fatalf("field %s missing on %s", name, typ.Name())
	}
	if got := field.Tag.Get("gorm"); !strings.Contains(got, tagContains) {
		t.Fatalf("field %s gorm tag = %q, want to contain %q", name, got, tagContains)
	}
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}
	return parsed
}
