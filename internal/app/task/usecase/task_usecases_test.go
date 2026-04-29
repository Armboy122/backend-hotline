// Package usecase_test tests the observable behavior of all task use cases
// (Get, Create, Update, Delete, ListByTeam, ListByFilter) using a configurable fake repo.
package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	taskusecase "backend-hotlines3/internal/app/task/usecase"
	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/port/outbound/repository"
)

// ── Configurable fake repo ────────────────────────────────────────────────────

type configurableFakeRepo struct {
	// return values
	getByIDResult  *taskdomain.Entity
	getByIDErr     error
	createResult   *taskdomain.Entity
	createErr      error
	updateResult   *taskdomain.Entity
	updateErr      error
	softDeleteErr  error
	listByTeamRes  []taskdomain.Entity
	listByTeamTot  int64
	listByTeamErr  error
	listByFilterRes []taskdomain.Entity
	listByFilterErr error

	// captured inputs
	capturedGetByID    repository.TaskGetQuery
	capturedCreate     repository.TaskCreateInput
	capturedUpdate     repository.TaskUpdateInput
	capturedDelete     repository.TaskDeleteCommand
	capturedListByTeam repository.TaskByTeamQuery
	capturedListByFltr repository.TaskByFilterQuery
}

func (f *configurableFakeRepo) List(_ context.Context, _ repository.TaskListQuery) ([]taskdomain.Entity, int64, error) {
	return nil, 0, nil
}
func (f *configurableFakeRepo) GetByID(_ context.Context, q repository.TaskGetQuery) (*taskdomain.Entity, error) {
	f.capturedGetByID = q
	return f.getByIDResult, f.getByIDErr
}
func (f *configurableFakeRepo) Create(_ context.Context, inp repository.TaskCreateInput) (*taskdomain.Entity, error) {
	f.capturedCreate = inp
	return f.createResult, f.createErr
}
func (f *configurableFakeRepo) Update(_ context.Context, inp repository.TaskUpdateInput) (*taskdomain.Entity, error) {
	f.capturedUpdate = inp
	return f.updateResult, f.updateErr
}
func (f *configurableFakeRepo) SoftDelete(_ context.Context, cmd repository.TaskDeleteCommand) error {
	f.capturedDelete = cmd
	return f.softDeleteErr
}
func (f *configurableFakeRepo) ListByTeam(_ context.Context, q repository.TaskByTeamQuery) ([]taskdomain.Entity, int64, error) {
	f.capturedListByTeam = q
	return f.listByTeamRes, f.listByTeamTot, f.listByTeamErr
}
func (f *configurableFakeRepo) ListByFilter(_ context.Context, q repository.TaskByFilterQuery) ([]taskdomain.Entity, error) {
	f.capturedListByFltr = q
	return f.listByFilterRes, f.listByFilterErr
}

// helper to build a domain entity
func sampleEntity() *taskdomain.Entity {
	now := time.Now()
	return &taskdomain.Entity{
		ID:          1,
		WorkDate:    time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		TeamID:      10,
		JobTypeID:   20,
		JobDetailID: 30,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ── GetTask ───────────────────────────────────────────────────────────────────

func TestGetTask_InvalidID(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewGetTaskUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.GetTaskInput{ID: 0})
	if !errors.Is(err, taskusecase.ErrInvalidTaskID) {
		t.Errorf("got %v, want ErrInvalidTaskID", err)
	}
}

func TestGetTask_NegativeID(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewGetTaskUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.GetTaskInput{ID: -5})
	if !errors.Is(err, taskusecase.ErrInvalidTaskID) {
		t.Errorf("got %v, want ErrInvalidTaskID", err)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	repo := &configurableFakeRepo{getByIDResult: nil} // nil = not found
	uc := taskusecase.NewGetTaskUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.GetTaskInput{ID: 1})
	if !errors.Is(err, taskusecase.ErrTaskNotFound) {
		t.Errorf("got %v, want ErrTaskNotFound", err)
	}
}

func TestGetTask_Success(t *testing.T) {
	want := sampleEntity()
	repo := &configurableFakeRepo{getByIDResult: want}
	uc := taskusecase.NewGetTaskUseCase(repo)

	got, err := uc.Execute(context.Background(), taskusecase.GetTaskInput{ID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("got ID=%d, want %d", got.ID, want.ID)
	}
	if repo.capturedGetByID.ID != 1 {
		t.Errorf("repo received ID=%d, want 1", repo.capturedGetByID.ID)
	}
}

// ── CreateTask ────────────────────────────────────────────────────────────────

func TestCreateTask_MissingWorkDate(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewCreateTaskUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.CreateTaskInput{
		TeamID: 1, JobTypeID: 1, JobDetailID: 1,
	})
	if !errors.Is(err, taskusecase.ErrWorkDateRequired) {
		t.Errorf("got %v, want ErrWorkDateRequired", err)
	}
}

func TestCreateTask_MissingTeamID(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewCreateTaskUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.CreateTaskInput{
		WorkDate: time.Now(), JobTypeID: 1, JobDetailID: 1,
	})
	if !errors.Is(err, taskusecase.ErrTeamIDRequired) {
		t.Errorf("got %v, want ErrTeamIDRequired", err)
	}
}

func TestCreateTask_MissingJobTypeID(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewCreateTaskUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.CreateTaskInput{
		WorkDate: time.Now(), TeamID: 1, JobDetailID: 1,
	})
	if !errors.Is(err, taskusecase.ErrJobTypeIDRequired) {
		t.Errorf("got %v, want ErrJobTypeIDRequired", err)
	}
}

func TestCreateTask_MissingJobDetailID(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewCreateTaskUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.CreateTaskInput{
		WorkDate: time.Now(), TeamID: 1, JobTypeID: 1,
	})
	if !errors.Is(err, taskusecase.ErrJobDetailIDRequired) {
		t.Errorf("got %v, want ErrJobDetailIDRequired", err)
	}
}

func TestCreateTask_Success(t *testing.T) {
	want := sampleEntity()
	repo := &configurableFakeRepo{createResult: want}
	uc := taskusecase.NewCreateTaskUseCase(repo)

	wd := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	got, err := uc.Execute(context.Background(), taskusecase.CreateTaskInput{
		WorkDate:    wd,
		TeamID:      10,
		JobTypeID:   20,
		JobDetailID: 30,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("got ID=%d, want %d", got.ID, want.ID)
	}
	if repo.capturedCreate.TeamID != 10 {
		t.Errorf("repo received TeamID=%d, want 10", repo.capturedCreate.TeamID)
	}
}

func TestCreateTask_RepoError(t *testing.T) {
	repo := &configurableFakeRepo{createErr: errors.New("db insert fail")}
	uc := taskusecase.NewCreateTaskUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.CreateTaskInput{
		WorkDate: time.Now(), TeamID: 1, JobTypeID: 1, JobDetailID: 1,
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ── UpdateTask ────────────────────────────────────────────────────────────────

func TestUpdateTask_InvalidID(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewUpdateTaskUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.UpdateTaskInput{ID: 0})
	if !errors.Is(err, taskusecase.ErrInvalidTaskID) {
		t.Errorf("got %v, want ErrInvalidTaskID", err)
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	repo := &configurableFakeRepo{updateResult: nil} // nil = not found
	uc := taskusecase.NewUpdateTaskUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.UpdateTaskInput{ID: 99})
	if !errors.Is(err, taskusecase.ErrTaskNotFound) {
		t.Errorf("got %v, want ErrTaskNotFound", err)
	}
}

func TestUpdateTask_Success(t *testing.T) {
	want := sampleEntity()
	repo := &configurableFakeRepo{updateResult: want}
	uc := taskusecase.NewUpdateTaskUseCase(repo)

	got, err := uc.Execute(context.Background(), taskusecase.UpdateTaskInput{ID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("got ID=%d, want %d", got.ID, want.ID)
	}
}

func TestUpdateTask_RepoError(t *testing.T) {
	repo := &configurableFakeRepo{updateErr: errors.New("db update fail")}
	uc := taskusecase.NewUpdateTaskUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.UpdateTaskInput{ID: 1})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ── DeleteTask ────────────────────────────────────────────────────────────────

func TestDeleteTask_InvalidID(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewDeleteTaskUseCase(repo)

	err := uc.Execute(context.Background(), taskusecase.DeleteTaskInput{ID: 0})
	if !errors.Is(err, taskusecase.ErrInvalidTaskID) {
		t.Errorf("got %v, want ErrInvalidTaskID", err)
	}
}

func TestDeleteTask_Success(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewDeleteTaskUseCase(repo)

	err := uc.Execute(context.Background(), taskusecase.DeleteTaskInput{ID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.capturedDelete.ID != 1 {
		t.Errorf("repo received ID=%d, want 1", repo.capturedDelete.ID)
	}
}

func TestDeleteTask_RepoError(t *testing.T) {
	repo := &configurableFakeRepo{softDeleteErr: errors.New("db delete fail")}
	uc := taskusecase.NewDeleteTaskUseCase(repo)

	err := uc.Execute(context.Background(), taskusecase.DeleteTaskInput{ID: 1})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ── ListTasksByTeam ───────────────────────────────────────────────────────────

func TestListTasksByTeam_PaginationNormalization(t *testing.T) {
	cases := []struct {
		name       string
		inputPage  int
		inputLimit int
		wantPage   int
		wantLimit  int
	}{
		{"page 0 → 1", 0, 50, 1, 50},
		{"page -5 → 1", -5, 50, 1, 50},
		{"limit 0 → 100", 1, 0, 1, 100},
		{"limit 501 → 100", 1, 501, 1, 100},
		{"limit 500 passes", 1, 500, 1, 500},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &configurableFakeRepo{}
			uc := taskusecase.NewListTasksByTeamUseCase(repo)

			out, err := uc.Execute(context.Background(), taskusecase.ListTasksByTeamInput{
				Page: tc.inputPage, Limit: tc.inputLimit,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.Page != tc.wantPage {
				t.Errorf("Page: got %d, want %d", out.Page, tc.wantPage)
			}
			if out.Limit != tc.wantLimit {
				t.Errorf("Limit: got %d, want %d", out.Limit, tc.wantLimit)
			}
		})
	}
}

func TestListTasksByTeam_FiltersPassedToRepo(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewListTasksByTeamUseCase(repo)

	wd := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	teamID := int64(10)
	jobTypeID := int64(20)

	_, err := uc.Execute(context.Background(), taskusecase.ListTasksByTeamInput{
		Page: 1, Limit: 50, WorkDate: &wd, TeamID: &teamID, JobTypeID: &jobTypeID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	q := repo.capturedListByTeam
	if q.WorkDate == nil || !q.WorkDate.Equal(wd) {
		t.Errorf("WorkDate: got %v, want %v", q.WorkDate, wd)
	}
	if q.TeamID == nil || *q.TeamID != 10 {
		t.Errorf("TeamID: got %v, want 10", q.TeamID)
	}
	if q.JobTypeID == nil || *q.JobTypeID != 20 {
		t.Errorf("JobTypeID: got %v, want 20", q.JobTypeID)
	}
}

func TestListTasksByTeam_RepoError(t *testing.T) {
	repo := &configurableFakeRepo{listByTeamErr: errors.New("db fail")}
	uc := taskusecase.NewListTasksByTeamUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.ListTasksByTeamInput{Page: 1, Limit: 10})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestListTasksByTeam_TotalFromRepo(t *testing.T) {
	repo := &configurableFakeRepo{listByTeamTot: 42}
	uc := taskusecase.NewListTasksByTeamUseCase(repo)

	out, _ := uc.Execute(context.Background(), taskusecase.ListTasksByTeamInput{Page: 1, Limit: 10})
	if out.Total != 42 {
		t.Errorf("got Total=%d, want 42", out.Total)
	}
}

// ── ListTasksByFilter ─────────────────────────────────────────────────────────

func TestListTasksByFilter_MissingYear(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewListTasksByFilterUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.ListTasksByFilterInput{
		Month: "06",
	})
	if !errors.Is(err, taskusecase.ErrYearMonthRequired) {
		t.Errorf("got %v, want ErrYearMonthRequired", err)
	}
}

func TestListTasksByFilter_MissingMonth(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewListTasksByFilterUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.ListTasksByFilterInput{
		Year: "2024",
	})
	if !errors.Is(err, taskusecase.ErrYearMonthRequired) {
		t.Errorf("got %v, want ErrYearMonthRequired", err)
	}
}

func TestListTasksByFilter_Success(t *testing.T) {
	entities := []taskdomain.Entity{*sampleEntity()}
	repo := &configurableFakeRepo{listByFilterRes: entities}
	uc := taskusecase.NewListTasksByFilterUseCase(repo)

	out, err := uc.Execute(context.Background(), taskusecase.ListTasksByFilterInput{
		Year: "2024", Month: "06",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Tasks) != 1 {
		t.Errorf("got %d tasks, want 1", len(out.Tasks))
	}
	if repo.capturedListByFltr.Year != "2024" {
		t.Errorf("repo Year=%s, want 2024", repo.capturedListByFltr.Year)
	}
	if repo.capturedListByFltr.Month != "06" {
		t.Errorf("repo Month=%s, want 06", repo.capturedListByFltr.Month)
	}
}

func TestListTasksByFilter_WithTeamID(t *testing.T) {
	repo := &configurableFakeRepo{}
	uc := taskusecase.NewListTasksByFilterUseCase(repo)

	teamID := int64(10)
	_, _ = uc.Execute(context.Background(), taskusecase.ListTasksByFilterInput{
		Year: "2024", Month: "06", TeamID: &teamID,
	})
	if repo.capturedListByFltr.TeamID == nil || *repo.capturedListByFltr.TeamID != 10 {
		t.Errorf("TeamID: got %v, want 10", repo.capturedListByFltr.TeamID)
	}
}

func TestListTasksByFilter_RepoError(t *testing.T) {
	repo := &configurableFakeRepo{listByFilterErr: errors.New("db fail")}
	uc := taskusecase.NewListTasksByFilterUseCase(repo)

	_, err := uc.Execute(context.Background(), taskusecase.ListTasksByFilterInput{
		Year: "2024", Month: "06",
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
}
