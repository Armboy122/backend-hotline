package service

import (
	"context"
	"errors"
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

func TestUpdateSettingsRequiresAdminAndAppliesPatch(t *testing.T) {
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

	_, err := svc.UpdateSettings(ctx, entity.Actor{UserID: 1, Role: "staff"}, SettingsPatch{})
	if !errors.Is(err, entity.ErrForbiddenAction) {
		t.Fatalf("expected forbidden for non-admin, got %v", err)
	}

	lockDay := 15
	maxMB := 25
	adminAfterLock := false
	updated, err := svc.UpdateSettings(ctx, entity.Actor{UserID: 2, Role: "admin"}, SettingsPatch{
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

	_, err := svc.PresignUpload(ctx, entity.Actor{UserID: 10, Role: "staff", TeamID: &teamOne}, 99, "plan.csv", "text/csv", false, &teamOne)
	if !errors.Is(err, entity.ErrInvalidFileType) {
		t.Fatalf("expected invalid file type, got %v", err)
	}

	_, err = svc.PresignUpload(ctx, entity.Actor{UserID: 10, Role: "staff", TeamID: &teamOne}, 99, "plan.pdf", "application/pdf", false, &teamTwo)
	if !errors.Is(err, entity.ErrForbiddenAction) {
		t.Fatalf("expected forbidden for another team, got %v", err)
	}

	result, err := svc.PresignUpload(ctx, entity.Actor{UserID: 10, Role: "staff", TeamID: &teamOne}, 99, "folder\\plan.pdf", "application/pdf", false, nil)
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

func TestConfirmUploadScopesNonAdminToOwnTeam(t *testing.T) {
	ctx := context.Background()
	actorTeam := int64(7)
	otherTeam := int64(8)
	repo := &fakeRepo{}
	svc := NewService(repo, &fakeStorage{})

	_, err := svc.ConfirmUpload(ctx, entity.Actor{UserID: 20, Role: "staff", TeamID: &actorTeam}, repository.PlanFileCreateInput{
		MonthlyPlanID: 5,
		TeamID:        &otherTeam,
		FileKey:       "key",
		FileURL:       "url",
		FileName:      "plan.pdf",
	})
	if !errors.Is(err, entity.ErrForbiddenAction) {
		t.Fatalf("expected forbidden for another team, got %v", err)
	}

	file, err := svc.ConfirmUpload(ctx, entity.Actor{UserID: 20, Role: "staff", TeamID: &actorTeam}, repository.PlanFileCreateInput{
		MonthlyPlanID: 5,
		FileKey:       "key",
		FileURL:       "url",
		FileName:      "plan.pdf",
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
	if !repo.createdSizeLog {
		t.Fatalf("expected file size log creation")
	}
}

func TestListFilesScopesNonAdminToActorTeam(t *testing.T) {
	ctx := context.Background()
	actorTeam := int64(3)
	repo := &fakeRepo{}
	svc := NewService(repo, &fakeStorage{})

	_, err := svc.ListFiles(ctx, entity.Actor{UserID: 30, Role: "staff", TeamID: &actorTeam}, 12, 99, "q")
	if err != nil {
		t.Fatalf("list files: %v", err)
	}
	if repo.lastListQuery.TeamID != actorTeam {
		t.Fatalf("expected actor team scope, got %d", repo.lastListQuery.TeamID)
	}

	_, err = svc.ListFiles(ctx, entity.Actor{UserID: 31, Role: "admin"}, 12, 99, "q")
	if err != nil {
		t.Fatalf("admin list files: %v", err)
	}
	if repo.lastListQuery.TeamID != 99 {
		t.Fatalf("expected requested team scope for admin, got %d", repo.lastListQuery.TeamID)
	}
}

type fakeRepo struct {
	settings        *entity.SettingsEntity
	updatedSettings bool
	createdSizeLog  bool
	lastListQuery   repository.PlanFileListQuery
}

func (r *fakeRepo) FindPeriod(context.Context, repository.PeriodFindQuery) (*entity.Entity, error) {
	return nil, nil
}

func (r *fakeRepo) FindOrCreatePeriod(_ context.Context, year, month int) (*entity.Entity, error) {
	return &entity.Entity{ID: 1, Year: year, Month: month}, nil
}

func (r *fakeRepo) GetOrCreateSettings(context.Context) (*entity.SettingsEntity, error) {
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
	r.lastListQuery = q
	return []entity.PlanFileEntity{}, nil
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
	return nil, nil
}

func (r *fakeRepo) ListTeams(context.Context) ([]repository.TeamInfo, error) {
	return nil, nil
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
