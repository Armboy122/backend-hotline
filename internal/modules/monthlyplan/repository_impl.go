package monthlyplan

import (
	"context"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
	gormadapter "backend-hotlines3/internal/adapter/out/persistence/gorm"
	"backend-hotlines3/internal/port/outbound/repository"

	"gorm.io/gorm"
)

// repoBridge wraps the existing gorm adapter so it satisfies the
// module-local Repository interface.  This is the same bridge pattern
// used by the task module.
type repoBridge struct {
	inner *gormadapter.MonthlyPlanRepository
}

// NewRepositoryImpl creates a module-local Repository backed by the
// existing gorm adapter.
func NewRepositoryImpl(db *gorm.DB) Repository {
	return &repoBridge{inner: gormadapter.NewMonthlyPlanRepository(db)}
}

func (b *repoBridge) FindPeriod(ctx context.Context, year, month int) (*mpdomain.Entity, error) {
	return b.inner.FindPeriod(ctx, repository.PeriodFindQuery{Year: year, Month: month})
}

func (b *repoBridge) FindOrCreatePeriod(ctx context.Context, year, month int) (*mpdomain.Entity, error) {
	return b.inner.FindOrCreatePeriod(ctx, year, month)
}

func (b *repoBridge) GetOrCreateSettings(ctx context.Context) (*mpdomain.SettingsEntity, error) {
	return b.inner.GetOrCreateSettings(ctx)
}

func (b *repoBridge) UpdateSettings(ctx context.Context, s *mpdomain.SettingsEntity) error {
	return b.inner.UpdateSettings(ctx, s)
}

func (b *repoBridge) ListFiles(ctx context.Context, planID int64, teamID int64, search string) ([]mpdomain.PlanFileEntity, error) {
	return b.inner.ListFiles(ctx, repository.PlanFileListQuery{
		MonthlyPlanID: planID,
		TeamID:        teamID,
		Search:        search,
	})
}

func (b *repoBridge) GetFileByID(ctx context.Context, id int64) (*mpdomain.PlanFileEntity, error) {
	return b.inner.GetFileByID(ctx, id)
}

func (b *repoBridge) CreateFile(ctx context.Context, input CreateFileInput) (*mpdomain.PlanFileEntity, error) {
	return b.inner.CreateFile(ctx, repository.PlanFileCreateInput{
		MonthlyPlanID: input.MonthlyPlanID,
		TeamID:        input.TeamID,
		UploadedByID:  input.UploadedByID,
		FileKey:       input.FileKey,
		FileURL:       input.FileURL,
		FileName:      input.FileName,
		FileSizeBytes: input.FileSizeBytes,
		Description:   input.Description,
		IsMasterPlan:  input.IsMasterPlan,
	})
}

func (b *repoBridge) SoftDeleteFile(ctx context.Context, id int64) error {
	return b.inner.SoftDeleteFile(ctx, repository.PlanFileSoftDeleteInput{ID: id})
}

func (b *repoBridge) RestoreFile(ctx context.Context, id int64) (*mpdomain.PlanFileEntity, error) {
	return b.inner.RestoreFile(ctx, repository.PlanFileRestoreInput{ID: id})
}

func (b *repoBridge) HardDeleteFile(ctx context.Context, id int64) error {
	return b.inner.HardDeleteFile(ctx, repository.PlanFileHardDeleteInput{ID: id})
}

func (b *repoBridge) CreateFileSizeLog(ctx context.Context, planFileID, uploadedByID, fileSizeBytes int64) error {
	return b.inner.CreateFileSizeLog(ctx, repository.FileSizeLogCreateInput{
		PlanFileID:    planFileID,
		UploadedByID:  uploadedByID,
		FileSizeBytes: fileSizeBytes,
	})
}

func (b *repoBridge) CountFilesByTeam(ctx context.Context, planID int64) ([]TeamCountRow, error) {
	rows, err := b.inner.CountFilesByTeam(ctx, repository.SubmissionQuery{PlanID: planID})
	if err != nil {
		return nil, err
	}
	out := make([]TeamCountRow, len(rows))
	for i, r := range rows {
		out[i] = TeamCountRow{TeamID: r.TeamID, Count: r.Count}
	}
	return out, nil
}

func (b *repoBridge) ListTeams(ctx context.Context) ([]TeamInfo, error) {
	teams, err := b.inner.ListTeams(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TeamInfo, len(teams))
	for i, t := range teams {
		out[i] = TeamInfo{ID: t.ID, Name: t.Name}
	}
	return out, nil
}
