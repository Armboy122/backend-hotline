package entity

import (
	"reflect"
	"testing"
	"time"
)

func TestLargeWorkTaskEntitySupportsPointExecutionContract(t *testing.T) {
	startedAt := time.Date(2026, 5, 11, 9, 30, 0, 0, time.UTC)
	completedAt := startedAt.Add(45 * time.Minute)
	startedBy := int64(77)
	completedBy := int64(78)
	lat := 13.756331
	lng := 100.501762
	pointCount := 1
	treeCount := 12
	itemCount := 3
	metadata := map[string]any{"priority": "high", "round": float64(1)}

	task := LargeWorkTask{
		ID:                9001,
		LargeWorkItemID:   501,
		AssignedTeamID:    12,
		Sequence:          2,
		PointLabel:        "จุดที่ 2 ซอยสุขุมวิท 1",
		Latitude:          &lat,
		Longitude:         &lng,
		WorkType:          "tree_trimming",
		WorkDetail:        strPtr("trim trees near feeder"),
		PointCount:        &pointCount,
		TreeCount:         &treeCount,
		ItemCount:         &itemCount,
		Notes:             strPtr("need traffic cones"),
		Status:            LargeWorkTaskStatusDone,
		BeforePhotoURLs:   []string{"https://cdn.example/before-1.jpg"},
		AfterPhotoURLs:    []string{"https://cdn.example/after-1.jpg"},
		CompletionNote:    strPtr("completed safely"),
		StartedByUserID:   &startedBy,
		StartedAt:         &startedAt,
		CompletedByUserID: &completedBy,
		CompletedAt:       &completedAt,
		Metadata:          metadata,
	}

	if task.LargeWorkItemID != 501 || task.AssignedTeamID != 12 {
		t.Fatalf("task plan/team = %d/%d, want 501/12", task.LargeWorkItemID, task.AssignedTeamID)
	}
	if !reflect.DeepEqual(task.BeforePhotoURLs, []string{"https://cdn.example/before-1.jpg"}) {
		t.Fatalf("before photo URLs = %#v", task.BeforePhotoURLs)
	}
	if got := task.Metadata["priority"]; got != "high" {
		t.Fatalf("metadata priority = %#v, want high", got)
	}
}

func TestLargeWorkTaskStatusesStayWithinExecutionQueueContract(t *testing.T) {
	statuses := []string{
		LargeWorkTaskStatusTodo,
		LargeWorkTaskStatusInProgress,
		LargeWorkTaskStatusDone,
		LargeWorkTaskStatusBlocked,
		LargeWorkTaskStatusCancelled,
	}
	want := []string{"todo", "in_progress", "done", "blocked", "cancelled"}
	if !reflect.DeepEqual(statuses, want) {
		t.Fatalf("large work task statuses = %#v, want %#v", statuses, want)
	}
}

func strPtr(v string) *string { return &v }
