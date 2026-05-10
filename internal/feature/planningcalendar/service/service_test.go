package service

import (
	"context"
	"testing"
	"time"

	largeworkentity "backend-hotlines3/internal/feature/largework/entity"
	largeworkrepository "backend-hotlines3/internal/feature/largework/repository"
	monthlyentity "backend-hotlines3/internal/feature/monthlyplan/entity"
	monthlyrepository "backend-hotlines3/internal/feature/monthlyplan/repository"
	monthlyservice "backend-hotlines3/internal/feature/monthlyplan/service"
	teamplanentity "backend-hotlines3/internal/feature/teamplan/entity"
	teamplanrepository "backend-hotlines3/internal/feature/teamplan/repository"
)

type fakeTeamPlanRepo struct {
	calls     int
	lastQuery teamplanrepository.ListQuery
	items     []teamplanentity.TeamPlan
	err       error
}

func (f *fakeTeamPlanRepo) ListAll(ctx context.Context, query teamplanrepository.ListQuery) ([]teamplanentity.TeamPlan, error) {
	f.calls++
	f.lastQuery = query
	return f.items, f.err
}

type fakeMonthlyRepo struct {
	periodCalls   int
	fileCalls     int
	settingsCalls int
	period        *monthlyentity.Entity
	files         []monthlyentity.PlanFileEntity
	settings      *monthlyentity.SettingsEntity
	err           error
}

func (f *fakeMonthlyRepo) FindPeriod(ctx context.Context, q monthlyrepository.PeriodFindQuery) (*monthlyentity.Entity, error) {
	f.periodCalls++
	return f.period, f.err
}

func (f *fakeMonthlyRepo) GetPeriodByID(ctx context.Context, id int64) (*monthlyentity.Entity, error) {
	panic("unexpected GetPeriodByID call")
}

func (f *fakeMonthlyRepo) FindOrCreatePeriod(ctx context.Context, year, month int) (*monthlyentity.Entity, error) {
	panic("unexpected FindOrCreatePeriod call")
}

func (f *fakeMonthlyRepo) FindOrCreatePeriodsForYear(ctx context.Context, year int) ([]monthlyentity.Entity, error) {
	panic("unexpected FindOrCreatePeriodsForYear call")
}

func (f *fakeMonthlyRepo) GetOrCreateSettings(ctx context.Context) (*monthlyentity.SettingsEntity, error) {
	f.settingsCalls++
	return f.settings, f.err
}

func (f *fakeMonthlyRepo) UpdateSettings(ctx context.Context, s *monthlyentity.SettingsEntity) error {
	panic("unexpected UpdateSettings call")
}

func (f *fakeMonthlyRepo) ListFiles(ctx context.Context, q monthlyrepository.PlanFileListQuery) ([]monthlyentity.PlanFileEntity, error) {
	f.fileCalls++
	return f.files, f.err
}

func (f *fakeMonthlyRepo) ListFilesByPlanIDs(ctx context.Context, q monthlyrepository.PlanFileBatchListQuery) (map[int64][]monthlyentity.PlanFileEntity, error) {
	panic("unexpected ListFilesByPlanIDs call")
}

func (f *fakeMonthlyRepo) GetFileByID(ctx context.Context, id int64) (*monthlyentity.PlanFileEntity, error) {
	panic("unexpected GetFileByID call")
}

func (f *fakeMonthlyRepo) CreateFile(ctx context.Context, input monthlyrepository.PlanFileCreateInput) (*monthlyentity.PlanFileEntity, error) {
	panic("unexpected CreateFile call")
}

func (f *fakeMonthlyRepo) SoftDeleteFile(ctx context.Context, input monthlyrepository.PlanFileSoftDeleteInput) error {
	panic("unexpected SoftDeleteFile call")
}

func (f *fakeMonthlyRepo) RestoreFile(ctx context.Context, input monthlyrepository.PlanFileRestoreInput) (*monthlyentity.PlanFileEntity, error) {
	panic("unexpected RestoreFile call")
}

func (f *fakeMonthlyRepo) HardDeleteFile(ctx context.Context, input monthlyrepository.PlanFileHardDeleteInput) error {
	panic("unexpected HardDeleteFile call")
}

func (f *fakeMonthlyRepo) CreateFileSizeLog(ctx context.Context, input monthlyrepository.FileSizeLogCreateInput) error {
	panic("unexpected CreateFileSizeLog call")
}

func (f *fakeMonthlyRepo) CountFilesByTeam(ctx context.Context, q monthlyrepository.SubmissionQuery) ([]monthlyrepository.TeamCountRow, error) {
	panic("unexpected CountFilesByTeam call")
}

func (f *fakeMonthlyRepo) ListTeams(ctx context.Context) ([]monthlyrepository.TeamInfo, error) {
	panic("unexpected ListTeams call")
}

type fakeLargeWorkRepo struct {
	listCalls int
	lastQuery largeworkrepository.ListQuery
	items     []largeworkentity.LargeWorkItem
	err       error
}

func (f *fakeLargeWorkRepo) List(ctx context.Context, q largeworkrepository.ListQuery) ([]largeworkentity.LargeWorkItem, int64, error) {
	f.listCalls++
	f.lastQuery = q
	return f.items, int64(len(f.items)), f.err
}

func (f *fakeLargeWorkRepo) GetByID(ctx context.Context, q largeworkrepository.GetQuery) (*largeworkentity.LargeWorkItem, error) {
	panic("unexpected GetByID call")
}

func (f *fakeLargeWorkRepo) Create(ctx context.Context, input largeworkrepository.CreateInput) (*largeworkentity.LargeWorkItem, error) {
	panic("unexpected Create call")
}

func (f *fakeLargeWorkRepo) Update(ctx context.Context, input largeworkrepository.UpdateInput) (*largeworkentity.LargeWorkItem, error) {
	panic("unexpected Update call")
}

func TestGetMonthBuildsCalendarFromTeamAndMonthlyPlans(t *testing.T) {
	teamID := int64(77)
	teamName := "Ops East"
	creatorID := int64(10)
	createdBy := "lead"
	desc := "Outage work"
	loc := "Substation 9"
	remarks := "carry spares"
	start := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	fileStart := time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)
	fileEnd := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)
	period := &monthlyentity.Entity{ID: 900, Year: 2026, Month: 5, IsLocked: false, CreatedAt: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)}

	teamRepo := &fakeTeamPlanRepo{items: []teamplanentity.TeamPlan{{
		ID:              100,
		TeamID:          teamID,
		CreatedByUserID: creatorID,
		Title:           "Pole repair",
		StartDate:       start,
		EndDate:         &end,
		LocationText:    loc,
		Status:          teamplanentity.StatusPlanned,
		Team:            &teamplanentity.TeamSummary{ID: teamID, Name: teamName},
		CreatedBy:       &teamplanentity.UserSummary{ID: creatorID, Username: createdBy},
	}}}
	monthlyRepo := &fakeMonthlyRepo{
		period: period,
		files: []monthlyentity.PlanFileEntity{{
			ID:            200,
			MonthlyPlanID: period.ID,
			TeamID:        &teamID,
			UploadedByID:  11,
			FileName:      "monthly-plan.pdf",
			Description:   &desc,
			WorkStartDate: &fileStart,
			WorkEndDate:   &fileEnd,
			Destination:   &loc,
			Remarks:       &remarks,
			CreatedAt:     time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
			UpdatedAt:     time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		}},
		settings: &monthlyentity.SettingsEntity{ID: 1, LockDay: 28, AdminCanUploadAfterLock: true},
	}

	svc := NewService(teamRepo, monthlyRepo, &fakeLargeWorkRepo{}, func() time.Time {
		return time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	})

	actor := Actor{UserID: 10, Role: "user", TeamID: &teamID}
	result, err := svc.GetMonth(context.Background(), actor, 2026, 5)
	if err != nil {
		t.Fatalf("GetMonth returned error: %v", err)
	}
	if teamRepo.calls != 1 {
		t.Fatalf("expected one team plan query, got %d", teamRepo.calls)
	}
	if monthlyRepo.periodCalls != 1 || monthlyRepo.fileCalls != 1 || monthlyRepo.settingsCalls != 1 {
		t.Fatalf("expected one monthly period/files/settings lookup, got period=%d files=%d settings=%d", monthlyRepo.periodCalls, monthlyRepo.fileCalls, monthlyRepo.settingsCalls)
	}
	if result.Year != 2026 || result.Month != 5 {
		t.Fatalf("unexpected month projection header: %#v", result)
	}
	if got, want := len(result.Days), 31; got != want {
		t.Fatalf("expected %d days, got %d", want, got)
	}
	byDate := map[string]int{}
	for _, day := range result.Days {
		byDate[day.Date] = len(day.Items)
	}
	if got := byDate["2026-05-02"]; got != 1 {
		t.Fatalf("expected team plan on 2026-05-02, got %d items", got)
	}
	if got := byDate["2026-05-03"]; got != 2 {
		t.Fatalf("expected both plans on 2026-05-03, got %d items", got)
	}
	if got := byDate["2026-05-05"]; got != 1 {
		t.Fatalf("expected monthly plan on 2026-05-05, got %d items", got)
	}
	dayThree := result.Days[2].Items
	var teamItem, monthlyItem *CalendarItem
	for i := range dayThree {
		item := &dayThree[i]
		switch item.SourceType {
		case "team_plan":
			teamItem = item
		case "monthly_plan":
			monthlyItem = item
		}
	}
	if teamItem == nil || monthlyItem == nil {
		t.Fatalf("expected both team and monthly items on 2026-05-03: %#v", dayThree)
	}
	if !teamItem.Actions.CanEdit {
		t.Fatalf("expected team owner to edit own team plan")
	}
	if !monthlyItem.Actions.CanDownload {
		t.Fatalf("expected team member to download own monthly plan file")
	}
}

func TestGetMonthMonthlyPlanCanEditMatchesMonthlyPlanUploadWindow(t *testing.T) {
	teamID := int64(77)
	period := &monthlyentity.Entity{ID: 901, Year: 2026, Month: 6, IsLocked: false}
	settings := &monthlyentity.SettingsEntity{ID: 1, LockDay: 23, AdminCanUploadAfterLock: false}
	fileStart := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	repo := &fakeMonthlyRepo{
		period:   period,
		settings: settings,
		files: []monthlyentity.PlanFileEntity{{
			ID:            300,
			MonthlyPlanID: period.ID,
			TeamID:        &teamID,
			UploadedByID:  11,
			FileName:      "june-plan.pdf",
			WorkStartDate: &fileStart,
		}},
	}
	clock := func() time.Time { return time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC) }
	actor := Actor{UserID: 10, Role: "admin", TeamID: &teamID}

	monthlySvc := monthlyservice.NewService(repo, nil)
	monthlySvc.SetClockForTest(clock)
	monthlyCanUpload, monthlyDeadline, err := monthlySvc.CanUploadForPeriod(context.Background(), actor.toMonthlyPlanActor(), 2026, 6)
	if err != nil {
		t.Fatalf("CanUploadForPeriod returned error: %v", err)
	}
	if monthlyCanUpload {
		t.Fatalf("test setup expected June upload to be locked after previous-month deadline %s", monthlyDeadline)
	}

	calendarSvc := NewService(&fakeTeamPlanRepo{}, repo, &fakeLargeWorkRepo{}, clock)
	calendar, err := calendarSvc.GetMonth(context.Background(), actor, 2026, 6)
	if err != nil {
		t.Fatalf("GetMonth returned error: %v", err)
	}
	item := findCalendarItem(calendar.Days, "2026-06-02", "monthly_plan")
	if item == nil {
		t.Fatal("expected monthly plan item on 2026-06-02")
	}
	if item.Actions.CanEdit != monthlyCanUpload {
		t.Fatalf("calendar canEdit = %v, want monthly plan canUpload = %v (deadline %s)", item.Actions.CanEdit, monthlyCanUpload, monthlyDeadline)
	}
}

func TestGetMonthIncludesLargeWorkItemsWithTeamRefs(t *testing.T) {
	teamID := int64(77)
	participantTeamID := int64(88)
	start := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC)
	largeRepo := &fakeLargeWorkRepo{items: []largeworkentity.LargeWorkItem{{
		ID:              501,
		OwnerTeamID:     teamID,
		CreatedByUserID: 9,
		Title:           "งานระดมทีม เปลี่ยนอุปกรณ์หลัก",
		WorkType:        testStringPtr("major_maintenance"),
		StartDate:       start,
		EndDate:         &end,
		WorkTime:        testStringPtr("08:30"),
		LocationText:    "Station A feeder F01",
		Notes:           testStringPtr("Need bucket truck and safety observer"),
		Status:          largeworkentity.LargeWorkStatusPlanned,
		Teams: []largeworkentity.LargeWorkTeam{
			{ID: teamID, Name: "Ops East", Role: largeworkentity.LargeWorkTeamRoleOwner, ParticipantStatus: largeworkentity.LargeWorkParticipantStatusAssigned},
			{ID: participantTeamID, Name: "Ops West", Role: largeworkentity.LargeWorkTeamRoleParticipant, ParticipantStatus: largeworkentity.LargeWorkParticipantStatusAcknowledged},
		},
	}}}
	teamRepo := &fakeTeamPlanRepo{}
	monthlyRepo := &fakeMonthlyRepo{}
	svc := NewService(teamRepo, monthlyRepo, largeRepo, time.Now)
	actor := Actor{UserID: 10, Role: "user", TeamID: &participantTeamID}

	result, err := svc.GetMonth(context.Background(), actor, 2026, 6)
	if err != nil {
		t.Fatalf("GetMonth returned error: %v", err)
	}
	if largeRepo.listCalls != 1 {
		t.Fatalf("expected one large-work query, got %d", largeRepo.listCalls)
	}
	if largeRepo.lastQuery.TeamID == nil || *largeRepo.lastQuery.TeamID != participantTeamID {
		t.Fatalf("large-work query team scope = %#v, want %d", largeRepo.lastQuery.TeamID, participantTeamID)
	}
	for _, date := range []string{"2026-06-10", "2026-06-11", "2026-06-12"} {
		if got := countCalendarItems(result.Days, date, SourceTypeLargeWork); got != 1 {
			t.Fatalf("expected 1 large-work item on %s, got %d", date, got)
		}
	}
	item := findCalendarItem(result.Days, "2026-06-10", SourceTypeLargeWork)
	if item == nil {
		t.Fatal("expected large-work item on 2026-06-10")
	}
	if item.Team == nil || item.Team.ID != teamID || item.Team.Name != "Ops East" {
		t.Fatalf("lead team = %#v, want Ops East/%d", item.Team, teamID)
	}
	if len(item.Teams) != 2 {
		t.Fatalf("teams = %#v, want 2 team refs", item.Teams)
	}
	if item.Actions.CanEdit || item.Actions.CanDelete {
		t.Fatalf("participant team should not be able to edit/delete large work: %#v", item.Actions)
	}
}

func TestGetMonthDeniesOwnerTeamLeadLargeWorkEditActionsInMVP(t *testing.T) {
	teamID := int64(77)
	start := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	largeRepo := &fakeLargeWorkRepo{items: []largeworkentity.LargeWorkItem{{
		ID:              501,
		OwnerTeamID:     teamID,
		CreatedByUserID: 9,
		Title:           "งานระดมทีม เปลี่ยนอุปกรณ์หลัก",
		StartDate:       start,
		LocationText:    "Station A feeder F01",
		Status:          largeworkentity.LargeWorkStatusPlanned,
		Teams: []largeworkentity.LargeWorkTeam{{
			ID: teamID, Name: "Ops East", Role: largeworkentity.LargeWorkTeamRoleOwner, ParticipantStatus: largeworkentity.LargeWorkParticipantStatusAssigned,
		}},
	}}}
	svc := NewService(&fakeTeamPlanRepo{}, &fakeMonthlyRepo{}, largeRepo, time.Now)
	actor := Actor{UserID: 10, Role: "team_lead", TeamID: &teamID}

	result, err := svc.GetMonth(context.Background(), actor, 2026, 6)
	if err != nil {
		t.Fatalf("GetMonth returned error: %v", err)
	}
	item := findCalendarItem(result.Days, "2026-06-10", SourceTypeLargeWork)
	if item == nil {
		t.Fatal("expected large-work item on 2026-06-10")
	}
	if item.Actions.CanEdit || item.Actions.CanDelete {
		t.Fatalf("owner team lead should not be able to edit/delete large work in MVP: %#v", item.Actions)
	}
}

func testStringPtr(v string) *string {
	s := v
	return &s
}

func countCalendarItems(days []CalendarDay, date, sourceType string) int {
	count := 0
	for _, day := range days {
		if day.Date != date {
			continue
		}
		for _, item := range day.Items {
			if item.SourceType == sourceType {
				count++
			}
		}
	}
	return count
}

func findCalendarItem(days []CalendarDay, date, sourceType string) *CalendarItem {
	for i := range days {
		if days[i].Date != date {
			continue
		}
		for j := range days[i].Items {
			if days[i].Items[j].SourceType == sourceType {
				return &days[i].Items[j]
			}
		}
	}
	return nil
}

func BenchmarkGetMonthBuildsCalendarWithoutPerDayRepositoryCalls(b *testing.B) {
	teamRepo := &fakeTeamPlanRepo{}
	monthlyRepo := &fakeMonthlyRepo{settings: &monthlyentity.SettingsEntity{ID: 1, LockDay: 28, AdminCanUploadAfterLock: true}}
	svc := NewService(teamRepo, monthlyRepo, &fakeLargeWorkRepo{}, time.Now)
	actor := Actor{UserID: 1, Role: "super_admin"}
	for i := 0; i < b.N; i++ {
		if _, err := svc.GetMonth(context.Background(), actor, 2026, 5); err != nil {
			b.Fatal(err)
		}
	}
	if teamRepo.calls == 0 || monthlyRepo.fileCalls == 0 {
		b.Fatal("expected repository calls during benchmark")
	}
}
