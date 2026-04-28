// Package usecase_test locks the observable behavior of ListTasksUseCase so
// that future pagination or filter changes are visible as test failures.
// Tests here do NOT connect to a database; they use a fakeTaskRepo.
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

// fakeTaskRepo captures the query sent to it so tests can assert normalized values.
type fakeTaskRepo struct {
	capturedQuery repository.TaskListQuery
	returnTasks   []taskdomain.Entity
	returnTotal   int64
	returnErr     error
}

func (f *fakeTaskRepo) List(_ context.Context, q repository.TaskListQuery) ([]taskdomain.Entity, int64, error) {
	f.capturedQuery = q
	return f.returnTasks, f.returnTotal, f.returnErr
}

// ── A4.1 Pagination normalization ─────────────────────────────────────────────

func TestListTasks_PaginationNormalization(t *testing.T) {
	cases := []struct {
		name       string
		inputPage  int
		inputLimit int
		wantPage   int
		wantLimit  int
	}{
		{"page 0 → 1", 0, 10, 1, 10},
		{"page -5 → 1", -5, 10, 1, 10},
		{"page 3 passes through", 3, 10, 3, 10},
		{"limit 0 → 50", 1, 0, 1, 50},
		{"limit -1 → 50", 1, -1, 1, 50},
		{"limit 101 → 50", 1, 101, 1, 50},
		{"limit 100 passes through", 1, 100, 1, 100},
		{"limit 1 passes through", 1, 1, 1, 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeTaskRepo{}
			uc := taskusecase.NewListTasksUseCase(repo)

			out, err := uc.Execute(context.Background(), taskusecase.ListTasksInput{
				Page:  tc.inputPage,
				Limit: tc.inputLimit,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if out.Page != tc.wantPage {
				t.Errorf("out.Page: got %d, want %d", out.Page, tc.wantPage)
			}
			if out.Limit != tc.wantLimit {
				t.Errorf("out.Limit: got %d, want %d", out.Limit, tc.wantLimit)
			}
			// Repo must receive the already-normalized values.
			if repo.capturedQuery.Page != tc.wantPage {
				t.Errorf("repo.Page: got %d, want %d", repo.capturedQuery.Page, tc.wantPage)
			}
			if repo.capturedQuery.Limit != tc.wantLimit {
				t.Errorf("repo.Limit: got %d, want %d", repo.capturedQuery.Limit, tc.wantLimit)
			}
		})
	}
}

// TestListTasks_FiltersPassedThroughToRepo verifies that the usecase does not
// discard or modify the caller-supplied filter before forwarding it to the repo.
func TestListTasks_FiltersPassedThroughToRepo(t *testing.T) {
	repo := &fakeTaskRepo{}
	uc := taskusecase.NewListTasksUseCase(repo)

	teamID := int64(42)
	jobTypeID := int64(7)
	feederID := int64(3)
	workDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	_, _ = uc.Execute(context.Background(), taskusecase.ListTasksInput{
		Page:  1,
		Limit: 10,
		Filter: repository.TaskListFilter{
			TeamID:    &teamID,
			JobTypeID: &jobTypeID,
			FeederID:  &feederID,
			WorkDate:  &workDate,
		},
	})

	q := repo.capturedQuery
	if q.Filter.TeamID == nil || *q.Filter.TeamID != 42 {
		t.Errorf("TeamID: got %v, want 42", q.Filter.TeamID)
	}
	if q.Filter.JobTypeID == nil || *q.Filter.JobTypeID != 7 {
		t.Errorf("JobTypeID: got %v, want 7", q.Filter.JobTypeID)
	}
	if q.Filter.FeederID == nil || *q.Filter.FeederID != 3 {
		t.Errorf("FeederID: got %v, want 3", q.Filter.FeederID)
	}
	if q.Filter.WorkDate == nil || !q.Filter.WorkDate.Equal(workDate) {
		t.Errorf("WorkDate: got %v, want %v", q.Filter.WorkDate, workDate)
	}
}

// TestListTasks_TotalFromRepoReturnedInOutput locks that the repo's total count
// is surfaced unchanged in the output struct.
func TestListTasks_TotalFromRepoReturnedInOutput(t *testing.T) {
	repo := &fakeTaskRepo{returnTotal: 99}
	uc := taskusecase.NewListTasksUseCase(repo)

	out, _ := uc.Execute(context.Background(), taskusecase.ListTasksInput{Page: 1, Limit: 10})
	if out.Total != 99 {
		t.Errorf("got Total=%d, want 99", out.Total)
	}
}

// TestListTasks_RepoErrorPropagates verifies that a repo error is returned to
// the caller unchanged and no output is produced.
func TestListTasks_RepoErrorPropagates(t *testing.T) {
	repo := &fakeTaskRepo{returnErr: errors.New("db failure")}
	uc := taskusecase.NewListTasksUseCase(repo)

	out, err := uc.Execute(context.Background(), taskusecase.ListTasksInput{Page: 1, Limit: 10})
	if err == nil {
		t.Error("expected error from repo, got nil")
	}
	if out != nil {
		t.Error("expected nil output on error, got non-nil")
	}
}
