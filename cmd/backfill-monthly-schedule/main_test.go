package main

import (
	"path/filepath"
	"testing"

	"backend-hotlines3/internal/feature/monthlyschedule/entity"
)

func TestReviewedJulyFixtureMapsOnlyAwayPeriodsByStableTeamCode(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "monthly-schedule", "2026-07-clinic-tool.json")
	input, err := readFixture(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	teams := make([]entity.Team, 0, len(input.Teams))
	for index, item := range input.Teams {
		code := item.TeamCode
		base := item.BaseArea
		crew := item.CrewType
		teams = append(teams, entity.Team{
			ID:                 int64(index + 1),
			Name:               item.BaseArea,
			Code:               &code,
			BaseArea:           &base,
			CrewType:           &crew,
			DisplayOrder:       item.DisplayOrder,
			MonthlyPlanVisible: true,
		})
	}
	assignments, missing, err := mapFixture(input, teams)
	if err != nil {
		t.Fatalf("map fixture: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("missing codes = %v", missing)
	}
	if len(assignments) != 24 {
		t.Fatalf("away assignments = %d, want 24", len(assignments))
	}
	for _, assignment := range assignments {
		if assignment.AssignmentType == "home" {
			t.Fatal("home periods must be derived, never imported")
		}
		if assignment.SourceType != "manual" {
			t.Fatalf("source type = %q, want manual", assignment.SourceType)
		}
	}
}
