package repository

import (
	"context"
	"reflect"
	"testing"
	"time"

	largeworkentity "backend-hotlines3/internal/feature/largework/entity"
	"backend-hotlines3/internal/models"

	"github.com/shopspring/decimal"
)

func TestLargeWorkTaskRepositoryContractIsExposed(t *testing.T) {
	var _ Repository = (*repository)(nil)
	var repo Repository
	_ = repo
	var taskRepo interface {
		ReplaceTasks(context.Context, ReplaceTasksInput) ([]largeworkentity.LargeWorkTask, error)
		ReplaceTasksAndParticipants(context.Context, ReplaceTasksAndParticipantsInput) ([]largeworkentity.LargeWorkTask, error)
		ListTasksByPlan(context.Context, ListTasksQuery) ([]largeworkentity.LargeWorkTask, error)
		ListAssignedTasks(context.Context, ListAssignedTasksQuery) ([]largeworkentity.LargeWorkTask, int64, error)
		GetTaskByID(context.Context, GetTaskQuery) (*largeworkentity.LargeWorkTask, error)
		UpdateTask(context.Context, UpdateTaskInput) (*largeworkentity.LargeWorkTask, error)
	} = repo
	_ = taskRepo

	input := ReplaceTasksInput{
		LargeWorkItemID: 501,
		Tasks: []TaskInput{{
			AssignedTeamID:  12,
			Sequence:        1,
			PointLabel:      "P1",
			WorkType:        "tree_trimming",
			BeforePhotoURLs: []string{"https://cdn.example/before.jpg"},
			AfterPhotoURLs:  []string{"https://cdn.example/after.jpg"},
			Metadata:        map[string]any{"source": "test"},
		}},
	}
	if input.LargeWorkItemID != 501 || len(input.Tasks) != 1 {
		t.Fatalf("replace task input not usable: %#v", input)
	}

	listByPlan := ListTasksQuery{LargeWorkItemID: 501}
	if listByPlan.LargeWorkItemID != 501 {
		t.Fatalf("list by plan query = %#v", listByPlan)
	}
	myTodos := ListAssignedTasksQuery{AssignedTeamID: 12, Statuses: []string{largeworkentity.LargeWorkTaskStatusTodo}}
	if myTodos.AssignedTeamID != 12 || !reflect.DeepEqual(myTodos.Statuses, []string{"todo"}) {
		t.Fatalf("assigned task query = %#v", myTodos)
	}
}

func TestMapTaskPreservesExecutionPointFields(t *testing.T) {
	startedAt := time.Date(2026, 5, 11, 8, 30, 0, 0, time.UTC)
	completedAt := startedAt.Add(time.Hour)
	startedBy := int64(77)
	completedBy := int64(78)
	pointCount := 1
	treeCount := 12
	itemCount := 3
	workDetail := "trim trees"
	notes := "bring cones"
	completionNote := "done"
	model := models.LargeWorkTask{
		ID:                9001,
		LargeWorkItemID:   501,
		AssignedTeamID:    12,
		Sequence:          2,
		PointLabel:        "P2",
		Latitude:          decimalPtr(decimal.NewFromFloat(13.756331)),
		Longitude:         decimalPtr(decimal.NewFromFloat(100.501762)),
		WorkType:          "tree_trimming",
		WorkDetail:        &workDetail,
		PointCount:        &pointCount,
		TreeCount:         &treeCount,
		ItemCount:         &itemCount,
		Notes:             &notes,
		Status:            largeworkentity.LargeWorkTaskStatusDone,
		BeforePhotoURLs:   models.StringArray{"https://cdn.example/before.jpg"},
		AfterPhotoURLs:    models.StringArray{"https://cdn.example/after.jpg"},
		CompletionNote:    &completionNote,
		StartedByUserID:   &startedBy,
		StartedAt:         &startedAt,
		CompletedByUserID: &completedBy,
		CompletedAt:       &completedAt,
		Metadata:          []byte(`{"source":"test"}`),
	}

	got := mapTask(model)
	if got.ID != 9001 || got.LargeWorkItemID != 501 || got.AssignedTeamID != 12 || got.Sequence != 2 {
		t.Fatalf("mapped identity/order fields = %#v", got)
	}
	if got.Latitude == nil || *got.Latitude != 13.756331 {
		t.Fatalf("mapped latitude = %#v", got.Latitude)
	}
	if got.Longitude == nil || *got.Longitude != 100.501762 {
		t.Fatalf("mapped longitude = %#v", got.Longitude)
	}
	if !reflect.DeepEqual(got.BeforePhotoURLs, []string{"https://cdn.example/before.jpg"}) {
		t.Fatalf("mapped before photos = %#v", got.BeforePhotoURLs)
	}
	if !reflect.DeepEqual(got.AfterPhotoURLs, []string{"https://cdn.example/after.jpg"}) {
		t.Fatalf("mapped after photos = %#v", got.AfterPhotoURLs)
	}
	if got.Metadata["source"] != "test" {
		t.Fatalf("mapped metadata = %#v", got.Metadata)
	}
}

func TestTaskInputToModelDefaultsMetadataObject(t *testing.T) {
	row, err := taskInputToModel(501, TaskInput{AssignedTeamID: 12, Sequence: 1, PointLabel: "P1", WorkType: "tree", Status: largeworkentity.LargeWorkTaskStatusTodo}, time.Date(2026, 5, 11, 8, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("taskInputToModel error: %v", err)
	}
	if string(row.Metadata) != "{}" {
		t.Fatalf("metadata = %q, want {}", string(row.Metadata))
	}
}

func decimalPtr(v decimal.Decimal) *decimal.Decimal { return &v }
