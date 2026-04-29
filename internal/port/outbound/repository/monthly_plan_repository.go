package repository

import (
	"context"
	"time"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
)

// ── Monthly Plan Period ──────────────────────────────────────────────────────

type PeriodFindQuery struct {
	Year  int
	Month int
}

type PeriodCreateInput struct {
	Year  int
	Month int
}

// ── Settings ─────────────────────────────────────────────────────────────────

type SettingsUpdateInput struct {
	LockDay                 *int
	AutoCreateDay           *int
	AutoCreateTarget        *string
	AllowedFileTypes        []string
	MaxFileSizeMB           *int
	ReminderStartDay        *int
	AdminCanUploadAfterLock *bool
}

// ── Plan File ────────────────────────────────────────────────────────────────

type PlanFileCreateInput struct {
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

type PlanFileListQuery struct {
	MonthlyPlanID int64
	TeamID        int64 // 0 = all teams
	Search        string
}

type PlanFileSoftDeleteInput struct {
	ID int64
}

type PlanFileRestoreInput struct {
	ID int64
}

type PlanFileHardDeleteInput struct {
	ID int64
}

// ── File Size Log ────────────────────────────────────────────────────────────

type FileSizeLogCreateInput struct {
	PlanFileID    int64
	UploadedByID  int64
	FileSizeBytes int64
}

// ── Submission Status ────────────────────────────────────────────────────────

type TeamCountRow struct {
	TeamID int64
	Count  int
}

type SubmissionQuery struct {
	PlanID int64
}

// ── MonthlyPlanRepository — all monthly plan persistence operations ──────────

type MonthlyPlanRepository interface {
	// Period
	FindPeriod(ctx context.Context, q PeriodFindQuery) (*mpdomain.Entity, error)
	FindOrCreatePeriod(ctx context.Context, year, month int) (*mpdomain.Entity, error)

	// Settings
	GetOrCreateSettings(ctx context.Context) (*mpdomain.SettingsEntity, error)
	UpdateSettings(ctx context.Context, s *mpdomain.SettingsEntity) error

	// Plan Files
	ListFiles(ctx context.Context, q PlanFileListQuery) ([]mpdomain.PlanFileEntity, error)
	GetFileByID(ctx context.Context, id int64) (*mpdomain.PlanFileEntity, error)
	CreateFile(ctx context.Context, input PlanFileCreateInput) (*mpdomain.PlanFileEntity, error)
	SoftDeleteFile(ctx context.Context, input PlanFileSoftDeleteInput) error
	RestoreFile(ctx context.Context, input PlanFileRestoreInput) (*mpdomain.PlanFileEntity, error)
	HardDeleteFile(ctx context.Context, input PlanFileHardDeleteInput) error

	// File Size Log
	CreateFileSizeLog(ctx context.Context, input FileSizeLogCreateInput) error

	// Submission Status
	CountFilesByTeam(ctx context.Context, q SubmissionQuery) ([]TeamCountRow, error)
	ListTeams(ctx context.Context) ([]TeamInfo, error)
}

// TeamInfo is a minimal team record for aggregation queries.
type TeamInfo struct {
	ID   int64
	Name string
}

// ── User lookup (reuse from other modules or keep separate) ──────────────────

type UserRepository interface {
	GetTeamIDByUserID(ctx context.Context, userID int64) (*int64, error)
}

// ── Clock dependency for testability ─────────────────────────────────────────

type Clock interface {
	Now() time.Time
}
