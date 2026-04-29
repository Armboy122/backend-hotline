package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend-hotlines3/internal/feature/task/entity"
	taskrepository "backend-hotlines3/internal/feature/task/repository"
)

type fakeRepository struct {
	capturedList     taskrepository.ListQuery
	capturedGetByID  taskrepository.GetQuery
	capturedCreate   taskrepository.CreateInput
	capturedUpdate   taskrepository.UpdateInput
	capturedDelete   taskrepository.DeleteCommand
	capturedByTeam   taskrepository.ByTeamQuery
	capturedByFilter taskrepository.ByFilterQuery

	listTasks []entity.Task
	listTotal int64
	listErr   error

	getResult    *entity.Task
	getErr       error
	createResult *entity.Task
	createErr    error
	updateResult *entity.Task
	updateErr    error
	deleteErr    error

	byTeamTasks []entity.Task
	byTeamTotal int64
	byTeamErr   error

	byFilterTasks []entity.Task
	byFilterErr   error
}

func (f *fakeRepository) List(_ context.Context, query taskrepository.ListQuery) ([]entity.Task, int64, error) {
	f.capturedList = query
	return f.listTasks, f.listTotal, f.listErr
}
func (f *fakeRepository) GetByID(_ context.Context, query taskrepository.GetQuery) (*entity.Task, error) {
	f.capturedGetByID = query
	return f.getResult, f.getErr
}
func (f *fakeRepository) Create(_ context.Context, input taskrepository.CreateInput) (*entity.Task, error) {
	f.capturedCreate = input
	return f.createResult, f.createErr
}
func (f *fakeRepository) Update(_ context.Context, input taskrepository.UpdateInput) (*entity.Task, error) {
	f.capturedUpdate = input
	return f.updateResult, f.updateErr
}
func (f *fakeRepository) SoftDelete(_ context.Context, cmd taskrepository.DeleteCommand) error {
	f.capturedDelete = cmd
	return f.deleteErr
}
func (f *fakeRepository) ListByTeam(_ context.Context, query taskrepository.ByTeamQuery) ([]entity.Task, int64, error) {
	f.capturedByTeam = query
	return f.byTeamTasks, f.byTeamTotal, f.byTeamErr
}
func (f *fakeRepository) ListByFilter(_ context.Context, query taskrepository.ByFilterQuery) ([]entity.Task, error) {
	f.capturedByFilter = query
	return f.byFilterTasks, f.byFilterErr
}

func TestListPaginationNormalization(t *testing.T) {
	cases := []struct {
		name       string
		inputPage  int
		inputLimit int
		wantPage   int
		wantLimit  int
	}{
		{"page 0", 0, 10, 1, 10},
		{"page negative", -5, 10, 1, 10},
		{"limit 0", 1, 0, 1, 50},
		{"limit negative", 1, -1, 1, 50},
		{"limit over max", 1, 101, 1, 50},
		{"limit max", 1, 100, 1, 100},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepository{}
			svc := NewService(repo)

			out, err := svc.List(context.Background(), ListTasksInput{Page: tc.inputPage, Limit: tc.inputLimit})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.Page != tc.wantPage || out.Limit != tc.wantLimit {
				t.Fatalf("output page/limit = (%d,%d), want (%d,%d)", out.Page, out.Limit, tc.wantPage, tc.wantLimit)
			}
			if repo.capturedList.Page != tc.wantPage || repo.capturedList.Limit != tc.wantLimit {
				t.Fatalf("repo page/limit = (%d,%d), want (%d,%d)", repo.capturedList.Page, repo.capturedList.Limit, tc.wantPage, tc.wantLimit)
			}
		})
	}
}

func TestListPassesFiltersAndTotal(t *testing.T) {
	repo := &fakeRepository{listTotal: 99}
	svc := NewService(repo)
	teamID := int64(42)
	jobTypeID := int64(7)
	feederID := int64(3)
	workDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	out, err := svc.List(context.Background(), ListTasksInput{
		Page:  1,
		Limit: 10,
		Filter: taskrepository.ListFilter{
			TeamID:    &teamID,
			JobTypeID: &jobTypeID,
			FeederID:  &feederID,
			WorkDate:  &workDate,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Total != 99 {
		t.Fatalf("Total = %d, want 99", out.Total)
	}
	if repo.capturedList.Filter.TeamID == nil || *repo.capturedList.Filter.TeamID != 42 {
		t.Fatalf("TeamID = %#v, want 42", repo.capturedList.Filter.TeamID)
	}
	if repo.capturedList.Filter.WorkDate == nil || !repo.capturedList.Filter.WorkDate.Equal(workDate) {
		t.Fatalf("WorkDate = %#v, want %v", repo.capturedList.Filter.WorkDate, workDate)
	}
}

func TestListPropagatesRepositoryError(t *testing.T) {
	repo := &fakeRepository{listErr: errors.New("db failure")}
	svc := NewService(repo)

	out, err := svc.List(context.Background(), ListTasksInput{Page: 1, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	if out != nil {
		t.Fatal("expected nil output")
	}
}

func TestGetByIDValidationAndNotFound(t *testing.T) {
	svc := NewService(&fakeRepository{})
	if _, err := svc.GetByID(context.Background(), 0); !errors.Is(err, ErrInvalidTaskID) {
		t.Fatalf("got %v, want ErrInvalidTaskID", err)
	}
	if _, err := svc.GetByID(context.Background(), 1); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("got %v, want ErrTaskNotFound", err)
	}
}

func TestCreateValidation(t *testing.T) {
	svc := NewService(&fakeRepository{})
	if _, err := svc.Create(context.Background(), CreateTaskInput{TeamID: 1, JobTypeID: 1, JobDetailID: 1}); !errors.Is(err, ErrWorkDateRequired) {
		t.Fatalf("got %v, want ErrWorkDateRequired", err)
	}
	if _, err := svc.Create(context.Background(), CreateTaskInput{WorkDate: time.Now(), JobTypeID: 1, JobDetailID: 1}); !errors.Is(err, ErrTeamIDRequired) {
		t.Fatalf("got %v, want ErrTeamIDRequired", err)
	}
	if _, err := svc.Create(context.Background(), CreateTaskInput{WorkDate: time.Now(), TeamID: 1, JobDetailID: 1}); !errors.Is(err, ErrJobTypeIDRequired) {
		t.Fatalf("got %v, want ErrJobTypeIDRequired", err)
	}
	if _, err := svc.Create(context.Background(), CreateTaskInput{WorkDate: time.Now(), TeamID: 1, JobTypeID: 1}); !errors.Is(err, ErrJobDetailIDRequired) {
		t.Fatalf("got %v, want ErrJobDetailIDRequired", err)
	}
}

func TestUpdateAndDeleteValidation(t *testing.T) {
	svc := NewService(&fakeRepository{})
	if _, err := svc.Update(context.Background(), UpdateTaskInput{ID: 0}); !errors.Is(err, ErrInvalidTaskID) {
		t.Fatalf("got %v, want ErrInvalidTaskID", err)
	}
	if err := svc.Delete(context.Background(), 0); !errors.Is(err, ErrInvalidTaskID) {
		t.Fatalf("got %v, want ErrInvalidTaskID", err)
	}
}

func TestListByTeamNormalization(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)

	out, err := svc.ListByTeam(context.Background(), ListTasksByTeamInput{Page: 0, Limit: 501})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Page != 1 || out.Limit != 100 {
		t.Fatalf("output page/limit = (%d,%d), want (1,100)", out.Page, out.Limit)
	}
	if repo.capturedByTeam.Page != 1 || repo.capturedByTeam.Limit != 100 {
		t.Fatalf("repo page/limit = (%d,%d), want (1,100)", repo.capturedByTeam.Page, repo.capturedByTeam.Limit)
	}
}

func TestListByFilterRequiresYearMonth(t *testing.T) {
	svc := NewService(&fakeRepository{})

	if _, err := svc.ListByFilter(context.Background(), ListTasksByFilterInput{Year: "", Month: "1"}); !errors.Is(err, ErrYearMonthRequired) {
		t.Fatalf("got %v, want ErrYearMonthRequired", err)
	}
	if _, err := svc.ListByFilter(context.Background(), ListTasksByFilterInput{Year: "2024", Month: ""}); !errors.Is(err, ErrYearMonthRequired) {
		t.Fatalf("got %v, want ErrYearMonthRequired", err)
	}
}
