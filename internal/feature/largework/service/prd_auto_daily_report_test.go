package service

import (
	"context"
	"errors"
	"testing"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/largework/entity"
)

func TestCompleteTaskSurfacesAutoDailyReportFailure(t *testing.T) {
	teamID := int64(8)
	workDetail := "ตัดกิ่งไม้ใกล้แนวสาย"
	plan := &entity.LargeWorkItem{ID: 501, OwnerTeamID: 7, StartDate: mustDate(t, "2026-06-10"), Status: entity.LargeWorkStatusInProgress, Teams: []entity.LargeWorkTeam{{ID: teamID}}}
	task := &entity.LargeWorkTask{
		ID:              11,
		LargeWorkItemID: 501,
		AssignedTeamID:  teamID,
		Status:          entity.LargeWorkTaskStatusInProgress,
		PointLabel:      "P-001",
		WorkType:        "tree_trim",
		WorkDetail:      &workDetail,
		BeforePhotoURLs: []string{"https://cdn.example/before.jpg"},
	}
	updated := *task
	updated.Status = entity.LargeWorkTaskStatusDone
	dailyErr := errors.New("daily report create failed")
	repo := &fakeRepository{getItem: plan, getTask: task, updateTask: &updated}
	dailyReports := &fakeDailyReportCreator{err: dailyErr}
	svc := NewServiceWithDailyReport(repo, dailyReports)

	done, err := svc.CompleteTask(context.Background(), actor(policy.RoleUser, &teamID), 11, CompleteTaskInput{
		AfterPhotoURLs: []string{"https://cdn.example/after.jpg"},
		CompletionNote: "done",
	})

	if !errors.Is(err, dailyErr) {
		t.Fatalf("CompleteTask error = %v, want daily report error", err)
	}
	if done != nil {
		t.Fatalf("CompleteTask returned done task despite auto-report failure: %#v", done)
	}
	if dailyReports.calls != 1 {
		t.Fatalf("daily report calls = %d, want 1", dailyReports.calls)
	}
}

func TestCompleteTaskDoesNotCreateDuplicateDailyReportForTerminalTask(t *testing.T) {
	teamID := int64(8)
	task := &entity.LargeWorkTask{ID: 11, LargeWorkItemID: 501, AssignedTeamID: teamID, Status: entity.LargeWorkTaskStatusDone}
	plan := &entity.LargeWorkItem{ID: 501, OwnerTeamID: 7, Status: entity.LargeWorkStatusInProgress, Teams: []entity.LargeWorkTeam{{ID: teamID}}}
	repo := &fakeRepository{getItem: plan, getTask: task, updateTask: task}
	dailyReports := &fakeDailyReportCreator{}
	svc := NewServiceWithDailyReport(repo, dailyReports)

	_, err := svc.CompleteTask(context.Background(), actor(policy.RoleUser, &teamID), 11, CompleteTaskInput{
		AfterPhotoURLs: []string{"https://cdn.example/after.jpg"},
		CompletionNote: "done again",
	})

	if !errors.Is(err, ErrInvalidStateTransition) {
		t.Fatalf("CompleteTask terminal task got %v, want ErrInvalidStateTransition", err)
	}
	if dailyReports.calls != 0 {
		t.Fatalf("daily report calls = %d, want 0 for already-done task", dailyReports.calls)
	}
	if repo.capturedUpdateTask.ID != 0 {
		t.Fatalf("repository update captured despite terminal task: %#v", repo.capturedUpdateTask)
	}
}
