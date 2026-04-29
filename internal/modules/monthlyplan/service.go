package monthlyplan

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
)

// Service is the module-local service boundary for MonthlyPlan.
// It is a Phase C placeholder that keeps module code self-contained
// while the legacy v1 handler still owns the HTTP behaviour.
type Service struct {
	repo    Repository
	storage StoragePort
	clock   func() time.Time
}

// NewService creates a new Service with the given repository and storage.
func NewService(repo Repository, storage StoragePort) *Service {
	return &Service{repo: repo, storage: storage, clock: time.Now}
}

// ── Settings ─────────────────────────────────────────────────────────────────

// GetSettings returns the current monthly plan settings.
func (s *Service) GetSettings(ctx context.Context) (*mpdomain.SettingsEntity, error) {
	settings, err := s.repo.GetOrCreateSettings(ctx)
	if err != nil {
		return nil, mpdomain.ErrSettingsLoadFail
	}
	return settings, nil
}

// UpdateSettings patches the monthly plan settings (admin only).
func (s *Service) UpdateSettings(ctx context.Context, actor mpdomain.Actor, patch SettingsPatch) (*mpdomain.SettingsEntity, error) {
	if !actor.IsAdmin() {
		return nil, mpdomain.ErrForbiddenAction
	}
	settings, err := s.repo.GetOrCreateSettings(ctx)
	if err != nil {
		return nil, mpdomain.ErrSettingsLoadFail
	}
	patch.apply(settings)
	if err := s.repo.UpdateSettings(ctx, settings); err != nil {
		return nil, mpdomain.ErrSettingsSaveFail
	}
	return settings, nil
}

// SettingsPatch holds optional fields for partial settings update.
type SettingsPatch struct {
	LockDay                 *int
	AutoCreateDay           *int
	AutoCreateTarget        *string
	AllowedFileTypes        []string
	MaxFileSizeMB           *int
	ReminderStartDay        *int
	AdminCanUploadAfterLock *bool
}

func (p SettingsPatch) apply(s *mpdomain.SettingsEntity) {
	if p.LockDay != nil {
		s.LockDay = *p.LockDay
	}
	if p.AutoCreateDay != nil {
		s.AutoCreateDay = *p.AutoCreateDay
	}
	if p.AutoCreateTarget != nil {
		s.AutoCreateTarget = *p.AutoCreateTarget
	}
	if p.AllowedFileTypes != nil {
		s.AllowedFileTypes = p.AllowedFileTypes
	}
	if p.MaxFileSizeMB != nil {
		s.MaxFileSizeMB = p.MaxFileSizeMB
	}
	if p.ReminderStartDay != nil {
		s.ReminderStartDay = *p.ReminderStartDay
	}
	if p.AdminCanUploadAfterLock != nil {
		s.AdminCanUploadAfterLock = *p.AdminCanUploadAfterLock
	}
}

// ── Period ───────────────────────────────────────────────────────────────────

// EnsurePeriod returns the period for the given year/month, creating it if needed.
func (s *Service) EnsurePeriod(ctx context.Context, year, month int) (*mpdomain.Entity, error) {
	if year < 2000 || year > 2100 || month < 1 || month > 12 {
		return nil, mpdomain.ErrInvalidPeriod
	}
	return s.repo.FindOrCreatePeriod(ctx, year, month)
}

// ── Files ────────────────────────────────────────────────────────────────────

// PresignUpload generates a presigned upload URL after validating permissions and file type.
func (s *Service) PresignUpload(ctx context.Context, actor mpdomain.Actor, planID int64, fileName, fileType string, isMaster bool, teamID *int64) (*PresignResult, error) {
	settings, err := s.repo.GetOrCreateSettings(ctx)
	if err != nil {
		return nil, mpdomain.ErrSettingsLoadFail
	}
	if !isAllowedType(fileType, settings.AllowedFileTypes) {
		return nil, mpdomain.ErrInvalidFileType
	}
	if isMaster && !actor.CanUploadMasterPlan() {
		return nil, mpdomain.ErrForbiddenAction
	}
	if !isMaster {
		effectiveTeamID := teamID
		if effectiveTeamID == nil {
			effectiveTeamID = actor.TeamID
		}
		if !actor.CanUploadForTeam(effectiveTeamID) {
			return nil, mpdomain.ErrForbiddenAction
		}
	}

	ext := filepath.Ext(fileName)
	now := s.clock()
	fileKey := fmt.Sprintf("monthly-plans/%d/%02d/%d-%s%s",
		now.Year(), now.Month(), planID,
		strings.ReplaceAll(strings.ReplaceAll(fileName, "/", "_"), "\\", "_"),
		ext,
	)

	result, err := s.storage.PresignUpload(ctx, fileKey, fileType, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("presign failed: %w", err)
	}
	return &PresignResult{UploadURL: result.UploadURL, FileURL: result.FileURL, FileKey: result.FileKey}, nil
}

// PresignResult mirrors storage.PresignResult for the module boundary.
type PresignResult struct {
	UploadURL string
	FileURL   string
	FileKey   string
}

// ConfirmUpload creates a file record after a successful upload.
func (s *Service) ConfirmUpload(ctx context.Context, actor mpdomain.Actor, input CreateFileInput) (*mpdomain.PlanFileEntity, error) {
	if input.IsMasterPlan && !actor.CanUploadMasterPlan() {
		return nil, mpdomain.ErrForbiddenAction
	}
	effectiveTeamID := input.TeamID
	if !input.IsMasterPlan {
		if effectiveTeamID == nil {
			effectiveTeamID = actor.TeamID
		}
		if !actor.CanUploadForTeam(effectiveTeamID) {
			return nil, mpdomain.ErrForbiddenAction
		}
	} else {
		effectiveTeamID = nil
	}

	input.TeamID = effectiveTeamID
	input.UploadedByID = actor.UserID

	file, err := s.repo.CreateFile(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("create file failed: %w", err)
	}
	_ = s.repo.CreateFileSizeLog(ctx, file.ID, actor.UserID, input.FileSizeBytes)
	return file, nil
}

// ListFiles returns plan files for the given period, scoped to the actor's team if non-admin.
func (s *Service) ListFiles(ctx context.Context, actor mpdomain.Actor, planID int64, teamID int64, search string) ([]mpdomain.PlanFileEntity, error) {
	effectiveTeamID := teamID
	if !actor.IsAdmin() {
		if actor.TeamID != nil {
			effectiveTeamID = *actor.TeamID
		} else {
			effectiveTeamID = -1
		}
	}
	return s.repo.ListFiles(ctx, planID, effectiveTeamID, search)
}

// GetFile retrieves a single plan file by ID with access control.
func (s *Service) GetFile(ctx context.Context, actor mpdomain.Actor, fileID int64) (*mpdomain.PlanFileEntity, error) {
	if fileID <= 0 {
		return nil, mpdomain.ErrInvalidFileID
	}
	file, err := s.repo.GetFileByID(ctx, fileID)
	if err != nil || file == nil {
		return nil, mpdomain.ErrFileNotFound
	}
	if !file.IsMasterPlan && !actor.CanAccessFile(file.TeamID) {
		return nil, mpdomain.ErrForbiddenAction
	}
	return file, nil
}

// SoftDeleteFile soft-deletes a plan file.
func (s *Service) SoftDeleteFile(ctx context.Context, actor mpdomain.Actor, fileID int64) error {
	if fileID <= 0 {
		return mpdomain.ErrInvalidFileID
	}
	file, err := s.repo.GetFileByID(ctx, fileID)
	if err != nil || file == nil {
		return mpdomain.ErrFileNotFound
	}
	if file.IsDeleted {
		return mpdomain.ErrFileNotDeleted
	}
	if !file.IsMasterPlan && !actor.CanAccessFile(file.TeamID) {
		return mpdomain.ErrForbiddenAction
	}
	return s.repo.SoftDeleteFile(ctx, fileID)
}

// RestoreFile restores a previously soft-deleted plan file.
func (s *Service) RestoreFile(ctx context.Context, actor mpdomain.Actor, fileID int64) (*mpdomain.PlanFileEntity, error) {
	if fileID <= 0 {
		return nil, mpdomain.ErrInvalidFileID
	}
	file, err := s.repo.GetFileByID(ctx, fileID)
	if err != nil || file == nil {
		return nil, mpdomain.ErrFileNotFound
	}
	if !file.IsDeleted {
		return nil, mpdomain.ErrFileNotDeleted
	}
	if !file.IsMasterPlan && !actor.CanAccessFile(file.TeamID) {
		return nil, mpdomain.ErrForbiddenAction
	}
	return s.repo.RestoreFile(ctx, fileID)
}

// HardDeleteFile permanently deletes a plan file and its storage object (admin only).
func (s *Service) HardDeleteFile(ctx context.Context, actor mpdomain.Actor, fileID int64) error {
	if !actor.IsAdmin() {
		return mpdomain.ErrForbiddenAction
	}
	if fileID <= 0 {
		return mpdomain.ErrInvalidFileID
	}
	file, err := s.repo.GetFileByID(ctx, fileID)
	if err != nil || file == nil {
		return mpdomain.ErrFileNotFound
	}
	_ = s.storage.DeleteObject(ctx, file.FileKey)
	return s.repo.HardDeleteFile(ctx, fileID)
}

// ── Submission Status ────────────────────────────────────────────────────────

// GetSubmissionStatus returns each team's submission status for a period (admin only).
func (s *Service) GetSubmissionStatus(ctx context.Context, actor mpdomain.Actor, planID int64) (*SubmissionOverview, error) {
	if !actor.IsAdmin() {
		return nil, mpdomain.ErrForbiddenAction
	}
	teams, err := s.repo.ListTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("list teams failed: %w", err)
	}
	counts, err := s.repo.CountFilesByTeam(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("count files failed: %w", err)
	}
	countMap := make(map[int64]int, len(counts))
	for _, c := range counts {
		countMap[c.TeamID] = c.Count
	}
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
	return &SubmissionOverview{Deadline: "", Teams: entries}, nil
}

// SubmissionOverview holds the submission status for all teams.
type SubmissionOverview struct {
	Deadline string
	Teams    []mpdomain.SubmissionStatusEntry
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func isAllowedType(fileType string, allowed []string) bool {
	for _, a := range allowed {
		if a == fileType {
			return true
		}
	}
	return false
}
