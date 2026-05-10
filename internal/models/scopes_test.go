package models

import (
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTaskByYearUsesIndexableWorkdateRange(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run db: %v", err)
	}

	stmt := db.Session(&gorm.Session{DryRun: true}).
		Model(&TaskDaily{}).
		Scopes(TaskByYear("2026")).
		Find(&[]TaskDaily{}).Statement
	sql := stmt.SQL.String()

	if strings.Contains(sql, "EXTRACT(YEAR") {
		t.Fatalf("TaskByYear SQL must not wrap workdate in EXTRACT because it bypasses workdate indexes: %s", sql)
	}
	if !strings.Contains(sql, `"workdate" >=`) || !strings.Contains(sql, `"workdate" <`) {
		t.Fatalf("TaskByYear SQL must use a half-open workdate range, got: %s", sql)
	}
}
