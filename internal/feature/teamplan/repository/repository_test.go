package repository

import (
	"testing"
	"time"
)

func TestExpandDateRangeInclusive(t *testing.T) {
	start := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	got := expandDateRange(start, &end)
	if len(got) != 3 {
		t.Fatalf("expandDateRange len = %d, want 3", len(got))
	}
	if !got[0].Equal(start) || !got[2].Equal(end) {
		t.Fatalf("expandDateRange = %v", got)
	}
}
