package dto

import (
	"encoding/json"
	"testing"
)

func TestReplaceTasksRequestAcceptsLiteLocationPayload(t *testing.T) {
	raw := []byte(`{"tasks":[{"assignedTeamId":8,"latitude":13.756331,"longitude":100.501762,"workDetail":"รายละเอียดหน้างาน","beforePhotoUrls":[" https://cdn.example/before-site.jpg ",""]}]}`)

	var req ReplaceTasksRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal lite replace tasks request: %v", err)
	}

	if len(req.Tasks) != 1 {
		t.Fatalf("tasks len = %d, want 1", len(req.Tasks))
	}
	task := req.Tasks[0]
	if task.AssignedTeamID != 8 {
		t.Fatalf("assigned team id = %d, want 8", task.AssignedTeamID)
	}
	if task.PointLabel != "" || task.WorkType != "" {
		t.Fatalf("lite task should not require pointLabel/workType, got pointLabel=%q workType=%q", task.PointLabel, task.WorkType)
	}
	if task.Latitude == nil || *task.Latitude != 13.756331 || task.Longitude == nil || *task.Longitude != 100.501762 {
		t.Fatalf("location not decoded: %#v", task)
	}
	if task.WorkDetail == nil || *task.WorkDetail != "รายละเอียดหน้างาน" {
		t.Fatalf("work detail not decoded: %#v", task.WorkDetail)
	}
	if len(task.BeforePhotoURLs) != 2 || task.BeforePhotoURLs[0] != " https://cdn.example/before-site.jpg " || task.BeforePhotoURLs[1] != "" {
		t.Fatalf("before photo urls not decoded for service cleanup: %#v", task.BeforePhotoURLs)
	}
}
