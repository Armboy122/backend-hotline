package usecase

import (
	"context"
	"testing"
	"time"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
	"backend-hotlines3/internal/port/outbound/repository"
	"backend-hotlines3/internal/port/outbound/storage"
)

// ── fakes ────────────────────────────────────────────────────────────────────

type fakeUploadRepo struct {
	settings  *mpdomain.SettingsEntity
	created   *repository.PlanFileCreateInput
	logCalled bool
	createErr error
}

func (f *fakeUploadRepo) GetOrCreateSettings(_ context.Context) (*mpdomain.SettingsEntity, error) {
	if f.settings == nil {
		return &mpdomain.SettingsEntity{
			AllowedFileTypes: []string{"application/pdf", "image/png"},
		}, nil
	}
	return f.settings, nil
}

func (f *fakeUploadRepo) CreateFile(_ context.Context, input repository.PlanFileCreateInput) (*mpdomain.PlanFileEntity, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.created = &input
	return &mpdomain.PlanFileEntity{
		ID:            1,
		MonthlyPlanID: input.MonthlyPlanID,
		TeamID:        input.TeamID,
		UploadedByID:  input.UploadedByID,
		FileKey:       input.FileKey,
		FileURL:       input.FileURL,
		FileName:      input.FileName,
		FileSizeBytes: input.FileSizeBytes,
		IsMasterPlan:  input.IsMasterPlan,
	}, nil
}

func (f *fakeUploadRepo) CreateFileSizeLog(_ context.Context, _ repository.FileSizeLogCreateInput) error {
	f.logCalled = true
	return nil
}

type fakeObjectStorage struct {
	presignResult *storage.PresignResult
	presignErr    error
	deletedKey    string
}

func (f *fakeObjectStorage) PresignUpload(_ context.Context, _, _ string, _ time.Duration) (*storage.PresignResult, error) {
	if f.presignErr != nil {
		return nil, f.presignErr
	}
	if f.presignResult != nil {
		return f.presignResult, nil
	}
	return &storage.PresignResult{
		UploadURL: "https://upload.example.com/key",
		FileURL:   "https://cdn.example.com/key",
		FileKey:   "monthly-plans/2026/05/1-plan.pdf",
	}, nil
}

func (f *fakeObjectStorage) PresignDownload(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "https://download.example.com/key", nil
}

func (f *fakeObjectStorage) DeleteObject(_ context.Context, fileKey string) error {
	f.deletedKey = fileKey
	return nil
}

func (f *fakeObjectStorage) PublicURL(fileKey string) string {
	return "https://cdn.example.com/" + fileKey
}

// ── PresignUpload tests ──────────────────────────────────────────────────────

func adminActor() mpdomain.Actor {
	teamID := int64(10)
	return mpdomain.Actor{UserID: 1, Role: "admin", TeamID: &teamID}
}

func staffActor() mpdomain.Actor {
	teamID := int64(20)
	return mpdomain.Actor{UserID: 2, Role: "staff", TeamID: &teamID}
}

func TestPresignUpload_InvalidFileType(t *testing.T) {
	uc := NewPresignUploadUseCase(&fakeUploadRepo{}, &fakeObjectStorage{})
	_, err := uc.Execute(context.Background(), PresignUploadInput{
		Actor:    adminActor(),
		PlanID:   1,
		FileName: "malware.exe",
		FileType: "application/x-msdownload",
	})
	if err != mpdomain.ErrInvalidFileType {
		t.Fatalf("expected ErrInvalidFileType, got: %v", err)
	}
}

func TestPresignUpload_MasterPlanNonAdmin(t *testing.T) {
	uc := NewPresignUploadUseCase(&fakeUploadRepo{}, &fakeObjectStorage{})
	_, err := uc.Execute(context.Background(), PresignUploadInput{
		Actor:    staffActor(),
		PlanID:   1,
		FileName: "master.pdf",
		FileType: "application/pdf",
		IsMaster: true,
	})
	if err != mpdomain.ErrForbiddenAction {
		t.Fatalf("expected ErrForbiddenAction, got: %v", err)
	}
}

func TestPresignUpload_StaffForOtherTeam(t *testing.T) {
	uc := NewPresignUploadUseCase(&fakeUploadRepo{}, &fakeObjectStorage{})
	otherTeam := int64(99)
	_, err := uc.Execute(context.Background(), PresignUploadInput{
		Actor:    staffActor(),
		PlanID:   1,
		FileName: "plan.pdf",
		FileType: "application/pdf",
		TeamID:   &otherTeam,
	})
	if err != mpdomain.ErrForbiddenAction {
		t.Fatalf("expected ErrForbiddenAction, got: %v", err)
	}
}

func TestPresignUpload_Success(t *testing.T) {
	uc := NewPresignUploadUseCase(&fakeUploadRepo{}, &fakeObjectStorage{})
	out, err := uc.Execute(context.Background(), PresignUploadInput{
		Actor:    adminActor(),
		PlanID:   1,
		FileName: "plan.pdf",
		FileType: "application/pdf",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.UploadURL == "" {
		t.Error("UploadURL should not be empty")
	}
	if out.FileKey == "" {
		t.Error("FileKey should not be empty")
	}
}

func TestPresignUpload_StorageError(t *testing.T) {
	uc := NewPresignUploadUseCase(&fakeUploadRepo{}, &fakeObjectStorage{
		presignErr: mpdomain.ErrSettingsLoadFail, // any error
	})
	_, err := uc.Execute(context.Background(), PresignUploadInput{
		Actor:    adminActor(),
		PlanID:   1,
		FileName: "plan.pdf",
		FileType: "application/pdf",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── ConfirmUpload tests ──────────────────────────────────────────────────────

func TestConfirmUpload_MasterPlanNonAdmin(t *testing.T) {
	uc := NewConfirmUploadUseCase(&fakeUploadRepo{})
	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		Actor:        staffActor(),
		PlanID:       1,
		FileKey:      "key",
		FileURL:      "url",
		FileName:     "master.pdf",
		FileSizeBytes: 1024,
		IsMasterPlan: true,
	})
	if err != mpdomain.ErrForbiddenAction {
		t.Fatalf("expected ErrForbiddenAction, got: %v", err)
	}
}

func TestConfirmUpload_Success(t *testing.T) {
	repo := &fakeUploadRepo{}
	uc := NewConfirmUploadUseCase(repo)
	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		Actor:         adminActor(),
		PlanID:        1,
		FileKey:       "monthly-plans/2026/05/1-plan.pdf",
		FileURL:       "https://cdn.example.com/key",
		FileName:      "plan.pdf",
		FileSizeBytes: 2048,
		IsMasterPlan:  false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.File.ID != 1 {
		t.Errorf("File.ID = %d, want 1", out.File.ID)
	}
	if !repo.logCalled {
		t.Error("CreateFileSizeLog should have been called")
	}
	if repo.created == nil {
		t.Fatal("CreateFile should have been called")
	}
	if repo.created.FileSizeBytes != 2048 {
		t.Errorf("FileSizeBytes = %d, want 2048", repo.created.FileSizeBytes)
	}
}

func TestConfirmUpload_StaffOwnTeam(t *testing.T) {
	repo := &fakeUploadRepo{}
	uc := NewConfirmUploadUseCase(repo)
	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		Actor:         staffActor(),
		PlanID:        1,
		FileKey:       "key",
		FileURL:       "url",
		FileName:      "team-plan.pdf",
		FileSizeBytes: 512,
		IsMasterPlan:  false,
		// TeamID nil → uses actor's team
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.File.ID != 1 {
		t.Errorf("File.ID = %d, want 1", out.File.ID)
	}
	if repo.created.TeamID == nil || *repo.created.TeamID != 20 {
		t.Errorf("TeamID should be 20 (actor's team)")
	}
}

func TestConfirmUpload_CreateFail(t *testing.T) {
	uc := NewConfirmUploadUseCase(&fakeUploadRepo{
		createErr: mpdomain.ErrPeriodNotFound,
	})
	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		Actor:         adminActor(),
		PlanID:        1,
		FileKey:       "key",
		FileURL:       "url",
		FileName:      "plan.pdf",
		FileSizeBytes: 1024,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
