package dto

import (
	"encoding/json"
	"testing"
)

func TestMonthlyPlanSettingsRoundOneJSONContract(t *testing.T) {
	resp := MonthlyPlanSettingsResponse{
		LockDay:                 23,
		AdminCanUploadAfterLock: true,
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	allowed := map[string]bool{
		"lockDay":                 true,
		"adminCanUploadAfterLock": true,
	}
	for key := range got {
		if !allowed[key] {
			t.Fatalf("unexpected round-1 settings key %q in JSON %s", key, payload)
		}
	}
	for key := range allowed {
		if _, ok := got[key]; !ok {
			t.Fatalf("missing round-1 settings key %q in JSON %s", key, payload)
		}
	}
}

func TestUpdateMonthlyPlanSettingsRoundOneRequestJSONContract(t *testing.T) {
	lockDay := 23
	afterLock := true
	req := UpdateMonthlyPlanSettingsRequest{
		LockDay:                 &lockDay,
		AdminCanUploadAfterLock: &afterLock,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	allowed := map[string]bool{
		"lockDay":                 true,
		"adminCanUploadAfterLock": true,
	}
	for key := range got {
		if !allowed[key] {
			t.Fatalf("unexpected round-1 settings request key %q in JSON %s", key, payload)
		}
	}
}
