package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"backend-hotlines3/internal/feature/monthlyplan/entity"
	"backend-hotlines3/internal/feature/monthlyplan/repository"
)

func TestEnsurePeriodRejectsInvalidInput(t *testing.T) {
	svc := NewService(&fakeRepo{}, &fakeStorage{})

	tests := []struct {
		name  string
		year  int
		month int
	}{
		{name: "year too early", year: 1999, month: 1},
		{name: "year too late", year: 2101, month: 1},
		{name: "month too low", year: 2026, month: 0},
		{name: "month too high", year: 2026, month: 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.EnsurePeriod(context.Background(), tt.year, tt.month)
			if !errors.Is(err, entity.ErrInvalidPeriod) {
				t.Fatalf("expected ErrInvalidPeriod, got %v", err)
			}
		})
	}
}

func TestUpdateSettingsRequiresSuperAdminAndAppliesPatch(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{settings: &entity.SettingsEntity{
		LockDay:                 10,
		AutoCreateDay:           1,
		AutoCreateTarget:        "current",
		AllowedFileTypes:        []string{"application/pdf"},
		ReminderStartDay:        20,
		AdminCanUploadAfterLock: true,
	}}
	svc := NewService(repo, &fakeStorage{})

	_, err := svc.UpdateSettings(ctx, entity.Actor{UserID: 1, Role: "admin"}, SettingsPatch{})
	if !errors.Is(err, entity.ErrForbiddenAction) {
		t.Fatalf("expected forbidden for non-super-admin, got %v", err)
	}

	lockDay := 15
	maxMB := 25
	adminAfterLock := false
	updated, err := svc.UpdateSettings(ctx, entity.Actor{UserID: 2, Role: "super_admin"}, SettingsPatch{
		LockDay:                 &lockDay,
		AllowedFileTypes:        []string{"text/csv"},
		MaxFileSizeMB:           &maxMB,
		AdminCanUploadAfterLock: &adminAfterLock,
	})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if updated.LockDay != 15 || len(updated.AllowedFileTypes) != 1 || updated.AllowedFileTypes[0] != "text/csv" {
		t.Fatalf("patch was not applied: %+v", updated)
	}
	if updated.MaxFileSizeMB == nil || *updated.MaxFileSizeMB != 25 {
		t.Fatalf("expected max size patch, got %+v", updated.MaxFileSizeMB)
	}
	if updated.AdminCanUploadAfterLock {
		t.Fatalf("expected admin upload after lock to be false")
	}
	if !repo.updatedSettings {
		t.Fatalf("expected settings to be persisted")
	}
}

func TestPresignUploadEnforcesFileTypeAndTeamScope(t *testing.T) {
	ctx := context.Background()
	teamOne := int64(1)
	teamTwo := int64(2)
	repo := &fakeRepo{settings: &entity.SettingsEntity{AllowedFileTypes: []string{"application/pdf"}}}
	storage := &fakeStorage{}
	svc := NewService(repo, storage)
	svc.clock = func() time.Time { return time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC) }

	_, err := svc.PresignUpload(ctx, entity.Actor{UserID: 10, Role: "team_lead", TeamID: &teamOne}, 99, "plan.csv", "text/csv", false, &teamOne)
	if !errors.Is(err, entity.ErrInvalidFileType) {
		t.Fatalf("expected invalid file type, got %v", err)
	}

	_, err = svc.PresignUpload(ctx, entity.Actor{UserID: 10, Role: "team_lead", TeamID: &teamOne}, 99, "plan.pdf", "application/pdf", false, &teamTwo)
	if !errors.Is(err, entity.ErrForbiddenAction) {
		t.Fatalf("expected forbidden for another team, got %v", err)
	}

	result, err := svc.PresignUpload(ctx, entity.Actor{UserID: 10, Role: "team_lead", TeamID: &teamOne}, 99, "folder\\plan.pdf", "application/pdf", false, nil)
	if err != nil {
		t.Fatalf("presign upload: %v", err)
	}
	if result.UploadURL == "" || result.FileURL == "" || result.FileKey == "" {
		t.Fatalf("expected presign result, got %+v", result)
	}
	if storage.lastUploadKey != "monthly-plans/2026/04/99-folder_plan.pdf.pdf" {
		t.Fatalf("unexpected file key %q", storage.lastUploadKey)
	}
}

func TestConfirmUploadAllowsTeamMembersOnlyForOwnTeamAndKeepsPlanFields(t *testing.T) {
	ctx := context.Background()
	actorTeam := int64(7)
	otherTeam := int64(8)
	repo := &fakeRepo{}
	svc := NewService(repo, &fakeStorage{})

	_, err := svc.ConfirmUpload(ctx, entity.Actor{UserID: 20, Role: "user", TeamID: &actorTeam}, repository.PlanFileCreateInput{
		MonthlyPlanID: 5,
		TeamID:        &otherTeam,
		FileKey:       "key",
		FileURL:       "url",
		FileName:      "plan.pdf",
	})
	if !errors.Is(err, entity.ErrForbiddenAction) {
		t.Fatalf("expected user forbidden for another team, got %v", err)
	}

	_, err = svc.ConfirmUpload(ctx, entity.Actor{UserID: 20, Role: "team_lead", TeamID: &actorTeam}, repository.PlanFileCreateInput{
		MonthlyPlanID: 5,
		TeamID:        &otherTeam,
		FileKey:       "key",
		FileURL:       "url",
		FileName:      "plan.pdf",
	})
	if !errors.Is(err, entity.ErrForbiddenAction) {
		t.Fatalf("expected team lead forbidden for another team, got %v", err)
	}

	workStart := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)
	workEnd := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)
	destination := "Station A feeder 1"
	remarks := "after team planning meeting"
	file, err := svc.ConfirmUpload(ctx, entity.Actor{UserID: 20, Role: "user", TeamID: &actorTeam}, repository.PlanFileCreateInput{
		MonthlyPlanID: 5,
		FileKey:       "key",
		FileURL:       "url",
		FileName:      "plan.pdf",
		WorkStartDate: &workStart,
		WorkEndDate:   &workEnd,
		Destination:   &destination,
		Remarks:       &remarks,
	})
	if err != nil {
		t.Fatalf("confirm upload: %v", err)
	}
	if file.TeamID == nil || *file.TeamID != actorTeam {
		t.Fatalf("expected actor team to be assigned, got %+v", file.TeamID)
	}
	if file.UploadedByID != 20 {
		t.Fatalf("expected actor upload id, got %d", file.UploadedByID)
	}
	if file.WorkStartDate == nil || !file.WorkStartDate.Equal(workStart) || file.WorkEndDate == nil || !file.WorkEndDate.Equal(workEnd) {
		t.Fatalf("expected work date range to be preserved, got start=%v end=%v", file.WorkStartDate, file.WorkEndDate)
	}
	if file.Destination == nil || *file.Destination != destination || file.Remarks == nil || *file.Remarks != remarks {
		t.Fatalf("expected destination/remarks to be preserved, got destination=%v remarks=%v", file.Destination, file.Remarks)
	}
	if !repo.createdSizeLog {
		t.Fatalf("expected file size log creation")
	}
}

func TestListFilesScopesNonAdminToActorTeam(t *testing.T) {
	ctx := context.Background()
	actorTeam := int64(3)
	repo := &fakeRepo{}
	svc := NewService(repo, &fakeStorage{})

	_, err := svc.ListFiles(ctx, entity.Actor{UserID: 30, Role: "team_lead", TeamID: &actorTeam}, 12, 99, "q")
	if err != nil {
		t.Fatalf("list files: %v", err)
	}
	if repo.lastListQuery.TeamID != actorTeam {
		t.Fatalf("expected actor team scope, got %d", repo.lastListQuery.TeamID)
	}

	_, err = svc.ListFiles(ctx, entity.Actor{UserID: 31, Role: "super_admin"}, 12, 99, "q")
	if err != nil {
		t.Fatalf("super admin list files: %v", err)
	}
	if repo.lastListQuery.TeamID != 99 {
		t.Fatalf("expected requested team scope for super admin, got %d", repo.lastListQuery.TeamID)
	}
}

func TestCanUploadForPeriodUsesPreviousMonthLockDayAndSuperAdminOverride(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{settings: &entity.SettingsEntity{LockDay: 23, AdminCanUploadAfterLock: false}}
	svc := NewService(repo, &fakeStorage{})
	teamID := int64(7)

	svc.clock = func() time.Time { return time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC) }
	allowed, deadline, err := svc.CanUploadForPeriod(ctx, entity.Actor{UserID: 1, Role: "team_lead", TeamID: &teamID}, 2026, 6)
	if err != nil {
		t.Fatalf("can upload before deadline: %v", err)
	}
	if !allowed || deadline != "2026-05-23" {
		t.Fatalf("expected June 2026 upload allowed through 2026-05-23, got allowed=%v deadline=%q", allowed, deadline)
	}

	svc.clock = func() time.Time { return time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC) }
	allowed, deadline, err = svc.CanUploadForPeriod(ctx, entity.Actor{UserID: 2, Role: "team_lead", TeamID: &teamID}, 2026, 6)
	if err != nil {
		t.Fatalf("can upload after deadline: %v", err)
	}
	if allowed || deadline != "2026-05-23" {
		t.Fatalf("expected team lead blocked after deadline without override, got allowed=%v deadline=%q", allowed, deadline)
	}

	allowed, _, err = svc.CanUploadForPeriod(ctx, entity.Actor{UserID: 3, Role: "super_admin"}, 2026, 6)
	if err != nil {
		t.Fatalf("can upload after deadline as super_admin: %v", err)
	}
	if !allowed {
		t.Fatalf("expected super_admin to bypass upload lock")
	}
}

func TestGetSubmissionStatusUsesDeterministicDeadlineForNextMonth(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{
		settings:   &entity.SettingsEntity{LockDay: 20},
		periodByID: &entity.Entity{ID: 77, Year: 2026, Month: 6},
		teams:      []repository.TeamInfo{{ID: 1, Name: "Alpha"}, {ID: 2, Name: "Beta"}},
		counts:     []repository.TeamCountRow{{TeamID: 1, Count: 2}},
	}
	svc := NewService(repo, &fakeStorage{})
	svc.clock = func() time.Time { return time.Date(2026, 5, 21, 8, 0, 0, 0, time.UTC) }

	out, err := svc.GetSubmissionStatus(ctx, entity.Actor{UserID: 1, Role: "super_admin"}, 77)
	if err != nil {
		t.Fatalf("get submission status: %v", err)
	}
	if out.Deadline != "2026-05-20" {
		t.Fatalf("expected previous-month deadline, got %q", out.Deadline)
	}
	if len(out.Teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(out.Teams))
	}
	if out.Teams[0].Status != "submitted" {
		t.Fatalf("expected submitted team with files, got %+v", out.Teams[0])
	}
	if out.Teams[1].Status != "missed" {
		t.Fatalf("expected team without files to be missed after deadline, got %+v", out.Teams[1])
	}

	svc.clock = func() time.Time { return time.Date(2026, 5, 19, 8, 0, 0, 0, time.UTC) }
	teamLeadTeamID := int64(2)
	out, err = svc.GetSubmissionStatus(ctx, entity.Actor{UserID: 1, Role: "team_lead", TeamID: &teamLeadTeamID}, 77)
	if err != nil {
		t.Fatalf("get submission status before deadline: %v", err)
	}
	if len(out.Teams) != 1 || out.Teams[0].TeamID != teamLeadTeamID {
		t.Fatalf("expected team lead to see only own team, got %+v", out.Teams)
	}
	if out.Teams[0].Status != "pending" {
		t.Fatalf("expected team without files to be pending before deadline, got %+v", out.Teams[0])
	}
}

func TestGetSubmissionStatusScopesNonSuperAdminToOwnTeam(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{
		settings:   &entity.SettingsEntity{LockDay: 23},
		periodByID: &entity.Entity{ID: 88, Year: 2026, Month: 6},
		teams: []repository.TeamInfo{
			{ID: 7, Name: "Actor team"},
			{ID: 8, Name: "Other team"},
		},
		counts: []repository.TeamCountRow{{TeamID: 8, Count: 1}},
	}
	svc := NewService(repo, &fakeStorage{})
	svc.clock = func() time.Time { return time.Date(2026, 5, 10, 8, 0, 0, 0, time.UTC) }
	actorTeamID := int64(7)

	for _, role := range []string{"team_lead", "user"} {
		t.Run(role, func(t *testing.T) {
			out, err := svc.GetSubmissionStatus(ctx, entity.Actor{UserID: 40, Role: role, TeamID: &actorTeamID}, 88)
			if err != nil {
				t.Fatalf("get submission status: %v", err)
			}
			if len(out.Teams) != 1 {
				t.Fatalf("expected own team row only, got %d", len(out.Teams))
			}
			if out.Teams[0].TeamID != actorTeamID {
				t.Fatalf("expected actor team row, got %+v", out.Teams)
			}
		})
	}
}

func TestGetSubmissionStatusStartsIndependentRepositoryLookupsConcurrently(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{
		settings:       &entity.SettingsEntity{LockDay: 20},
		periodByID:     &entity.Entity{ID: 77, Year: 2026, Month: 6},
		teams:          []repository.TeamInfo{{ID: 1, Name: "Alpha"}},
		counts:         []repository.TeamCountRow{{TeamID: 1, Count: 1}},
		operationDelay: 35 * time.Millisecond,
	}
	svc := NewService(repo, &fakeStorage{})

	_, err := svc.GetSubmissionStatus(ctx, entity.Actor{UserID: 1, Role: "super_admin"}, 77)
	if err != nil {
		t.Fatalf("get submission status: %v", err)
	}

	starts := repo.operationStarts()
	for _, name := range []string{"settings", "period", "teams", "counts"} {
		if starts[name].IsZero() {
			t.Fatalf("%s lookup was not called; starts=%v", name, starts)
		}
	}
	minStart, maxStart := starts["settings"], starts["settings"]
	for _, start := range starts {
		if start.Before(minStart) {
			minStart = start
		}
		if start.After(maxStart) {
			maxStart = start
		}
	}
	if spread := maxStart.Sub(minStart); spread >= repo.operationDelay {
		t.Fatalf("independent status lookups started over %s; want under %s to avoid sequential DB round trips", spread, repo.operationDelay)
	}
}

func TestGetYearOverviewUsesBatchedPeriodAndFileLookups(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{
		settings: &entity.SettingsEntity{LockDay: 23},
		periodsForYear: []entity.Entity{
			{ID: 1, Year: 2026, Month: 1}, {ID: 2, Year: 2026, Month: 2},
			{ID: 3, Year: 2026, Month: 3}, {ID: 4, Year: 2026, Month: 4},
			{ID: 5, Year: 2026, Month: 5}, {ID: 6, Year: 2026, Month: 6},
			{ID: 7, Year: 2026, Month: 7}, {ID: 8, Year: 2026, Month: 8},
			{ID: 9, Year: 2026, Month: 9}, {ID: 10, Year: 2026, Month: 10},
			{ID: 11, Year: 2026, Month: 11}, {ID: 12, Year: 2026, Month: 12},
		},
		filesByPlan: map[int64][]entity.PlanFileEntity{
			6: {{ID: 66, MonthlyPlanID: 6, FileName: "june.pdf", IsMasterPlan: true}},
		},
	}
	svc := NewService(repo, &fakeStorage{})
	svc.clock = func() time.Time { return time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC) }

	out, err := svc.GetYearOverview(ctx, entity.Actor{UserID: 1, Role: "super_admin"}, 2026)
	if err != nil {
		t.Fatalf("get year overview: %v", err)
	}
	if len(out.Months) != 12 {
		t.Fatalf("months = %d, want 12", len(out.Months))
	}
	if repo.findOrCreatePeriodCalls != 0 || repo.listFilesCalls != 0 {
		t.Fatalf("year overview used per-month repository calls: periods=%d files=%d", repo.findOrCreatePeriodCalls, repo.listFilesCalls)
	}
	if repo.periodsForYearCalls != 1 || repo.filesByPlanCalls != 1 {
		t.Fatalf("batched calls = periods:%d files:%d, want 1/1", repo.periodsForYearCalls, repo.filesByPlanCalls)
	}
	if got := out.Months[5].Files; len(got) != 1 || got[0].FileName != "june.pdf" {
		t.Fatalf("june files = %+v, want batched file result", got)
	}
}

type fakeRepo struct {
	settings                *entity.SettingsEntity
	periodByID              *entity.Entity
	periodsForYear          []entity.Entity
	filesByPlan             map[int64][]entity.PlanFileEntity
	teams                   []repository.TeamInfo
	counts                  []repository.TeamCountRow
	updatedSettings         bool
	createdSizeLog          bool
	lastListQuery           repository.PlanFileListQuery
	lastBatchQuery          repository.PlanFileBatchListQuery
	findOrCreatePeriodCalls int
	periodsForYearCalls     int
	listFilesCalls          int
	filesByPlanCalls        int
	operationDelay          time.Duration
	mu                      sync.Mutex
	starts                  map[string]time.Time
}

func (r *fakeRepo) noteStart(name string) {
	if r.operationDelay == 0 {
		return
	}
	r.mu.Lock()
	if r.starts == nil {
		r.starts = make(map[string]time.Time)
	}
	r.starts[name] = time.Now()
	r.mu.Unlock()
	time.Sleep(r.operationDelay)
}

func (r *fakeRepo) operationStarts() map[string]time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]time.Time, len(r.starts))
	for k, v := range r.starts {
		out[k] = v
	}
	return out
}

func (r *fakeRepo) FindPeriod(context.Context, repository.PeriodFindQuery) (*entity.Entity, error) {
	return nil, nil
}

func (r *fakeRepo) GetPeriodByID(context.Context, int64) (*entity.Entity, error) {
	r.noteStart("period")
	return r.periodByID, nil
}

func (r *fakeRepo) FindOrCreatePeriod(_ context.Context, year, month int) (*entity.Entity, error) {
	r.findOrCreatePeriodCalls++
	return &entity.Entity{ID: 1, Year: year, Month: month}, nil
}

func (r *fakeRepo) FindOrCreatePeriodsForYear(_ context.Context, year int) ([]entity.Entity, error) {
	r.periodsForYearCalls++
	if r.periodsForYear != nil {
		return r.periodsForYear, nil
	}
	periods := make([]entity.Entity, 0, 12)
	for month := 1; month <= 12; month++ {
		periods = append(periods, entity.Entity{ID: int64(month), Year: year, Month: month})
	}
	return periods, nil
}

func (r *fakeRepo) GetOrCreateSettings(context.Context) (*entity.SettingsEntity, error) {
	r.noteStart("settings")
	if r.settings == nil {
		r.settings = &entity.SettingsEntity{AllowedFileTypes: []string{"application/pdf"}}
	}
	return r.settings, nil
}

func (r *fakeRepo) UpdateSettings(_ context.Context, s *entity.SettingsEntity) error {
	r.settings = s
	r.updatedSettings = true
	return nil
}

func (r *fakeRepo) ListFiles(_ context.Context, q repository.PlanFileListQuery) ([]entity.PlanFileEntity, error) {
	r.listFilesCalls++
	r.lastListQuery = q
	return []entity.PlanFileEntity{}, nil
}

func (r *fakeRepo) ListFilesByPlanIDs(_ context.Context, q repository.PlanFileBatchListQuery) (map[int64][]entity.PlanFileEntity, error) {
	r.filesByPlanCalls++
	r.lastBatchQuery = q
	if r.filesByPlan != nil {
		return r.filesByPlan, nil
	}
	return map[int64][]entity.PlanFileEntity{}, nil
}

func (r *fakeRepo) GetFileByID(context.Context, int64) (*entity.PlanFileEntity, error) {
	return nil, nil
}

func (r *fakeRepo) CreateFile(_ context.Context, input repository.PlanFileCreateInput) (*entity.PlanFileEntity, error) {
	return &entity.PlanFileEntity{
		ID:            42,
		MonthlyPlanID: input.MonthlyPlanID,
		TeamID:        input.TeamID,
		UploadedByID:  input.UploadedByID,
		FileKey:       input.FileKey,
		FileURL:       input.FileURL,
		FileName:      input.FileName,
		WorkStartDate: input.WorkStartDate,
		WorkEndDate:   input.WorkEndDate,
		Destination:   input.Destination,
		Remarks:       input.Remarks,
		IsMasterPlan:  input.IsMasterPlan,
	}, nil
}

func (r *fakeRepo) SoftDeleteFile(context.Context, repository.PlanFileSoftDeleteInput) error {
	return nil
}

func (r *fakeRepo) RestoreFile(context.Context, repository.PlanFileRestoreInput) (*entity.PlanFileEntity, error) {
	return nil, nil
}

func (r *fakeRepo) HardDeleteFile(context.Context, repository.PlanFileHardDeleteInput) error {
	return nil
}

func (r *fakeRepo) CreateFileSizeLog(context.Context, repository.FileSizeLogCreateInput) error {
	r.createdSizeLog = true
	return nil
}

func (r *fakeRepo) CountFilesByTeam(context.Context, repository.SubmissionQuery) ([]repository.TeamCountRow, error) {
	r.noteStart("counts")
	return r.counts, nil
}

func (r *fakeRepo) ListTeams(context.Context) ([]repository.TeamInfo, error) {
	r.noteStart("teams")
	return r.teams, nil
}

type fakeStorage struct {
	lastUploadKey string
}

func (s *fakeStorage) PresignUpload(_ context.Context, fileKey, contentType string, _ time.Duration) (*repository.PresignResult, error) {
	s.lastUploadKey = fileKey
	return &repository.PresignResult{
		UploadURL: "https://upload.example/" + contentType,
		FileURL:   "https://files.example/" + fileKey,
		FileKey:   fileKey,
	}, nil
}

func (s *fakeStorage) PresignDownload(context.Context, string, time.Duration) (string, error) {
	return "https://download.example", nil
}

func (s *fakeStorage) DeleteObject(context.Context, string) error {
	return nil
}

func (s *fakeStorage) PublicURL(fileKey string) string {
	return "https://files.example/" + fileKey
}
