package repository

import (
	"strings"
	"testing"
	"time"

	"backend-hotlines3/internal/feature/largework/entity"
)

func TestBuildLargeWorkDailyReportDetailMapsPRDFields(t *testing.T) {
	workDetail := "ตัดกิ่งไม้ใกล้แนวสาย"
	detail := buildLargeWorkDailyReportDetail(AutoDailyReportInput{
		Plan:           entity.LargeWorkItem{ID: 501, Title: "งานระดมทีม เปลี่ยนอุปกรณ์หลัก"},
		Task:           entity.LargeWorkTask{ID: 11, PointLabel: "P-001", WorkType: "tree_trim", WorkDetail: &workDetail},
		CompletionNote: "ดำเนินการเสร็จและคืนสภาพพื้นที่แล้ว",
	})

	for _, fragment := range []string{
		"งานระดมทีม: งานระดมทีม เปลี่ยนอุปกรณ์หลัก",
		"จุดงาน: P-001",
		"ประเภทงาน: tree_trim",
		"รายละเอียด: ตัดกิ่งไม้ใกล้แนวสาย",
		"ผลการทำงาน: ดำเนินการเสร็จและคืนสภาพพื้นที่แล้ว",
	} {
		if !strings.Contains(detail, fragment) {
			t.Fatalf("daily report detail missing %q\n%s", fragment, detail)
		}
	}
}

func TestBuildLargeWorkDailyReportDetailFallsBackToPlanID(t *testing.T) {
	detail := buildLargeWorkDailyReportDetail(AutoDailyReportInput{Plan: entity.LargeWorkItem{ID: 501}})
	if detail != "บันทึกอัตโนมัติจากงานระดมทีม #501" {
		t.Fatalf("fallback detail = %q", detail)
	}
}

func TestLargeWorkDailyReportIdempotencyHelpers(t *testing.T) {
	for _, raw := range []string{
		`ERROR: duplicate key value violates unique constraint "idx_task_dailies_large_work_task_id"`,
		`pq: duplicate key value violates unique constraint (SQLSTATE 23505)`,
		`UNIQUE constraint failed: task_dailies.large_work_task_id`,
	} {
		if !isDuplicateDailyReportError(assertError(raw)) {
			t.Fatalf("duplicate helper did not recognize %q", raw)
		}
	}
	if isDuplicateDailyReportError(assertError("connection refused")) {
		t.Fatalf("duplicate helper must not hide non-duplicate database errors")
	}

	got := dateOnly(time.Date(2026, 5, 19, 14, 30, 0, 0, time.FixedZone("ICT", 7*60*60)))
	want := time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("dateOnly = %s, want %s", got, want)
	}
}

type stringError string

func (e stringError) Error() string { return string(e) }

func assertError(raw string) error { return stringError(raw) }
