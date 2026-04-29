package usecase

import (
	"context"
	"testing"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
)

// ── fake settings repo ───────────────────────────────────────────────────────

type fakeSettingsRepo struct {
	settings *mpdomain.SettingsEntity
	updateFn func(s *mpdomain.SettingsEntity) error
}

func (f *fakeSettingsRepo) GetOrCreateSettings(_ context.Context) (*mpdomain.SettingsEntity, error) {
	if f.settings == nil {
		return &mpdomain.SettingsEntity{
			ID:                      1,
			LockDay:                 23,
			AutoCreateDay:           1,
			AutoCreateTarget:        "next_month",
			AllowedFileTypes:        []string{"application/pdf"},
			ReminderStartDay:        20,
			AdminCanUploadAfterLock: true,
		}, nil
	}
	return f.settings, nil
}

func (f *fakeSettingsRepo) UpdateSettings(_ context.Context, s *mpdomain.SettingsEntity) error {
	if f.updateFn != nil {
		return f.updateFn(s)
	}
	f.settings = s
	return nil
}

// ── GetSettings tests ────────────────────────────────────────────────────────

func TestGetSettings_Success(t *testing.T) {
	uc := NewGetSettingsUseCase(&fakeSettingsRepo{})
	out, err := uc.Execute(context.Background(), GetSettingsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Settings.LockDay != 23 {
		t.Errorf("LockDay = %d, want 23", out.Settings.LockDay)
	}
	if out.Settings.AutoCreateTarget != "next_month" {
		t.Errorf("AutoCreateTarget = %q, want %q", out.Settings.AutoCreateTarget, "next_month")
	}
}

func TestGetSettings_RepoError(t *testing.T) {
	uc := NewGetSettingsUseCase(&fakeSettingsRepo{
		updateFn: func(s *mpdomain.SettingsEntity) error { return nil },
	})
	// Override GetOrCreateSettings to fail
	repo := &faultySettingsRepo{}
	uc = NewGetSettingsUseCase(repo)
	_, err := uc.Execute(context.Background(), GetSettingsInput{})
	if err != mpdomain.ErrSettingsLoadFail {
		t.Fatalf("expected ErrSettingsLoadFail, got: %v", err)
	}
}

type faultySettingsRepo struct{}

func (f *faultySettingsRepo) GetOrCreateSettings(_ context.Context) (*mpdomain.SettingsEntity, error) {
	return nil, mpdomain.ErrSettingsLoadFail
}
func (f *faultySettingsRepo) UpdateSettings(_ context.Context, _ *mpdomain.SettingsEntity) error {
	return nil
}

// ── UpdateSettings tests ─────────────────────────────────────────────────────

func TestUpdateSettings_ForbiddenNonAdmin(t *testing.T) {
	uc := NewUpdateSettingsUseCase(&fakeSettingsRepo{})
	_, err := uc.Execute(context.Background(), UpdateSettingsInput{
		Actor: mpdomain.Actor{UserID: 1, Role: "staff"},
	})
	if err != mpdomain.ErrForbiddenAction {
		t.Fatalf("expected ErrForbiddenAction, got: %v", err)
	}
}

func TestUpdateSettings_Success(t *testing.T) {
	repo := &fakeSettingsRepo{}
	uc := NewUpdateSettingsUseCase(repo)

	newLockDay := 25
	out, err := uc.Execute(context.Background(), UpdateSettingsInput{
		Actor:   mpdomain.Actor{UserID: 1, Role: "admin"},
		LockDay: &newLockDay,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Settings.LockDay != 25 {
		t.Errorf("LockDay = %d, want 25", out.Settings.LockDay)
	}
}

func TestUpdateSettings_AllFields(t *testing.T) {
	repo := &fakeSettingsRepo{}
	uc := NewUpdateSettingsUseCase(repo)

	lockDay := 28
	autoDay := 2
	target := "current_month"
	types := []string{"application/pdf", "image/png"}
	maxMB := 50
	reminderDay := 18
	adminUpload := false

	out, err := uc.Execute(context.Background(), UpdateSettingsInput{
		Actor:                   mpdomain.Actor{UserID: 1, Role: "admin"},
		LockDay:                 &lockDay,
		AutoCreateDay:           &autoDay,
		AutoCreateTarget:        &target,
		AllowedFileTypes:        types,
		MaxFileSizeMB:           &maxMB,
		ReminderStartDay:        &reminderDay,
		AdminCanUploadAfterLock: &adminUpload,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Settings.LockDay != 28 {
		t.Errorf("LockDay = %d, want 28", out.Settings.LockDay)
	}
	if out.Settings.AutoCreateDay != 2 {
		t.Errorf("AutoCreateDay = %d, want 2", out.Settings.AutoCreateDay)
	}
	if out.Settings.AutoCreateTarget != "current_month" {
		t.Errorf("AutoCreateTarget = %q", out.Settings.AutoCreateTarget)
	}
	if len(out.Settings.AllowedFileTypes) != 2 {
		t.Errorf("AllowedFileTypes len = %d, want 2", len(out.Settings.AllowedFileTypes))
	}
	if *out.Settings.MaxFileSizeMB != 50 {
		t.Errorf("MaxFileSizeMB = %d, want 50", *out.Settings.MaxFileSizeMB)
	}
	if out.Settings.ReminderStartDay != 18 {
		t.Errorf("ReminderStartDay = %d, want 18", out.Settings.ReminderStartDay)
	}
	if out.Settings.AdminCanUploadAfterLock != false {
		t.Error("AdminCanUploadAfterLock should be false")
	}
}

func TestUpdateSettings_SaveFail(t *testing.T) {
	uc := NewUpdateSettingsUseCase(&fakeSettingsRepo{
		updateFn: func(s *mpdomain.SettingsEntity) error {
			return mpdomain.ErrSettingsSaveFail
		},
	})
	lockDay := 25
	_, err := uc.Execute(context.Background(), UpdateSettingsInput{
		Actor:   mpdomain.Actor{UserID: 1, Role: "admin"},
		LockDay: &lockDay,
	})
	if err != mpdomain.ErrSettingsSaveFail {
		t.Fatalf("expected ErrSettingsSaveFail, got: %v", err)
	}
}

// ── EnsurePeriod tests ───────────────────────────────────────────────────────

type fakePeriodRepo struct {
	period *mpdomain.Entity
	err    error
}

func (f *fakePeriodRepo) FindOrCreatePeriod(_ context.Context, _, _ int) (*mpdomain.Entity, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.period != nil {
		return f.period, nil
	}
	return &mpdomain.Entity{ID: 1, Year: 2026, Month: 5}, nil
}

func TestEnsurePeriod_InvalidYear(t *testing.T) {
	uc := NewEnsurePeriodUseCase(&fakePeriodRepo{})
	_, err := uc.Execute(context.Background(), EnsurePeriodInput{Year: 1999, Month: 5})
	if err != mpdomain.ErrInvalidPeriod {
		t.Fatalf("expected ErrInvalidPeriod, got: %v", err)
	}
}

func TestEnsurePeriod_InvalidMonth(t *testing.T) {
	uc := NewEnsurePeriodUseCase(&fakePeriodRepo{})
	_, err := uc.Execute(context.Background(), EnsurePeriodInput{Year: 2026, Month: 0})
	if err != mpdomain.ErrInvalidPeriod {
		t.Fatalf("expected ErrInvalidPeriod, got: %v", err)
	}
}

func TestEnsurePeriod_Success(t *testing.T) {
	uc := NewEnsurePeriodUseCase(&fakePeriodRepo{})
	out, err := uc.Execute(context.Background(), EnsurePeriodInput{Year: 2026, Month: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Period.Year != 2026 || out.Period.Month != 5 {
		t.Errorf("Period = %d/%d, want 2026/5", out.Period.Year, out.Period.Month)
	}
}
