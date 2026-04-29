package usecase

import (
	"context"
	"fmt"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
	"backend-hotlines3/internal/port/outbound/repository"
	"backend-hotlines3/internal/port/outbound/storage"
)

// ── ListFiles ────────────────────────────────────────────────────────────────

type ListFilesInput struct {
	Actor   mpdomain.Actor
	PlanID  int64
	TeamID  int64 // 0 = all
	Search  string
}

type ListFilesOutput struct {
	Files []mpdomain.PlanFileEntity
}

type listFilesRepo interface {
	ListFiles(ctx context.Context, q repository.PlanFileListQuery) ([]mpdomain.PlanFileEntity, error)
}

type ListFilesUseCase struct {
	repo listFilesRepo
}

func NewListFilesUseCase(repo listFilesRepo) *ListFilesUseCase {
	return &ListFilesUseCase{repo: repo}
}

func (uc *ListFilesUseCase) Execute(ctx context.Context, input ListFilesInput) (*ListFilesOutput, error) {
	// Non-admin: force filter to own team
	effectiveTeamID := input.TeamID
	if !input.Actor.IsAdmin() {
		if input.Actor.TeamID != nil {
			effectiveTeamID = *input.Actor.TeamID
		} else {
			effectiveTeamID = -1 // no results
		}
	}

	files, err := uc.repo.ListFiles(ctx, repository.PlanFileListQuery{
		MonthlyPlanID: input.PlanID,
		TeamID:        effectiveTeamID,
		Search:        input.Search,
	})
	if err != nil {
		return nil, fmt.Errorf("list files failed: %w", err)
	}
	return &ListFilesOutput{Files: files}, nil
}

// ── GetFile ──────────────────────────────────────────────────────────────────

type GetFileInput struct {
	Actor  mpdomain.Actor
	FileID int64
}

type GetFileOutput struct {
	File mpdomain.PlanFileEntity
}

type getFileRepo interface {
	GetFileByID(ctx context.Context, id int64) (*mpdomain.PlanFileEntity, error)
}

type GetFileUseCase struct {
	repo getFileRepo
}

func NewGetFileUseCase(repo getFileRepo) *GetFileUseCase {
	return &GetFileUseCase{repo: repo}
}

func (uc *GetFileUseCase) Execute(ctx context.Context, input GetFileInput) (*GetFileOutput, error) {
	if input.FileID <= 0 {
		return nil, mpdomain.ErrInvalidFileID
	}

	file, err := uc.repo.GetFileByID(ctx, input.FileID)
	if err != nil {
		return nil, mpdomain.ErrFileNotFound
	}
	if file == nil {
		return nil, mpdomain.ErrFileNotFound
	}

	// Non-admin can only access own team's files or master plans
	if !file.IsMasterPlan && !input.Actor.CanAccessFile(file.TeamID) {
		return nil, mpdomain.ErrForbiddenAction
	}

	return &GetFileOutput{File: *file}, nil
}

// ── SoftDeleteFile ───────────────────────────────────────────────────────────

type SoftDeleteFileInput struct {
	Actor  mpdomain.Actor
	FileID int64
}

type SoftDeleteFileOutput struct{}

type softDeleteRepo interface {
	GetFileByID(ctx context.Context, id int64) (*mpdomain.PlanFileEntity, error)
	SoftDeleteFile(ctx context.Context, input repository.PlanFileSoftDeleteInput) error
}

type SoftDeleteFileUseCase struct {
	repo softDeleteRepo
}

func NewSoftDeleteFileUseCase(repo softDeleteRepo) *SoftDeleteFileUseCase {
	return &SoftDeleteFileUseCase{repo: repo}
}

func (uc *SoftDeleteFileUseCase) Execute(ctx context.Context, input SoftDeleteFileInput) (*SoftDeleteFileOutput, error) {
	if input.FileID <= 0 {
		return nil, mpdomain.ErrInvalidFileID
	}

	file, err := uc.repo.GetFileByID(ctx, input.FileID)
	if err != nil || file == nil {
		return nil, mpdomain.ErrFileNotFound
	}
	if file.IsDeleted {
		return nil, mpdomain.ErrFileNotDeleted // already deleted
	}
	if !file.IsMasterPlan && !input.Actor.CanAccessFile(file.TeamID) {
		return nil, mpdomain.ErrForbiddenAction
	}

	if err := uc.repo.SoftDeleteFile(ctx, repository.PlanFileSoftDeleteInput{ID: input.FileID}); err != nil {
		return nil, fmt.Errorf("soft delete failed: %w", err)
	}
	return &SoftDeleteFileOutput{}, nil
}

// ── RestoreFile ──────────────────────────────────────────────────────────────

type RestoreFileInput struct {
	Actor  mpdomain.Actor
	FileID int64
}

type RestoreFileOutput struct {
	File mpdomain.PlanFileEntity
}

type restoreRepo interface {
	GetFileByID(ctx context.Context, id int64) (*mpdomain.PlanFileEntity, error)
	RestoreFile(ctx context.Context, input repository.PlanFileRestoreInput) (*mpdomain.PlanFileEntity, error)
}

type RestoreFileUseCase struct {
	repo restoreRepo
}

func NewRestoreFileUseCase(repo restoreRepo) *RestoreFileUseCase {
	return &RestoreFileUseCase{repo: repo}
}

func (uc *RestoreFileUseCase) Execute(ctx context.Context, input RestoreFileInput) (*RestoreFileOutput, error) {
	if input.FileID <= 0 {
		return nil, mpdomain.ErrInvalidFileID
	}

	file, err := uc.repo.GetFileByID(ctx, input.FileID)
	if err != nil || file == nil {
		return nil, mpdomain.ErrFileNotFound
	}
	if !file.IsDeleted {
		return nil, mpdomain.ErrFileNotDeleted // not deleted, can't restore
	}
	if !file.IsMasterPlan && !input.Actor.CanAccessFile(file.TeamID) {
		return nil, mpdomain.ErrForbiddenAction
	}

	restored, err := uc.repo.RestoreFile(ctx, repository.PlanFileRestoreInput{ID: input.FileID})
	if err != nil {
		return nil, fmt.Errorf("restore failed: %w", err)
	}
	return &RestoreFileOutput{File: *restored}, nil
}

// ── HardDeleteFile ───────────────────────────────────────────────────────────

type HardDeleteFileInput struct {
	Actor  mpdomain.Actor
	FileID int64
}

type HardDeleteFileOutput struct{}

type hardDeleteRepo interface {
	GetFileByID(ctx context.Context, id int64) (*mpdomain.PlanFileEntity, error)
	HardDeleteFile(ctx context.Context, input repository.PlanFileHardDeleteInput) error
}

type HardDeleteFileUseCase struct {
	repo   hardDeleteRepo
	storage storage.ObjectStorage
}

func NewHardDeleteFileUseCase(repo hardDeleteRepo, s storage.ObjectStorage) *HardDeleteFileUseCase {
	return &HardDeleteFileUseCase{repo: repo, storage: s}
}

func (uc *HardDeleteFileUseCase) Execute(ctx context.Context, input HardDeleteFileInput) (*HardDeleteFileOutput, error) {
	if !input.Actor.IsAdmin() {
		return nil, mpdomain.ErrForbiddenAction
	}
	if input.FileID <= 0 {
		return nil, mpdomain.ErrInvalidFileID
	}

	file, err := uc.repo.GetFileByID(ctx, input.FileID)
	if err != nil || file == nil {
		return nil, mpdomain.ErrFileNotFound
	}

	// Delete from storage (best-effort)
	_ = uc.storage.DeleteObject(ctx, file.FileKey)

	// Hard delete from DB
	if err := uc.repo.HardDeleteFile(ctx, repository.PlanFileHardDeleteInput{ID: input.FileID}); err != nil {
		return nil, fmt.Errorf("hard delete failed: %w", err)
	}
	return &HardDeleteFileOutput{}, nil
}

// ── GetSubmissionStatus ──────────────────────────────────────────────────────

type GetSubmissionStatusInput struct {
	Actor  mpdomain.Actor
	PlanID int64
}

type GetSubmissionStatusOutput struct {
	Deadline string
	Teams    []mpdomain.SubmissionStatusEntry
}

type submissionRepo interface {
	GetOrCreateSettings(ctx context.Context) (*mpdomain.SettingsEntity, error)
	FindOrCreatePeriod(ctx context.Context, year, month int) (*mpdomain.Entity, error)
	CountFilesByTeam(ctx context.Context, q repository.SubmissionQuery) ([]repository.TeamCountRow, error)
	ListTeams(ctx context.Context) ([]repository.TeamInfo, error)
}

type GetSubmissionStatusUseCase struct {
	repo submissionRepo
}

func NewGetSubmissionStatusUseCase(repo submissionRepo) *GetSubmissionStatusUseCase {
	return &GetSubmissionStatusUseCase{repo: repo}
}

func (uc *GetSubmissionStatusUseCase) Execute(ctx context.Context, input GetSubmissionStatusInput) (*GetSubmissionStatusOutput, error) {
	if !input.Actor.IsAdmin() {
		return nil, mpdomain.ErrForbiddenAction
	}

	_, err := uc.repo.GetOrCreateSettings(ctx)
	if err != nil {
		return nil, mpdomain.ErrSettingsLoadFail
	}

	teams, err := uc.repo.ListTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("list teams failed: %w", err)
	}

	counts, err := uc.repo.CountFilesByTeam(ctx, repository.SubmissionQuery{PlanID: input.PlanID})
	if err != nil {
		return nil, fmt.Errorf("count files failed: %w", err)
	}

	// Build count map
	countMap := make(map[int64]int, len(counts))
	for _, c := range counts {
		countMap[c.TeamID] = c.Count
	}

	// Determine status for each team
	deadline := "" // TODO: compute from settings.LockDay
	entries := make([]mpdomain.SubmissionStatusEntry, 0, len(teams))
	for _, team := range teams {
		count := countMap[team.ID]
		status := "pending"
		if count > 0 {
			status = "submitted"
		}
		entries = append(entries, mpdomain.SubmissionStatusEntry{
			TeamID:    team.ID,
			TeamName:  team.Name,
			Status:    status,
			FileCount: count,
		})
	}

	return &GetSubmissionStatusOutput{
		Deadline: deadline,
		Teams:    entries,
	}, nil
}

// ── Ensure compile-time interface satisfaction ───────────────────────────────
var _ listFilesRepo = (repository.MonthlyPlanRepository)(nil)
var _ getFileRepo = (repository.MonthlyPlanRepository)(nil)
var _ softDeleteRepo = (repository.MonthlyPlanRepository)(nil)
var _ restoreRepo = (repository.MonthlyPlanRepository)(nil)
var _ hardDeleteRepo = (repository.MonthlyPlanRepository)(nil)
var _ submissionRepo = (repository.MonthlyPlanRepository)(nil)
