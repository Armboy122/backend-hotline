package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend-hotlines3/internal/feature/largework/entity"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

const (
	defaultLargeWorkJobTypeName   = "งานระดมทีม"
	defaultLargeWorkJobDetailName = "บันทึกอัตโนมัติจากงานระดมทีม"
)

type AutoDailyReportInput struct {
	Plan              entity.LargeWorkItem
	Task              entity.LargeWorkTask
	CompletedByUserID int64
	CompletedAt       time.Time
	CompletionNote    string
	AfterPhotoURLs    []string
}

type DailyReportCreator interface {
	CreateFromLargeWorkTask(context.Context, AutoDailyReportInput) error
}

type dailyReportCreator struct {
	db *gorm.DB
}

func NewDailyReportCreator(db *gorm.DB) DailyReportCreator {
	return &dailyReportCreator{db: db}
}

func (r *dailyReportCreator) CreateFromLargeWorkTask(ctx context.Context, input AutoDailyReportInput) error {
	if r == nil || r.db == nil || input.Task.ID <= 0 || input.Task.AssignedTeamID <= 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.TaskDaily
		err := tx.Where("large_work_task_id = ? AND deletedat IS NULL", input.Task.ID).First(&existing).Error
		if err == nil {
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		jobTypeID, jobDetailID, err := r.resolveJobRefs(tx, input.Task)
		if err != nil {
			return err
		}

		sourceType := "large_work"
		sourceID := input.Plan.ID
		largeWorkTaskID := input.Task.ID
		now := time.Now()
		workDate := dateOnly(input.Plan.StartDate)
		if workDate.IsZero() {
			workDate = dateOnly(input.CompletedAt)
		}

		report := models.TaskDaily{
			WorkDate:        workDate,
			JobTypeID:       jobTypeID,
			JobDetailID:     jobDetailID,
			FeederID:        input.Plan.FeederID,
			URLsBefore:      models.StringArray(input.Task.BeforePhotoURLs),
			URLsAfter:       models.StringArray(input.AfterPhotoURLs),
			CreatedAt:       now,
			UpdatedAt:       now,
			Detail:          stringPtr(buildLargeWorkDailyReportDetail(input)),
			TeamID:          input.Task.AssignedTeamID,
			Latitude:        floatPtrToDecimal(input.Task.Latitude),
			Longitude:       floatPtrToDecimal(input.Task.Longitude),
			SourceType:      &sourceType,
			SourceID:        &sourceID,
			LargeWorkTaskID: &largeWorkTaskID,
		}
		if err := tx.Create(&report).Error; err != nil {
			if isDuplicateDailyReportError(err) {
				return nil
			}
			return err
		}
		return nil
	})
}

func (r *dailyReportCreator) resolveJobRefs(tx *gorm.DB, task entity.LargeWorkTask) (int64, int64, error) {
	jobTypeID := metadataInt64(task.Metadata, "jobTypeId")
	jobDetailID := metadataInt64(task.Metadata, "jobDetailId")
	if jobTypeID > 0 && jobDetailID > 0 {
		return jobTypeID, jobDetailID, nil
	}

	jobType := models.JobType{Name: defaultLargeWorkJobTypeName}
	if err := tx.Where("name = ?", defaultLargeWorkJobTypeName).FirstOrCreate(&jobType).Error; err != nil {
		return 0, 0, err
	}

	var jobDetail models.JobDetail
	if err := tx.Where("name = ?", defaultLargeWorkJobDetailName).First(&jobDetail).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, err
		}
		now := time.Now()
		jobDetail = models.JobDetail{
			Name:      defaultLargeWorkJobDetailName,
			JobTypeID: &jobType.ID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Create(&jobDetail).Error; err != nil {
			return 0, 0, err
		}
	}

	return jobType.ID, jobDetail.ID, nil
}

func buildLargeWorkDailyReportDetail(input AutoDailyReportInput) string {
	parts := []string{}
	if title := strings.TrimSpace(input.Plan.Title); title != "" {
		parts = append(parts, "งานระดมทีม: "+title)
	}
	if point := strings.TrimSpace(input.Task.PointLabel); point != "" {
		parts = append(parts, "จุดงาน: "+point)
	}
	if workType := strings.TrimSpace(input.Task.WorkType); workType != "" {
		parts = append(parts, "ประเภทงาน: "+workType)
	}
	if input.Task.WorkDetail != nil && strings.TrimSpace(*input.Task.WorkDetail) != "" {
		parts = append(parts, "รายละเอียด: "+strings.TrimSpace(*input.Task.WorkDetail))
	}
	if note := strings.TrimSpace(input.CompletionNote); note != "" {
		parts = append(parts, "ผลการทำงาน: "+note)
	}
	if len(parts) == 0 {
		return fmt.Sprintf("บันทึกอัตโนมัติจากงานระดมทีม #%d", input.Plan.ID)
	}
	return strings.Join(parts, "\n")
}

func metadataInt64(metadata map[string]any, key string) int64 {
	if metadata == nil {
		return 0
	}
	switch value := metadata[key].(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case json.Number:
		out, _ := value.Int64()
		return out
	case string:
		out, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		return out
	default:
		return 0
	}
}

func isDuplicateDailyReportError(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique") || strings.Contains(lower, "23505")
}

func dateOnly(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func stringPtr(value string) *string {
	return &value
}
