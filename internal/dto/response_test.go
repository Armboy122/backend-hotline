package dto

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStandardResponseOmitsEmptyEnvelopeFields(t *testing.T) {
	payload := StandardResponse{
		Success: true,
		Data: map[string]string{
			"hello": "world",
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	jsonText := string(b)
	if strings.Contains(jsonText, "meta") {
		t.Fatalf("json unexpectedly contains meta: %s", jsonText)
	}
	if strings.Contains(jsonText, "error") {
		t.Fatalf("json unexpectedly contains error: %s", jsonText)
	}
}

func TestTaskResponseEncodesNestedTaskGraph(t *testing.T) {
	payload := TaskResponse{
		ID:          99,
		WorkDate:    "2026-04-28",
		TeamID:      1,
		JobTypeID:   2,
		JobDetailID: 3,
		FeederID:    int64Ptr(7),
		NumPole:     stringPtr("12"),
		DeviceCode:  stringPtr("DEV-1"),
		Detail:      stringPtr("replace fuse"),
		URLsBefore:  []string{"before-1"},
		URLsAfter:   []string{"after-1"},
		Team:        &TeamNested{ID: 1, Name: "Team Alpha"},
		JobType:     &JobTypeNested{ID: 2, Name: "Inspection"},
		JobDetail:   &JobDetailNested{ID: 3, Name: "Fuse Check"},
		Feeder: &FeederNestedForTask{
			ID:   7,
			Code: "F-007",
			Station: &StationNestedSimple{
				Name: "Station North",
				OperationCenter: &OperationCenterNested{
					ID:   55,
					Name: "Center A",
				},
			},
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["workDate"] != "2026-04-28" {
		t.Fatalf("workDate = %#v, want 2026-04-28", decoded["workDate"])
	}
	team := decoded["team"].(map[string]any)
	if team["name"] != "Team Alpha" {
		t.Fatalf("team.name = %#v, want Team Alpha", team["name"])
	}
	feeder := decoded["feeder"].(map[string]any)
	station := feeder["station"].(map[string]any)
	operationCenter := station["operationCenter"].(map[string]any)
	if operationCenter["name"] != "Center A" {
		t.Fatalf("operationCenter.name = %#v, want Center A", operationCenter["name"])
	}
}

func int64Ptr(v int64) *int64    { return &v }
func stringPtr(v string) *string { return &v }
