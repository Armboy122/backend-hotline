package monthlyplan

import (
	"context"
	"time"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
	"backend-hotlines3/internal/port/outbound/storage"
)

// ── Repository types (re-exported from legacy port for transitional compat) ──

// Repository is the module-local persistence interface alias.
// During the migration it mirrors repository.MonthlyPlanRepository
// so that the module is self-contained.
type Repository interface {
	// Period
	FindPeriod(ctx context.Context, year, month int) (*mpdomain.Entity, error)
	FindOrCreatePeriod(ctx context.Context, year, month int) (*mpdomain.Entity, error)

	// Settings
	GetOrCreateSettings(ctx context.Context) (*mpdomain.SettingsEntity, error)
	UpdateSettings(ctx context.Context, s *mpdomain.SettingsEntity) error

	// Plan Files
	ListFiles(ctx context.Context, planID int64, teamID int64, search string) ([]mpdomain.PlanFileEntity, error)
	GetFileByID(ctx context.Context, id int64) (*mpdomain.PlanFileEntity, error)
	CreateFile(ctx context.Context, input CreateFileInput) (*mpdomain.PlanFileEntity, error)
	SoftDeleteFile(ctx context.Context, id int64) error
	RestoreFile(ctx context.Context, id int64) (*mpdomain.PlanFileEntity, error)
	HardDeleteFile(ctx context.Context, id int64) error

	// File Size Log
	CreateFileSizeLog(ctx context.Context, planFileID, uploadedByID, fileSizeBytes int64) error

	// Submission Status
	CountFilesByTeam(ctx context.Context, planID int64) ([]TeamCountRow, error)
	ListTeams(ctx context.Context) ([]TeamInfo, error)
}

// CreateFileInput holds the data needed to create a plan file record.
type CreateFileInput struct {
	MonthlyPlanID int64
	TeamID        *int64
	UploadedByID  int64
	FileKey       string
	FileURL       string
	FileName      string
	FileSizeBytes int64
	Description   *string
	IsMasterPlan  bool
}

// TeamCountRow is one row from the file-count-by-team aggregation.
type TeamCountRow struct {
	TeamID int64
	Count  int
}

// TeamInfo is a minimal team record for aggregation queries.
type TeamInfo struct {
	ID   int64
	Name string
}

// ── StoragePort abstracts R2/S3 operations ────────────────────────────────────

type StoragePort = storage.ObjectStorage

// Clock dependency for testability.
type Clock interface {
	Now() time.Time
}
