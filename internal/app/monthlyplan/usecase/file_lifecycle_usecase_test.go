package usecase

import (
	"context"
	"fmt"
	"testing"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
	"backend-hotlines3/internal/port/outbound/repository"
)

// ── shared test helpers ──────────────────────────────────────────────────────

func int64Ptr(v int64) *int64 { return &v }
func strPtr(v string) *string { return &v }

// ── ListFiles tests ──────────────────────────────────────────────────────────

type fakeListRepo struct {
	files []mpdomain.PlanFileEntity
	query repository.PlanFileListQuery
	err   error
}

func (f *fakeListRepo) ListFiles(_ context.Context, q repository.PlanFileListQuery) ([]mpdomain.PlanFileEntity, error) {
	f.query = q
	if f.err != nil {
		return nil, f.err
	}
	return f.files, nil
}

func TestListFiles_AdminSeesAll(t *testing.T) {
	repo := &fakeListRepo{files: []mpdomain.PlanFileEntity{
		{ID: 1, FileName: "a.pdf"},
		{ID: 2, FileName: "b.pdf"},
	}}
	uc := NewListFilesUseCase(repo)
	out, err := uc.Execute(context.Background(), ListFilesInput{
		Actor:  adminActor(),
		PlanID: 1,
		TeamID: 0, // all
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Files) != 2 {
		t.Errorf("Files count = %d, want 2", len(out.Files))
	}
	if repo.query.TeamID != 0 {
		t.Errorf("query TeamID = %d, want 0 (all)", repo.query.TeamID)
	}
}

func TestListFiles_StaffSeesOwnTeam(t *testing.T) {
	repo := &fakeListRepo{}
	uc := NewListFilesUseCase(repo)
	_, err := uc.Execute(context.Background(), ListFilesInput{
		Actor:  staffActor(),
		PlanID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.query.TeamID != 20 {
		t.Errorf("query TeamID = %d, want 20 (staff's team)", repo.query.TeamID)
	}
}

// ── GetFile tests ────────────────────────────────────────────────────────────

type fakeGetFileRepo struct {
	file *mpdomain.PlanFileEntity
	err  error
}

func (f *fakeGetFileRepo) GetFileByID(_ context.Context, _ int64) (*mpdomain.PlanFileEntity, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.file, nil
}

func TestGetFile_InvalidID(t *testing.T) {
	uc := NewGetFileUseCase(&fakeGetFileRepo{})
	_, err := uc.Execute(context.Background(), GetFileInput{Actor: adminActor(), FileID: 0})
	if err != mpdomain.ErrInvalidFileID {
		t.Fatalf("expected ErrInvalidFileID, got: %v", err)
	}
}

func TestGetFile_NotFound(t *testing.T) {
	uc := NewGetFileUseCase(&fakeGetFileRepo{})
	_, err := uc.Execute(context.Background(), GetFileInput{Actor: adminActor(), FileID: 99})
	if err != mpdomain.ErrFileNotFound {
		t.Fatalf("expected ErrFileNotFound, got: %v", err)
	}
}

func TestGetFile_Success(t *testing.T) {
	uc := NewGetFileUseCase(&fakeGetFileRepo{
		file: &mpdomain.PlanFileEntity{ID: 1, FileName: "plan.pdf"},
	})
	out, err := uc.Execute(context.Background(), GetFileInput{Actor: adminActor(), FileID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.File.FileName != "plan.pdf" {
		t.Errorf("FileName = %q, want %q", out.File.FileName, "plan.pdf")
	}
}

func TestGetFile_StaffAccessOtherTeam(t *testing.T) {
	uc := NewGetFileUseCase(&fakeGetFileRepo{
		file: &mpdomain.PlanFileEntity{ID: 1, TeamID: int64Ptr(99), IsMasterPlan: false},
	})
	_, err := uc.Execute(context.Background(), GetFileInput{Actor: staffActor(), FileID: 1})
	if err != mpdomain.ErrForbiddenAction {
		t.Fatalf("expected ErrForbiddenAction, got: %v", err)
	}
}

func TestGetFile_StaffAccessMasterPlan(t *testing.T) {
	uc := NewGetFileUseCase(&fakeGetFileRepo{
		file: &mpdomain.PlanFileEntity{ID: 1, IsMasterPlan: true},
	})
	out, err := uc.Execute(context.Background(), GetFileInput{Actor: staffActor(), FileID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.File.ID != 1 {
		t.Errorf("File.ID = %d, want 1", out.File.ID)
	}
}

// ── SoftDeleteFile tests ─────────────────────────────────────────────────────

type fakeSoftDeleteRepo struct {
	file    *mpdomain.PlanFileEntity
	deleted bool
	err     error
}

func (f *fakeSoftDeleteRepo) GetFileByID(_ context.Context, _ int64) (*mpdomain.PlanFileEntity, error) {
	return f.file, f.err
}

func (f *fakeSoftDeleteRepo) SoftDeleteFile(_ context.Context, _ repository.PlanFileSoftDeleteInput) error {
	f.deleted = true
	return f.err
}

func TestSoftDelete_Success(t *testing.T) {
	repo := &fakeSoftDeleteRepo{file: &mpdomain.PlanFileEntity{ID: 1, TeamID: int64Ptr(10)}}
	uc := NewSoftDeleteFileUseCase(repo)
	_, err := uc.Execute(context.Background(), SoftDeleteFileInput{Actor: adminActor(), FileID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.deleted {
		t.Error("SoftDeleteFile should have been called")
	}
}

func TestSoftDelete_AlreadyDeleted(t *testing.T) {
	repo := &fakeSoftDeleteRepo{file: &mpdomain.PlanFileEntity{ID: 1, IsDeleted: true}}
	uc := NewSoftDeleteFileUseCase(repo)
	_, err := uc.Execute(context.Background(), SoftDeleteFileInput{Actor: adminActor(), FileID: 1})
	if err != mpdomain.ErrFileNotDeleted {
		t.Fatalf("expected ErrFileNotDeleted, got: %v", err)
	}
}

// ── RestoreFile tests ────────────────────────────────────────────────────────

type fakeRestoreRepo struct {
	file     *mpdomain.PlanFileEntity
	restored *mpdomain.PlanFileEntity
	err      error
}

func (f *fakeRestoreRepo) GetFileByID(_ context.Context, _ int64) (*mpdomain.PlanFileEntity, error) {
	return f.file, f.err
}

func (f *fakeRestoreRepo) RestoreFile(_ context.Context, _ repository.PlanFileRestoreInput) (*mpdomain.PlanFileEntity, error) {
	return f.restored, f.err
}

func TestRestoreFile_NotDeleted(t *testing.T) {
	repo := &fakeRestoreRepo{file: &mpdomain.PlanFileEntity{ID: 1, IsDeleted: false}}
	uc := NewRestoreFileUseCase(repo)
	_, err := uc.Execute(context.Background(), RestoreFileInput{Actor: adminActor(), FileID: 1})
	if err != mpdomain.ErrFileNotDeleted {
		t.Fatalf("expected ErrFileNotDeleted, got: %v", err)
	}
}

func TestRestoreFile_Success(t *testing.T) {
	repo := &fakeRestoreRepo{
		file:     &mpdomain.PlanFileEntity{ID: 1, IsDeleted: true, TeamID: int64Ptr(10)},
		restored: &mpdomain.PlanFileEntity{ID: 1, IsDeleted: false},
	}
	uc := NewRestoreFileUseCase(repo)
	out, err := uc.Execute(context.Background(), RestoreFileInput{Actor: adminActor(), FileID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.File.IsDeleted {
		t.Error("restored file should not be deleted")
	}
}

// ── HardDeleteFile tests ─────────────────────────────────────────────────────

type fakeHardDeleteRepo struct {
	file    *mpdomain.PlanFileEntity
	deleted bool
	err     error
}

func (f *fakeHardDeleteRepo) GetFileByID(_ context.Context, _ int64) (*mpdomain.PlanFileEntity, error) {
	return f.file, f.err
}

func (f *fakeHardDeleteRepo) HardDeleteFile(_ context.Context, _ repository.PlanFileHardDeleteInput) error {
	f.deleted = true
	return f.err
}

func TestHardDelete_NonAdmin(t *testing.T) {
	uc := NewHardDeleteFileUseCase(&fakeHardDeleteRepo{}, &fakeObjectStorage{})
	_, err := uc.Execute(context.Background(), HardDeleteFileInput{Actor: staffActor(), FileID: 1})
	if err != mpdomain.ErrForbiddenAction {
		t.Fatalf("expected ErrForbiddenAction, got: %v", err)
	}
}

func TestHardDelete_Success(t *testing.T) {
	repo := &fakeHardDeleteRepo{file: &mpdomain.PlanFileEntity{ID: 1, FileKey: "monthly-plans/key.pdf"}}
	s := &fakeObjectStorage{}
	uc := NewHardDeleteFileUseCase(repo, s)
	_, err := uc.Execute(context.Background(), HardDeleteFileInput{Actor: adminActor(), FileID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.deleted {
		t.Error("HardDeleteFile should have been called")
	}
	if s.deletedKey != "monthly-plans/key.pdf" {
		t.Errorf("storage deleted key = %q, want %q", s.deletedKey, "monthly-plans/key.pdf")
	}
}

// ── GetSubmissionStatus tests ────────────────────────────────────────────────

type fakeSubmissionRepo struct {
	settings *mpdomain.SettingsEntity
	period   *mpdomain.Entity
	teams    []repository.TeamInfo
	counts   []repository.TeamCountRow
	err      error
}

func (f *fakeSubmissionRepo) GetOrCreateSettings(_ context.Context) (*mpdomain.SettingsEntity, error) {
	if f.settings == nil {
		return &mpdomain.SettingsEntity{LockDay: 23}, nil
	}
	return f.settings, nil
}

func (f *fakeSubmissionRepo) FindOrCreatePeriod(_ context.Context, _, _ int) (*mpdomain.Entity, error) {
	if f.period == nil {
		return &mpdomain.Entity{ID: 1, Year: 2026, Month: 5}, nil
	}
	return f.period, nil
}

func (f *fakeSubmissionRepo) CountFilesByTeam(_ context.Context, _ repository.SubmissionQuery) ([]repository.TeamCountRow, error) {
	return f.counts, f.err
}

func (f *fakeSubmissionRepo) ListTeams(_ context.Context) ([]repository.TeamInfo, error) {
	return f.teams, f.err
}

func TestSubmissionStatus_NonAdmin(t *testing.T) {
	uc := NewGetSubmissionStatusUseCase(&fakeSubmissionRepo{})
	_, err := uc.Execute(context.Background(), GetSubmissionStatusInput{Actor: staffActor(), PlanID: 1})
	if err != mpdomain.ErrForbiddenAction {
		t.Fatalf("expected ErrForbiddenAction, got: %v", err)
	}
}

func TestSubmissionStatus_Success(t *testing.T) {
	repo := &fakeSubmissionRepo{
		teams: []repository.TeamInfo{
			{ID: 1, Name: "Team A"},
			{ID: 2, Name: "Team B"},
		},
		counts: []repository.TeamCountRow{
			{TeamID: 1, Count: 3},
		},
	}
	uc := NewGetSubmissionStatusUseCase(repo)
	out, err := uc.Execute(context.Background(), GetSubmissionStatusInput{Actor: adminActor(), PlanID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Teams) != 2 {
		t.Fatalf("Teams count = %d, want 2", len(out.Teams))
	}
	if out.Teams[0].Status != "submitted" {
		t.Errorf("Team A status = %q, want %q", out.Teams[0].Status, "submitted")
	}
	if out.Teams[0].FileCount != 3 {
		t.Errorf("Team A fileCount = %d, want 3", out.Teams[0].FileCount)
	}
	if out.Teams[1].Status != "pending" {
		t.Errorf("Team B status = %q, want %q", out.Teams[1].Status, "pending")
	}
}

// ── isAllowedType test ───────────────────────────────────────────────────────

func TestIsAllowedType(t *testing.T) {
	allowed := []string{"application/pdf", "image/png"}
	if !isAllowedType("application/pdf", allowed) {
		t.Error("pdf should be allowed")
	}
	if isAllowedType("application/exe", allowed) {
		t.Error("exe should not be allowed")
	}
}

// Suppress unused import warning
var _ = fmt.Sprintf
