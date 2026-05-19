package service

import (
	"context"
	"errors"
	"testing"
	"time"

	largeworkentity "backend-hotlines3/internal/feature/largework/entity"
	largeworkrepository "backend-hotlines3/internal/feature/largework/repository"
	monthlyentity "backend-hotlines3/internal/feature/monthlyplan/entity"
	monthlyrepository "backend-hotlines3/internal/feature/monthlyplan/repository"
	taskentity "backend-hotlines3/internal/feature/task/entity"
	teamplanentity "backend-hotlines3/internal/feature/teamplan/entity"
	teamplanrepository "backend-hotlines3/internal/feature/teamplan/repository"
)

type fakeTeamPlanRepo struct {
	listQuery     teamplanrepository.ListQuery
	getByIDQuery  teamplanrepository.GetQuery
	listResult    []teamplanentity.TeamPlan
	getByIDResult *teamplanentity.TeamPlan
	listErr       error
	getByIDErr    error
}

func (f *fakeTeamPlanRepo) List(_ context.Context, query teamplanrepository.ListQuery) ([]teamplanentity.TeamPlan, int64, error) {
	f.listQuery = query
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	return f.listResult, int64(len(f.listResult)), nil
}

func (f *fakeTeamPlanRepo) GetByID(_ context.Context, query teamplanrepository.GetQuery) (*teamplanentity.TeamPlan, error) {
	f.getByIDQuery = query
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}
	return f.getByIDResult, nil
}

type fakeMonthlyRepo struct {
	periodQuery  monthlyrepository.PeriodFindQuery
	listQuery    monthlyrepository.PlanFileListQuery
	getByIDID    int64
	periodResult *monthlyentity.Entity
	filesResult  []monthlyentity.PlanFileEntity
	fileResult   *monthlyentity.PlanFileEntity
	periodErr    error
	listErr      error
	getByIDErr   error
}

func (f *fakeMonthlyRepo) FindPeriod(_ context.Context, q monthlyrepository.PeriodFindQuery) (*monthlyentity.Entity, error) {
	f.periodQuery = q
	if f.periodErr != nil {
		return nil, f.periodErr
	}
	return f.periodResult, nil
}

func (f *fakeMonthlyRepo) ListFiles(_ context.Context, q monthlyrepository.PlanFileListQuery) ([]monthlyentity.PlanFileEntity, error) {
	f.listQuery = q
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.filesResult, nil
}

func (f *fakeMonthlyRepo) GetFileByID(_ context.Context, id int64) (*monthlyentity.PlanFileEntity, error) {
	f.getByIDID = id
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}
	return f.fileResult, nil
}

type fakeLargeWorkRepo struct {
	listQuery largeworkrepository.ListQuery
	getQuery  largeworkrepository.GetQuery
	listItems []largeworkentity.LargeWorkItem
	item      *largeworkentity.LargeWorkItem
	listErr   error
	getErr    error
}

func (f *fakeLargeWorkRepo) List(_ context.Context, q largeworkrepository.ListQuery) ([]largeworkentity.LargeWorkItem, int64, error) {
	f.listQuery = q
	return f.listItems, int64(len(f.listItems)), f.listErr
}

func (f *fakeLargeWorkRepo) GetByID(_ context.Context, q largeworkrepository.GetQuery) (*largeworkentity.LargeWorkItem, error) {
	f.getQuery = q
	return f.item, f.getErr
}

func (f *fakeLargeWorkRepo) Create(context.Context, largeworkrepository.CreateInput) (*largeworkentity.LargeWorkItem, error) {
	panic("unexpected Create call")
}

func (f *fakeLargeWorkRepo) Update(context.Context, largeworkrepository.UpdateInput) (*largeworkentity.LargeWorkItem, error) {
	panic("unexpected Update call")
}

func TestListSourcesUsesOwnTeamScopeAndMergesTeamAndMonthlySources(t *testing.T) {
	teamID := int64(7)
	participantTeamID := int64(8)
	workDate := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	teamRepo := &fakeTeamPlanRepo{listResult: []teamplanentity.TeamPlan{{ID: 101, TeamID: teamID, Title: "Patrol feeder A", StartDate: ptrTime(workDate), LocationText: "Bang Khen", Notes: testStringPtr("prepare notice"), Team: &teamplanentity.TeamSummary{ID: teamID, Name: "Ops East"}}}}
	monthlyRepo := &fakeMonthlyRepo{
		periodResult: &monthlyentity.Entity{ID: 88, Year: 2026, Month: 6},
		filesResult:  []monthlyentity.PlanFileEntity{{ID: 202, MonthlyPlanID: 88, TeamID: &teamID, Description: testStringPtr("Monthly outage"), Destination: testStringPtr("Substation 9"), Remarks: testStringPtr("approved offline"), WorkStartDate: &workDate, TeamName: testStringPtr("Ops East")}},
	}
	largeRepo := &fakeLargeWorkRepo{listItems: []largeworkentity.LargeWorkItem{{
		ID:           303,
		OwnerTeamID:  teamID,
		Title:        "Big outage support",
		WorkType:     testStringPtr("large work"),
		StartDate:    workDate,
		EndDate:      &workDate,
		LocationText: "Substation 9",
		Status:       largeworkentity.LargeWorkStatusPlanned,
		Teams:        []largeworkentity.LargeWorkTeam{{ID: teamID, Name: "Ops East", Role: largeworkentity.LargeWorkTeamRoleOwner, ParticipantStatus: largeworkentity.LargeWorkParticipantStatusAssigned}, {ID: participantTeamID, Name: "Ops West", Role: largeworkentity.LargeWorkTeamRoleParticipant, ParticipantStatus: largeworkentity.LargeWorkParticipantStatusAcknowledged}},
	}}}
	svc := NewService(teamRepo, monthlyRepo, largeRepo)
	actor := taskActor(teamID, "user")

	out, err := svc.ListSources(context.Background(), actor, ListSourcesInput{WorkDate: &workDate})
	if err != nil {
		t.Fatalf("ListSources returned error: %v", err)
	}
	if teamRepo.listQuery.TeamID == nil || *teamRepo.listQuery.TeamID != teamID {
		t.Fatalf("team query = %#v, want team %d", teamRepo.listQuery.TeamID, teamID)
	}
	if teamRepo.listQuery.From.Format("2006-01-02") != "2026-06-04" || teamRepo.listQuery.To.Format("2006-01-02") != "2026-06-04" {
		t.Fatalf("team query date range = %+v", teamRepo.listQuery)
	}
	if monthlyRepo.listQuery.MonthlyPlanID != 88 || monthlyRepo.listQuery.TeamID != teamID {
		t.Fatalf("monthly query = %+v, want period 88 team 7", monthlyRepo.listQuery)
	}
	if largeRepo.listQuery.TeamID == nil || *largeRepo.listQuery.TeamID != teamID {
		t.Fatalf("large work query team id = %#v, want %d", largeRepo.listQuery.TeamID, teamID)
	}
	if len(out.Items) != 3 {
		t.Fatalf("items len = %d, want 3", len(out.Items))
	}
	if out.Items[0].SourceType != SourceTypeLargeWork && out.Items[1].SourceType != SourceTypeLargeWork && out.Items[2].SourceType != SourceTypeLargeWork {
		t.Fatalf("expected large work source in %#v", out.Items)
	}
	lw := findSourceCandidate(out.Items, SourceTypeLargeWork)
	if lw == nil || lw.SourceID != 303 {
		t.Fatalf("large work candidate = %#v, want source 303", lw)
	}
	if lw.TeamID == nil || *lw.TeamID != teamID {
		t.Fatalf("large work team id = %#v, want %d", lw.TeamID, teamID)
	}
	if lw.TeamName == nil || *lw.TeamName != "Ops East" {
		t.Fatalf("large work team name = %#v, want Ops East", lw.TeamName)
	}
}

func TestFromPlanPrefillTeamPlanMapsSelectedDateTeamFeederAndDetail(t *testing.T) {
	teamID := int64(7)
	feederID := int64(33)
	teamRepo := &fakeTeamPlanRepo{getByIDResult: &teamplanentity.TeamPlan{
		ID:              101,
		TeamID:          teamID,
		CreatedByUserID: 42,
		Title:           "Patrol feeder A",
		StartDate:       ptrTime(time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)),
		LocationText:    "Bang Khen",
		FeederID:        &feederID,
		Notes:           testStringPtr("carry spares"),
		Team:            &teamplanentity.TeamSummary{ID: teamID, Name: "Ops East"},
	}}
	svc := NewService(teamRepo, &fakeMonthlyRepo{}, &fakeLargeWorkRepo{})
	actor := taskActor(teamID, "user")
	selected := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)

	out, err := svc.FromPlan(context.Background(), actor, FromPlanInput{SourceType: SourceTypeTeamPlan, SourceID: 101, WorkDate: &selected})
	if err != nil {
		t.Fatalf("FromPlan returned error: %v", err)
	}
	if teamRepo.getByIDQuery.ID != 101 {
		t.Fatalf("captured query = %+v, want source id 101", teamRepo.getByIDQuery)
	}
	if out.Source.SourceType != SourceTypeTeamPlan || out.Source.SourceID != 101 {
		t.Fatalf("source = %#v, want team plan 101", out.Source)
	}
	if out.Prefill.WorkDate != "2026-06-04" {
		t.Fatalf("workDate = %q, want 2026-06-04", out.Prefill.WorkDate)
	}
	if out.Prefill.TeamID == nil || *out.Prefill.TeamID != teamID {
		t.Fatalf("teamId = %#v, want %d", out.Prefill.TeamID, teamID)
	}
	if out.Prefill.FeederID == nil || *out.Prefill.FeederID != feederID {
		t.Fatalf("feederId = %#v, want %d", out.Prefill.FeederID, feederID)
	}
	if out.Prefill.Detail == nil || *out.Prefill.Detail == "" {
		t.Fatal("detail should be populated")
	}
	if len(out.Warnings) == 0 {
		t.Fatal("expected warnings for incomplete daily report fields")
	}
}

func TestFromPlanPrefillMonthlyPlanUsesSelectedDateAndDestination(t *testing.T) {
	teamID := int64(9)
	selected := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	file := &monthlyentity.PlanFileEntity{
		ID:            202,
		MonthlyPlanID: 99,
		TeamID:        &teamID,
		UploadedByID:  12,
		FileName:      "monthly-plan.pdf",
		Description:   testStringPtr("Monthly outage"),
		WorkStartDate: &selected,
		Destination:   testStringPtr("Substation 9"),
		Remarks:       testStringPtr("approved offline"),
		TeamName:      testStringPtr("Ops South"),
	}
	svc := NewService(&fakeTeamPlanRepo{}, &fakeMonthlyRepo{fileResult: file}, &fakeLargeWorkRepo{})
	actor := taskActor(teamID, "team_lead")

	out, err := svc.FromPlan(context.Background(), actor, FromPlanInput{SourceType: SourceTypeMonthlyPlan, SourceID: 202})
	if err != nil {
		t.Fatalf("FromPlan returned error: %v", err)
	}
	if out.Source.SourceType != SourceTypeMonthlyPlan || out.Source.SourceID != 202 {
		t.Fatalf("source = %#v, want monthly plan 202", out.Source)
	}
	if out.Source.Title != "Monthly outage" {
		t.Fatalf("source title = %q, want Monthly outage", out.Source.Title)
	}
	if out.Prefill.WorkDate != "2026-06-05" {
		t.Fatalf("workDate = %q, want 2026-06-05", out.Prefill.WorkDate)
	}
	if out.Prefill.TeamID == nil || *out.Prefill.TeamID != teamID {
		t.Fatalf("teamId = %#v, want %d", out.Prefill.TeamID, teamID)
	}
	if out.Prefill.Detail == nil || *out.Prefill.Detail == "" {
		t.Fatal("detail should be populated")
	}
	if out.Prefill.FeederID != nil {
		t.Fatalf("feederId = %#v, want nil for monthly plan", out.Prefill.FeederID)
	}
}

func TestFromPlanRejectsForbiddenAndUnsupportedSources(t *testing.T) {
	teamID := int64(7)
	otherTeamID := int64(8)
	teamRepo := &fakeTeamPlanRepo{getByIDResult: &teamplanentity.TeamPlan{ID: 101, TeamID: otherTeamID, Title: "Other team", StartDate: ptrTime(time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)), LocationText: "X"}}
	largeRepo := &fakeLargeWorkRepo{item: &largeworkentity.LargeWorkItem{
		ID:           303,
		OwnerTeamID:  teamID,
		Title:        "Big outage support",
		StartDate:    time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC),
		LocationText: "Depot",
		Status:       largeworkentity.LargeWorkStatusPlanned,
		Teams:        []largeworkentity.LargeWorkTeam{{ID: teamID, Name: "Ops East", Role: largeworkentity.LargeWorkTeamRoleOwner, ParticipantStatus: largeworkentity.LargeWorkParticipantStatusAssigned}},
	}}
	svc := NewService(teamRepo, &fakeMonthlyRepo{}, largeRepo)
	actor := taskActor(teamID, "user")

	_, err := svc.FromPlan(context.Background(), actor, FromPlanInput{SourceType: SourceTypeTeamPlan, SourceID: 101})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("team plan cross-team error = %v, want forbidden", err)
	}

	out, err := svc.FromPlan(context.Background(), actor, FromPlanInput{SourceType: SourceTypeLargeWork, SourceID: 303})
	if err != nil {
		t.Fatalf("large work source returned error: %v", err)
	}
	if out.Source.SourceType != SourceTypeLargeWork || out.Source.SourceID != 303 {
		t.Fatalf("large work source = %#v, want 303", out.Source)
	}
	if out.Prefill.TeamID == nil || *out.Prefill.TeamID != teamID {
		t.Fatalf("large work team id = %#v, want %d", out.Prefill.TeamID, teamID)
	}
	if out.Prefill.Detail == nil || *out.Prefill.Detail == "" {
		t.Fatal("large work detail should be populated")
	}
}

func taskActor(teamID int64, role string) taskentity.Actor {
	return taskentity.Actor{UserID: 42, Role: role, TeamID: &teamID}
}

func findSourceCandidate(items []SourceCandidate, sourceType string) *SourceCandidate {
	for i := range items {
		if items[i].SourceType == sourceType {
			return &items[i]
		}
	}
	return nil
}

func ptrTime(t time.Time) *time.Time { return &t }

func testStringPtr(v string) *string { s := v; return &s }
