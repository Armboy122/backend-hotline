package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
	"backend-hotlines3/internal/port/outbound/repository"
	"backend-hotlines3/internal/port/outbound/storage"
)

// ── PresignUpload ────────────────────────────────────────────────────────────

type PresignUploadInput struct {
	Actor      mpdomain.Actor
	PlanID     int64
	FileName   string
	FileType   string
	IsMaster   bool
	TeamID     *int64 // override team (admin only)
}

type PresignUploadOutput struct {
	UploadURL string
	FileURL   string
	FileKey   string
}

type presignUploadDeps interface {
	GetOrCreateSettings(ctx context.Context) (*mpdomain.SettingsEntity, error)
}

type PresignUploadUseCase struct {
	repo    presignUploadDeps
	storage storage.ObjectStorage
	clock   func() time.Time
}

func NewPresignUploadUseCase(repo presignUploadDeps, s storage.ObjectStorage) *PresignUploadUseCase {
	return &PresignUploadUseCase{repo: repo, storage: s, clock: time.Now}
}

func (uc *PresignUploadUseCase) Execute(ctx context.Context, input PresignUploadInput) (*PresignUploadOutput, error) {
	// Load settings for allowed file types
	settings, err := uc.repo.GetOrCreateSettings(ctx)
	if err != nil {
		return nil, mpdomain.ErrSettingsLoadFail
	}

	// Validate file type
	if !isAllowedType(input.FileType, settings.AllowedFileTypes) {
		return nil, mpdomain.ErrInvalidFileType
	}

	// Master plan: only admin
	if input.IsMaster && !input.Actor.CanUploadMasterPlan() {
		return nil, mpdomain.ErrForbiddenAction
	}

	// Team upload: check permissions
	if !input.IsMaster {
		effectiveTeamID := input.TeamID
		if effectiveTeamID == nil {
			effectiveTeamID = input.Actor.TeamID
		}
		if !input.Actor.CanUploadForTeam(effectiveTeamID) {
			return nil, mpdomain.ErrForbiddenAction
		}
	}

	// Generate file key
	ext := filepath.Ext(input.FileName)
	now := uc.clock()
	fileKey := fmt.Sprintf("monthly-plans/%d/%02d/%d-%s%s",
		now.Year(), now.Month(), input.PlanID,
		strings.ReplaceAll(strings.ReplaceAll(input.FileName, "/", "_"), "\\", "_"),
		ext,
	)

	result, err := uc.storage.PresignUpload(ctx, fileKey, input.FileType, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("presign failed: %w", err)
	}

	return &PresignUploadOutput{
		UploadURL: result.UploadURL,
		FileURL:   result.FileURL,
		FileKey:   result.FileKey,
	}, nil
}

// isAllowedType checks if the given MIME type is in the allowed list.
func isAllowedType(fileType string, allowed []string) bool {
	for _, a := range allowed {
		if a == fileType {
			return true
		}
	}
	return false
}

// ── ConfirmUpload ────────────────────────────────────────────────────────────

type ConfirmUploadInput struct {
	Actor        mpdomain.Actor
	PlanID       int64
	FileKey      string
	FileURL      string
	FileName     string
	FileSizeBytes int64
	Description  *string
	IsMasterPlan bool
	TeamID       *int64 // override team (admin only)
}

type ConfirmUploadOutput struct {
	File mpdomain.PlanFileEntity
}

type confirmUploadRepo interface {
	GetOrCreateSettings(ctx context.Context) (*mpdomain.SettingsEntity, error)
	CreateFile(ctx context.Context, input repository.PlanFileCreateInput) (*mpdomain.PlanFileEntity, error)
	CreateFileSizeLog(ctx context.Context, input repository.FileSizeLogCreateInput) error
}

type ConfirmUploadUseCase struct {
	repo confirmUploadRepo
}

func NewConfirmUploadUseCase(repo confirmUploadRepo) *ConfirmUploadUseCase {
	return &ConfirmUploadUseCase{repo: repo}
}

func (uc *ConfirmUploadUseCase) Execute(ctx context.Context, input ConfirmUploadInput) (*ConfirmUploadOutput, error) {
	// Master plan: only admin
	if input.IsMasterPlan && !input.Actor.CanUploadMasterPlan() {
		return nil, mpdomain.ErrForbiddenAction
	}

	// Determine effective team
	effectiveTeamID := input.TeamID
	if !input.IsMasterPlan {
		if effectiveTeamID == nil {
			effectiveTeamID = input.Actor.TeamID
		}
		if !input.Actor.CanUploadForTeam(effectiveTeamID) {
			return nil, mpdomain.ErrForbiddenAction
		}
	} else {
		// Master plans have no team
		effectiveTeamID = nil
	}

	// Create file record
	file, err := uc.repo.CreateFile(ctx, repository.PlanFileCreateInput{
		MonthlyPlanID: input.PlanID,
		TeamID:        effectiveTeamID,
		UploadedByID:  input.Actor.UserID,
		FileKey:       input.FileKey,
		FileURL:       input.FileURL,
		FileName:      input.FileName,
		FileSizeBytes: input.FileSizeBytes,
		Description:   input.Description,
		IsMasterPlan:  input.IsMasterPlan,
	})
	if err != nil {
		return nil, fmt.Errorf("create file failed: %w", err)
	}

	// Log file size
	_ = uc.repo.CreateFileSizeLog(ctx, repository.FileSizeLogCreateInput{
		PlanFileID:    file.ID,
		UploadedByID:  input.Actor.UserID,
		FileSizeBytes: input.FileSizeBytes,
	})

	return &ConfirmUploadOutput{File: *file}, nil
}

// ── Ensure compile-time interface satisfaction ───────────────────────────────
var _ presignUploadDeps = (repository.MonthlyPlanRepository)(nil)
var _ confirmUploadRepo = (repository.MonthlyPlanRepository)(nil)
