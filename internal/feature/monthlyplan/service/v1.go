package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	mpdomain "backend-hotlines3/internal/feature/monthlyplan/entity"
	"backend-hotlines3/internal/feature/monthlyplan/repository"

	"golang.org/x/sync/errgroup"
)

// Service owns the monthly-plan usecase composition.
// ── Settings ─────────────────────────────────────────────────────────────────

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
	if actor.Role != "super_admin" {
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

// CanUploadForPeriod evaluates the previous-month lock window using the DB settings.
func (s *Service) CanUploadForPeriod(ctx context.Context, actor mpdomain.Actor, year, month int) (bool, string, error) {
	if year < 2000 || year > 2100 || month < 1 || month > 12 {
		return false, "", mpdomain.ErrInvalidPeriod
	}
	settings, err := s.repo.GetOrCreateSettings(ctx)
	if err != nil {
		return false, "", mpdomain.ErrSettingsLoadFail
	}
	deadline := submissionDeadline(year, month, settings.LockDay)
	locked := s.clock().After(deadline)
	canSubmit := actor.CanUploadMasterPlan() || actor.CanUploadForTeam(actor.TeamID)
	if !canSubmit {
		return false, deadline.Format("2006-01-02"), nil
	}
	if !locked {
		return true, deadline.Format("2006-01-02"), nil
	}
	return actor.CanUploadAfterLock(settings.AdminCanUploadAfterLock), deadline.Format("2006-01-02"), nil
}

type MonthOverview struct {
	Period    mpdomain.Entity
	Deadline  string
	IsLocked  bool
	Status    string
	CanUpload bool
	Files     []mpdomain.PlanFileEntity
}

type YearOverview struct {
	Year   int
	Months []MonthOverview
}

// GetYearOverview returns all 12 monthly plan buckets for a year.
func (s *Service) GetYearOverview(ctx context.Context, actor mpdomain.Actor, year int) (*YearOverview, error) {
	if year < 2000 || year > 2100 {
		return nil, mpdomain.ErrInvalidPeriod
	}
	settings, err := s.repo.GetOrCreateSettings(ctx)
	if err != nil {
		return nil, mpdomain.ErrSettingsLoadFail
	}
	periods, err := s.repo.FindOrCreatePeriodsForYear(ctx, year)
	if err != nil {
		return nil, err
	}
	periodByMonth := make(map[int]mpdomain.Entity, len(periods))
	planIDs := make([]int64, 0, len(periods))
	for _, period := range periods {
		periodByMonth[period.Month] = period
		planIDs = append(planIDs, period.ID)
	}

	effectiveTeamID := int64(0)
	if actor.Role != "super_admin" {
		if actor.TeamID != nil {
			effectiveTeamID = *actor.TeamID
		} else {
			effectiveTeamID = -1
		}
	}
	filesByPlan, err := s.repo.ListFilesByPlanIDs(ctx, repository.PlanFileBatchListQuery{MonthlyPlanIDs: planIDs, TeamID: effectiveTeamID})
	if err != nil {
		return nil, err
	}

	overview := &YearOverview{Year: year, Months: make([]MonthOverview, 0, 12)}
	for month := 1; month <= 12; month++ {
		period, ok := periodByMonth[month]
		if !ok {
			return nil, fmt.Errorf("missing monthly plan period for %d-%02d", year, month)
		}
		deadline := submissionDeadline(year, month, settings.LockDay)
		locked := s.clock().After(deadline)
		canSubmit := actor.CanUploadMasterPlan() || actor.CanUploadForTeam(actor.TeamID)
		canUpload := canSubmit && (!locked || actor.CanUploadAfterLock(settings.AdminCanUploadAfterLock))
		files := filesByPlan[period.ID]
		status := "open"
		if locked {
			status = "locked"
		}
		if len(files) > 0 {
			status = "has_files"
		}
		overview.Months = append(overview.Months, MonthOverview{
			Period:    period,
			Deadline:  deadline.Format("2006-01-02"),
			IsLocked:  locked,
			Status:    status,
			CanUpload: canUpload,
			Files:     files,
		})
	}
	return overview, nil
}

type ConvertApprovedToPlanningInput struct {
	Year            int
	Month           int
	ApprovedFileID  int64
	SelectedTeamIDs []int64
}

type ConvertApprovedToPlanningResult struct {
	PlanningItemsCreated int
	SourceFileID         int64
	ConvertedAt          string
}

// ConvertApprovedToPlanning validates an approved/master monthly-plan file as a planning source.
// Planning Calendar is currently a projection over monthly-plan files, so this operation is
// intentionally idempotent: once the approved file is accepted as the source, selected teams
// are counted as planning items and the calendar projection can render from the file metadata.
func (s *Service) ConvertApprovedToPlanning(ctx context.Context, actor mpdomain.Actor, input ConvertApprovedToPlanningInput) (*ConvertApprovedToPlanningResult, error) {
	if !actor.CanUploadMasterPlan() {
		return nil, mpdomain.ErrForbiddenAction
	}
	if input.Year < 2000 || input.Year > 2100 || input.Month < 1 || input.Month > 12 {
		return nil, mpdomain.ErrInvalidPeriod
	}
	if input.ApprovedFileID <= 0 || len(input.SelectedTeamIDs) == 0 {
		return nil, mpdomain.ErrInvalidFileID
	}
	period, err := s.repo.FindOrCreatePeriod(ctx, input.Year, input.Month)
	if err != nil {
		return nil, err
	}
	if period == nil {
		return nil, mpdomain.ErrPeriodNotFound
	}
	file, err := s.repo.GetFileByID(ctx, input.ApprovedFileID)
	if err != nil || file == nil {
		return nil, mpdomain.ErrFileNotFound
	}
	if !file.IsMasterPlan || file.IsDeleted || file.MonthlyPlanID != period.ID {
		return nil, mpdomain.ErrFileNotFound
	}
	seen := make(map[int64]struct{}, len(input.SelectedTeamIDs))
	for _, teamID := range input.SelectedTeamIDs {
		if teamID <= 0 {
			return nil, mpdomain.ErrInvalidFileID
		}
		seen[teamID] = struct{}{}
	}
	return &ConvertApprovedToPlanningResult{
		PlanningItemsCreated: len(seen),
		SourceFileID:         file.ID,
		ConvertedAt:          s.clock().UTC().Format(time.RFC3339),
	}, nil
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
func (s *Service) ConfirmUpload(ctx context.Context, actor mpdomain.Actor, input repository.PlanFileCreateInput) (*mpdomain.PlanFileEntity, error) {
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
	if input.UploadedByID == 0 {
		input.UploadedByID = actor.UserID
	}

	file, err := s.repo.CreateFile(ctx, repository.PlanFileCreateInput{
		MonthlyPlanID: input.MonthlyPlanID,
		TeamID:        input.TeamID,
		UploadedByID:  input.UploadedByID,
		FileKey:       input.FileKey,
		FileURL:       input.FileURL,
		FileName:      input.FileName,
		FileSizeBytes: input.FileSizeBytes,
		Description:   input.Description,
		WorkStartDate: input.WorkStartDate,
		WorkEndDate:   input.WorkEndDate,
		Destination:   input.Destination,
		Remarks:       input.Remarks,
		IsMasterPlan:  input.IsMasterPlan,
	})
	if err != nil {
		return nil, fmt.Errorf("create file failed: %w", err)
	}
	_ = s.repo.CreateFileSizeLog(ctx, repository.FileSizeLogCreateInput{
		PlanFileID:    file.ID,
		UploadedByID:  actor.UserID,
		FileSizeBytes: input.FileSizeBytes,
	})
	return file, nil
}

// ListFiles returns plan files for the given period, scoped to the actor's team if non-admin.
func (s *Service) ListFiles(ctx context.Context, actor mpdomain.Actor, planID int64, teamID int64, search string) ([]mpdomain.PlanFileEntity, error) {
	effectiveTeamID := teamID
	if actor.Role != "super_admin" {
		if actor.TeamID != nil {
			effectiveTeamID = *actor.TeamID
		} else {
			effectiveTeamID = -1
		}
	}
	return s.repo.ListFiles(ctx, repository.PlanFileListQuery{
		MonthlyPlanID: planID,
		TeamID:        effectiveTeamID,
		Search:        search,
	})
}

// GetFile retrieves a single plan file by ID with access control.
func (s *Service) GetFile(ctx context.Context, actor mpdomain.Actor, fileID int64) (*mpdomain.PlanFileEntity, error) {
	file, err := s.repo.GetFileByID(ctx, fileID)
	if err != nil || file == nil {
		return nil, mpdomain.ErrFileNotFound
	}
	if !file.IsMasterPlan && !actor.CanAccessFile(file.TeamID) {
		return nil, mpdomain.ErrForbiddenAction
	}
	return file, nil
}

// PresignDownload generates a download URL for an existing file key.
func (s *Service) PresignDownload(ctx context.Context, fileKey string) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("download not configured")
	}
	return s.storage.PresignDownload(ctx, fileKey, 15*time.Minute)
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
	if file.IsMasterPlan {
		if !actor.CanUploadMasterPlan() {
			return mpdomain.ErrForbiddenAction
		}
	} else if !actor.CanAccessFile(file.TeamID) {
		return mpdomain.ErrForbiddenAction
	}
	return s.repo.SoftDeleteFile(ctx, repository.PlanFileSoftDeleteInput{ID: fileID})
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
	return s.repo.RestoreFile(ctx, repository.PlanFileRestoreInput{ID: fileID})
}

// HardDeleteFile permanently deletes a plan file and its storage object (admin only).
func (s *Service) HardDeleteFile(ctx context.Context, actor mpdomain.Actor, fileID int64) error {
	if actor.Role != "super_admin" {
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
	return s.repo.HardDeleteFile(ctx, repository.PlanFileHardDeleteInput{ID: fileID})
}

// ── Submission Status ────────────────────────────────────────────────────────

// GetSubmissionStatus returns each team's submission status for a period.
// Route-level auth is required; team_lead/user can view every team row for awareness
// while file upload/download permissions remain enforced separately by team scope.
func (s *Service) GetSubmissionStatus(ctx context.Context, actor mpdomain.Actor, planID int64) (*SubmissionOverview, error) {
	var (
		settings *mpdomain.SettingsEntity
		period   *mpdomain.Entity
		teams    []repository.TeamInfo
		counts   []repository.TeamCountRow
	)
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		settings, err = s.repo.GetOrCreateSettings(gCtx)
		if err != nil {
			return mpdomain.ErrSettingsLoadFail
		}
		return nil
	})
	g.Go(func() error {
		var err error
		period, err = s.repo.GetPeriodByID(gCtx, planID)
		if err != nil {
			return fmt.Errorf("get period failed: %w", err)
		}
		if period == nil {
			return mpdomain.ErrPeriodNotFound
		}
		return nil
	})
	g.Go(func() error {
		var err error
		teams, err = s.repo.ListTeams(gCtx)
		if err != nil {
			return fmt.Errorf("list teams failed: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		counts, err = s.repo.CountFilesByTeam(gCtx, repository.SubmissionQuery{PlanID: planID})
		if err != nil {
			return fmt.Errorf("count files failed: %w", err)
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	if actor.Role != "super_admin" {
		if actor.TeamID == nil {
			teams = nil
			counts = nil
		} else {
			filteredTeams := make([]repository.TeamInfo, 0, 1)
			for _, team := range teams {
				if team.ID == *actor.TeamID {
					filteredTeams = append(filteredTeams, team)
					break
				}
			}
			teams = filteredTeams
			filteredCounts := make([]repository.TeamCountRow, 0, 1)
			for _, count := range counts {
				if count.TeamID == *actor.TeamID {
					filteredCounts = append(filteredCounts, count)
					break
				}
			}
			counts = filteredCounts
		}
	}
	deadline := submissionDeadline(period.Year, period.Month, settings.LockDay)
	missed := s.clock().After(deadline)
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
		} else if missed {
			status = "missed"
		}
		entries = append(entries, mpdomain.SubmissionStatusEntry{
			TeamID:    team.ID,
			TeamName:  team.Name,
			Status:    status,
			FileCount: count,
		})
	}
	return &SubmissionOverview{Deadline: deadline.Format("2006-01-02"), Teams: entries}, nil
}

func submissionDeadline(planYear, planMonth, lockDay int) time.Time {
	if lockDay < 1 {
		lockDay = 1
	}
	planMonthStart := time.Date(planYear, time.Month(planMonth), 1, 23, 59, 59, 0, time.UTC)
	deadlineMonthStart := planMonthStart.AddDate(0, -1, 0)
	lastDay := deadlineMonthStart.AddDate(0, 1, -1).Day()
	if lockDay > lastDay {
		lockDay = lastDay
	}
	return time.Date(deadlineMonthStart.Year(), deadlineMonthStart.Month(), lockDay, 23, 59, 59, 0, time.UTC)
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
