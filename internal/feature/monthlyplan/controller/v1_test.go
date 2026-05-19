package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-hotlines3/internal/feature/monthlyplan/entity"
	"backend-hotlines3/internal/feature/monthlyplan/repository"
	"backend-hotlines3/internal/feature/monthlyplan/service"

	"github.com/gin-gonic/gin"
)

func TestGetYearOverviewReturnsAllMonthsWithLockActionsAndFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &overviewFakeRepo{
		settings: &entity.SettingsEntity{LockDay: 23, AllowedFileTypes: []string{"application/pdf"}},
		filesByPlan: map[int64][]entity.PlanFileEntity{
			6: {{ID: 601, MonthlyPlanID: 6, UploadedByID: 1, FileName: "june-master.pdf", FileURL: "https://files.example/june.pdf", IsMasterPlan: true}},
		},
	}
	svc := service.NewService(repo, &overviewFakeStorage{})
	svc.SetClockForTest(func() time.Time { return time.Date(2026, 5, 24, 8, 0, 0, 0, time.UTC) })
	ctrl := NewController(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "team_lead")
		c.Set("team_id", int64(7))
	})
	r.GET("/v1/monthly-plans/:year/overview", ctrl.GetYearOverview)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/monthly-plans/2026/overview", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Year   int `json:"year"`
			Months []struct {
				Month    int    `json:"month"`
				Deadline string `json:"deadline"`
				IsLocked bool   `json:"isLocked"`
				Actions  struct {
					CanUpload bool `json:"canUpload"`
				} `json:"actions"`
				Files []struct {
					ID           int64  `json:"id"`
					FileName     string `json:"fileName"`
					IsMasterPlan bool   `json:"isMasterPlan"`
				} `json:"files"`
			} `json:"months"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || body.Data.Year != 2026 || len(body.Data.Months) != 12 {
		t.Fatalf("unexpected overview envelope: %+v", body)
	}
	june := body.Data.Months[5]
	if june.Month != 6 || june.Deadline != "2026-05-23" || !june.IsLocked || june.Actions.CanUpload {
		t.Fatalf("unexpected June lock/actions shape: %+v", june)
	}
	if len(june.Files) != 1 || june.Files[0].FileName != "june-master.pdf" || !june.Files[0].IsMasterPlan {
		t.Fatalf("unexpected June files: %+v", june.Files)
	}
}

type overviewFakeRepo struct {
	settings    *entity.SettingsEntity
	filesByPlan map[int64][]entity.PlanFileEntity
}

func (r *overviewFakeRepo) FindPeriod(context.Context, repository.PeriodFindQuery) (*entity.Entity, error) {
	return nil, nil
}

func (r *overviewFakeRepo) GetPeriodByID(_ context.Context, id int64) (*entity.Entity, error) {
	return &entity.Entity{ID: id, Year: 2026, Month: int(id)}, nil
}

func (r *overviewFakeRepo) FindOrCreatePeriod(_ context.Context, year, month int) (*entity.Entity, error) {
	return &entity.Entity{ID: int64(month), Year: year, Month: month, CreatedAt: time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)}, nil
}

func (r *overviewFakeRepo) FindOrCreatePeriodsForYear(_ context.Context, year int) ([]entity.Entity, error) {
	periods := make([]entity.Entity, 0, 12)
	for month := 1; month <= 12; month++ {
		periods = append(periods, entity.Entity{ID: int64(month), Year: year, Month: month, CreatedAt: time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)})
	}
	return periods, nil
}

func (r *overviewFakeRepo) GetOrCreateSettings(context.Context) (*entity.SettingsEntity, error) {
	return r.settings, nil
}

func (r *overviewFakeRepo) UpdateSettings(context.Context, *entity.SettingsEntity) error { return nil }

func (r *overviewFakeRepo) ListFiles(_ context.Context, q repository.PlanFileListQuery) ([]entity.PlanFileEntity, error) {
	return r.filesByPlan[q.MonthlyPlanID], nil
}

func (r *overviewFakeRepo) ListFilesByPlanIDs(_ context.Context, q repository.PlanFileBatchListQuery) (map[int64][]entity.PlanFileEntity, error) {
	result := make(map[int64][]entity.PlanFileEntity, len(q.MonthlyPlanIDs))
	for _, id := range q.MonthlyPlanIDs {
		if files := r.filesByPlan[id]; files != nil {
			result[id] = files
		}
	}
	return result, nil
}

func (r *overviewFakeRepo) GetFileByID(context.Context, int64) (*entity.PlanFileEntity, error) {
	return nil, nil
}
func (r *overviewFakeRepo) CreateFile(context.Context, repository.PlanFileCreateInput) (*entity.PlanFileEntity, error) {
	return nil, nil
}
func (r *overviewFakeRepo) SoftDeleteFile(context.Context, repository.PlanFileSoftDeleteInput) error {
	return nil
}
func (r *overviewFakeRepo) RestoreFile(context.Context, repository.PlanFileRestoreInput) (*entity.PlanFileEntity, error) {
	return nil, nil
}
func (r *overviewFakeRepo) HardDeleteFile(context.Context, repository.PlanFileHardDeleteInput) error {
	return nil
}
func (r *overviewFakeRepo) CreateFileSizeLog(context.Context, repository.FileSizeLogCreateInput) error {
	return nil
}
func (r *overviewFakeRepo) CountFilesByTeam(context.Context, repository.SubmissionQuery) ([]repository.TeamCountRow, error) {
	return nil, nil
}
func (r *overviewFakeRepo) ListTeams(context.Context) ([]repository.TeamInfo, error) { return nil, nil }

type overviewFakeStorage struct{}

func (s *overviewFakeStorage) PresignUpload(context.Context, string, string, time.Duration) (*repository.PresignResult, error) {
	return nil, nil
}
func (s *overviewFakeStorage) PresignDownload(context.Context, string, time.Duration) (string, error) {
	return "", nil
}
func (s *overviewFakeStorage) DeleteObject(context.Context, string) error { return nil }
func (s *overviewFakeStorage) PublicURL(fileKey string) string            { return fileKey }

func TestGetDownloadURLBlocksViewerEvenWithMatchingTeam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	teamID := int64(7)

	baseRepo := &overviewFakeRepo{
		settings:    &entity.SettingsEntity{LockDay: 23},
		filesByPlan: nil,
	}

	// Override GetFileByID to return a file belonging to the viewer's team
	fileEntity := &entity.PlanFileEntity{
		ID:            42,
		MonthlyPlanID: 6,
		TeamID:        &teamID,
		UploadedByID:  1,
		FileName:      "team-plan.pdf",
		FileKey:       "uploads/team-plan.pdf",
		FileURL:       "https://files.example/team-plan.pdf",
		IsMasterPlan:  false,
	}
	dlRepo := &downloadTestFakeRepo{inner: baseRepo, file: fileEntity}
	dlStorage := &downloadTestFakeStorage{}
	svc := service.NewService(dlRepo, dlStorage)
	ctrl := NewController(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(99))
		c.Set("role", "viewer")
		c.Set("team_id", uint(teamID))
	})
	r.GET("/v1/monthly-plans/files/:id/download", ctrl.GetDownloadURL)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/monthly-plans/files/42/download", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for viewer download, got %d body=%s", w.Code, w.Body.String())
	}
}

type downloadTestFakeRepo struct {
	inner *overviewFakeRepo
	file  *entity.PlanFileEntity
}

func (r *downloadTestFakeRepo) FindPeriod(ctx context.Context, q repository.PeriodFindQuery) (*entity.Entity, error) {
	return r.inner.FindPeriod(ctx, q)
}
func (r *downloadTestFakeRepo) GetPeriodByID(ctx context.Context, id int64) (*entity.Entity, error) {
	return r.inner.GetPeriodByID(ctx, id)
}
func (r *downloadTestFakeRepo) FindOrCreatePeriod(ctx context.Context, year, month int) (*entity.Entity, error) {
	return r.inner.FindOrCreatePeriod(ctx, year, month)
}
func (r *downloadTestFakeRepo) FindOrCreatePeriodsForYear(ctx context.Context, year int) ([]entity.Entity, error) {
	return r.inner.FindOrCreatePeriodsForYear(ctx, year)
}
func (r *downloadTestFakeRepo) GetOrCreateSettings(ctx context.Context) (*entity.SettingsEntity, error) {
	return r.inner.GetOrCreateSettings(ctx)
}
func (r *downloadTestFakeRepo) UpdateSettings(ctx context.Context, s *entity.SettingsEntity) error {
	return r.inner.UpdateSettings(ctx, s)
}
func (r *downloadTestFakeRepo) ListFiles(ctx context.Context, q repository.PlanFileListQuery) ([]entity.PlanFileEntity, error) {
	return r.inner.ListFiles(ctx, q)
}
func (r *downloadTestFakeRepo) ListFilesByPlanIDs(ctx context.Context, q repository.PlanFileBatchListQuery) (map[int64][]entity.PlanFileEntity, error) {
	return r.inner.ListFilesByPlanIDs(ctx, q)
}
func (r *downloadTestFakeRepo) GetFileByID(context.Context, int64) (*entity.PlanFileEntity, error) {
	return r.file, nil
}
func (r *downloadTestFakeRepo) CreateFile(ctx context.Context, i repository.PlanFileCreateInput) (*entity.PlanFileEntity, error) {
	return r.inner.CreateFile(ctx, i)
}
func (r *downloadTestFakeRepo) SoftDeleteFile(ctx context.Context, i repository.PlanFileSoftDeleteInput) error {
	return r.inner.SoftDeleteFile(ctx, i)
}
func (r *downloadTestFakeRepo) RestoreFile(ctx context.Context, i repository.PlanFileRestoreInput) (*entity.PlanFileEntity, error) {
	return r.inner.RestoreFile(ctx, i)
}
func (r *downloadTestFakeRepo) HardDeleteFile(ctx context.Context, i repository.PlanFileHardDeleteInput) error {
	return r.inner.HardDeleteFile(ctx, i)
}
func (r *downloadTestFakeRepo) CreateFileSizeLog(ctx context.Context, i repository.FileSizeLogCreateInput) error {
	return r.inner.CreateFileSizeLog(ctx, i)
}
func (r *downloadTestFakeRepo) CountFilesByTeam(ctx context.Context, q repository.SubmissionQuery) ([]repository.TeamCountRow, error) {
	return r.inner.CountFilesByTeam(ctx, q)
}
func (r *downloadTestFakeRepo) ListTeams(ctx context.Context) ([]repository.TeamInfo, error) {
	return r.inner.ListTeams(ctx)
}

type downloadTestFakeStorage struct{}

func (s *downloadTestFakeStorage) PresignUpload(context.Context, string, string, time.Duration) (*repository.PresignResult, error) {
	return nil, nil
}
func (s *downloadTestFakeStorage) PresignDownload(context.Context, string, time.Duration) (string, error) {
	return "https://presigned.example/download", nil
}
func (s *downloadTestFakeStorage) DeleteObject(context.Context, string) error { return nil }
func (s *downloadTestFakeStorage) PublicURL(fileKey string) string            { return fileKey }
