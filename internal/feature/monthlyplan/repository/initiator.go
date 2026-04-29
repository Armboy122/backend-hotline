package repository

import (
	"context"
	"time"

	"backend-hotlines3/internal/feature/monthlyplan/entity"
	"backend-hotlines3/pkg/s3"

	"gorm.io/gorm"
)

type PeriodFindQuery struct {
	Year  int
	Month int
}

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
	TeamID        int64
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

type FileSizeLogCreateInput struct {
	PlanFileID    int64
	UploadedByID  int64
	FileSizeBytes int64
}

type TeamCountRow struct {
	TeamID int64 `gorm:"column:team_id"`
	Count  int   `gorm:"column:count"`
}

type SubmissionQuery struct {
	PlanID int64
}

type TeamInfo struct {
	ID   int64
	Name string
}

type Repository interface {
	FindPeriod(ctx context.Context, q PeriodFindQuery) (*entity.Entity, error)
	FindOrCreatePeriod(ctx context.Context, year, month int) (*entity.Entity, error)
	GetOrCreateSettings(ctx context.Context) (*entity.SettingsEntity, error)
	UpdateSettings(ctx context.Context, s *entity.SettingsEntity) error
	ListFiles(ctx context.Context, q PlanFileListQuery) ([]entity.PlanFileEntity, error)
	GetFileByID(ctx context.Context, id int64) (*entity.PlanFileEntity, error)
	CreateFile(ctx context.Context, input PlanFileCreateInput) (*entity.PlanFileEntity, error)
	SoftDeleteFile(ctx context.Context, input PlanFileSoftDeleteInput) error
	RestoreFile(ctx context.Context, input PlanFileRestoreInput) (*entity.PlanFileEntity, error)
	HardDeleteFile(ctx context.Context, input PlanFileHardDeleteInput) error
	CreateFileSizeLog(ctx context.Context, input FileSizeLogCreateInput) error
	CountFilesByTeam(ctx context.Context, q SubmissionQuery) ([]TeamCountRow, error)
	ListTeams(ctx context.Context) ([]TeamInfo, error)
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

type PresignResult struct {
	UploadURL string
	FileURL   string
	FileKey   string
}

type StoragePort interface {
	PresignUpload(ctx context.Context, fileKey, contentType string, expiration time.Duration) (*PresignResult, error)
	PresignDownload(ctx context.Context, fileKey string, expiration time.Duration) (string, error)
	DeleteObject(ctx context.Context, fileKey string) error
	PublicURL(fileKey string) string
}

func NewR2Storage(client *s3.R2Client) StoragePort {
	return &r2Storage{client: client}
}
