package usecase

import (
	"context"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
	"backend-hotlines3/internal/port/outbound/repository"
)

// ── GetSettings ──────────────────────────────────────────────────────────────

type GetSettingsInput struct{}

type GetSettingsOutput struct {
	Settings mpdomain.SettingsEntity
}

// single-method port
type settingsGetter interface {
	GetOrCreateSettings(ctx context.Context) (*mpdomain.SettingsEntity, error)
}

type GetSettingsUseCase struct {
	repo settingsGetter
}

func NewGetSettingsUseCase(repo settingsGetter) *GetSettingsUseCase {
	return &GetSettingsUseCase{repo: repo}
}

func (uc *GetSettingsUseCase) Execute(ctx context.Context, _ GetSettingsInput) (*GetSettingsOutput, error) {
	s, err := uc.repo.GetOrCreateSettings(ctx)
	if err != nil {
		return nil, mpdomain.ErrSettingsLoadFail
	}
	return &GetSettingsOutput{Settings: *s}, nil
}

// ── UpdateSettings ───────────────────────────────────────────────────────────

type UpdateSettingsInput struct {
	Actor                   mpdomain.Actor
	LockDay                 *int
	AutoCreateDay           *int
	AutoCreateTarget        *string
	AllowedFileTypes        []string
	MaxFileSizeMB           *int
	ReminderStartDay        *int
	AdminCanUploadAfterLock *bool
}

type UpdateSettingsOutput struct {
	Settings mpdomain.SettingsEntity
}

// composite port for update (read + write)
type settingsReadWriter interface {
	settingsGetter
	UpdateSettings(ctx context.Context, s *mpdomain.SettingsEntity) error
}

type UpdateSettingsUseCase struct {
	repo settingsReadWriter
}

func NewUpdateSettingsUseCase(repo settingsReadWriter) *UpdateSettingsUseCase {
	return &UpdateSettingsUseCase{repo: repo}
}

func (uc *UpdateSettingsUseCase) Execute(ctx context.Context, input UpdateSettingsInput) (*UpdateSettingsOutput, error) {
	if !input.Actor.IsAdmin() {
		return nil, mpdomain.ErrForbiddenAction
	}

	s, err := uc.repo.GetOrCreateSettings(ctx)
	if err != nil {
		return nil, mpdomain.ErrSettingsLoadFail
	}

	// Apply patches
	if input.LockDay != nil {
		s.LockDay = *input.LockDay
	}
	if input.AutoCreateDay != nil {
		s.AutoCreateDay = *input.AutoCreateDay
	}
	if input.AutoCreateTarget != nil {
		s.AutoCreateTarget = *input.AutoCreateTarget
	}
	if input.AllowedFileTypes != nil {
		s.AllowedFileTypes = input.AllowedFileTypes
	}
	if input.MaxFileSizeMB != nil {
		s.MaxFileSizeMB = input.MaxFileSizeMB
	}
	if input.ReminderStartDay != nil {
		s.ReminderStartDay = *input.ReminderStartDay
	}
	if input.AdminCanUploadAfterLock != nil {
		s.AdminCanUploadAfterLock = *input.AdminCanUploadAfterLock
	}

	if err := uc.repo.UpdateSettings(ctx, s); err != nil {
		return nil, mpdomain.ErrSettingsSaveFail
	}
	return &UpdateSettingsOutput{Settings: *s}, nil
}

// ── EnsurePeriod ─────────────────────────────────────────────────────────────

type EnsurePeriodInput struct {
	Year  int
	Month int
}

type EnsurePeriodOutput struct {
	Period  mpdomain.Entity
	Created bool
}

type periodFindCreator interface {
	FindOrCreatePeriod(ctx context.Context, year, month int) (*mpdomain.Entity, error)
}

type EnsurePeriodUseCase struct {
	repo periodFindCreator
}

func NewEnsurePeriodUseCase(repo periodFindCreator) *EnsurePeriodUseCase {
	return &EnsurePeriodUseCase{repo: repo}
}

func (uc *EnsurePeriodUseCase) Execute(ctx context.Context, input EnsurePeriodInput) (*EnsurePeriodOutput, error) {
	if input.Year < 2000 || input.Year > 2100 || input.Month < 1 || input.Month > 12 {
		return nil, mpdomain.ErrInvalidPeriod
	}
	p, err := uc.repo.FindOrCreatePeriod(ctx, input.Year, input.Month)
	if err != nil {
		return nil, err
	}
	return &EnsurePeriodOutput{Period: *p}, nil
}

// ── Ensure compile-time interface satisfaction ───────────────────────────────
var _ settingsGetter = (repository.MonthlyPlanRepository)(nil)
var _ settingsReadWriter = (repository.MonthlyPlanRepository)(nil)
var _ periodFindCreator = (repository.MonthlyPlanRepository)(nil)
