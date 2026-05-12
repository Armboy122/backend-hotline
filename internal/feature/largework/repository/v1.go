package repository

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"backend-hotlines3/internal/feature/largework/entity"
	"backend-hotlines3/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var ErrSchemaUnavailable = errors.New("large work task schema is unavailable; run database migration for large_work_tasks")

type ListQuery struct {
	Page     int
	Limit    int
	From     *time.Time
	To       *time.Time
	TeamID   *int64
	Statuses []string
}

type GetQuery struct {
	ID int64
}

type CreateInput struct {
	OwnerTeamID        int64
	CreatedByUserID    int64
	ParticipantTeamIDs []int64
	Title              string
	WorkType           *string
	StartDate          time.Time
	EndDate            *time.Time
	WorkTime           *string
	LocationText       string
	PEAID              *int64
	OperationCenterID  *int64
	FeederID           *int64
	StationID          *int64
	Notes              *string
	Status             string
}

type UpdateInput struct {
	ID                 int64
	OwnerTeamID        *int64
	ParticipantTeamIDs []int64
	Title              *string
	WorkType           *string
	StartDate          *time.Time
	EndDate            *time.Time
	WorkTime           *string
	LocationText       *string
	PEAID              *int64
	OperationCenterID  *int64
	FeederID           *int64
	StationID          *int64
	Notes              *string
	Status             *string
}

type TaskInput struct {
	AssignedTeamID    int64
	Sequence          int
	PointLabel        string
	Latitude          *float64
	Longitude         *float64
	WorkType          string
	WorkDetail        *string
	PointCount        *int
	TreeCount         *int
	ItemCount         *int
	Notes             *string
	Status            string
	BeforePhotoURLs   []string
	AfterPhotoURLs    []string
	CompletionNote    *string
	StartedByUserID   *int64
	StartedAt         *time.Time
	CompletedByUserID *int64
	CompletedAt       *time.Time
	Metadata          map[string]any
}

type ReplaceTasksInput struct {
	LargeWorkItemID int64
	Tasks           []TaskInput
}

type ReplaceTasksAndParticipantsInput struct {
	LargeWorkItemID    int64
	OwnerTeamID        int64
	ParticipantTeamIDs []int64
	Tasks              []TaskInput
}

type ListTasksQuery struct {
	LargeWorkItemID int64
	Statuses        []string
}

type ListAssignedTasksQuery struct {
	AssignedTeamID int64
	Statuses       []string
	Page           int
	Limit          int
}

type GetTaskQuery struct {
	ID int64
}

type UpdateTaskInput struct {
	ID                int64
	Status            *string
	BeforePhotoURLs   []string
	AfterPhotoURLs    []string
	CompletionNote    *string
	Notes             *string
	StartedByUserID   *int64
	StartedAt         *time.Time
	CompletedByUserID *int64
	CompletedAt       *time.Time
}

type Repository interface {
	List(context.Context, ListQuery) ([]entity.LargeWorkItem, int64, error)
	GetByID(context.Context, GetQuery) (*entity.LargeWorkItem, error)
	Create(context.Context, CreateInput) (*entity.LargeWorkItem, error)
	Update(context.Context, UpdateInput) (*entity.LargeWorkItem, error)
	ReplaceTasks(context.Context, ReplaceTasksInput) ([]entity.LargeWorkTask, error)
	ReplaceTasksAndParticipants(context.Context, ReplaceTasksAndParticipantsInput) ([]entity.LargeWorkTask, error)
	ListTasksByPlan(context.Context, ListTasksQuery) ([]entity.LargeWorkTask, error)
	ListAssignedTasks(context.Context, ListAssignedTasksQuery) ([]entity.LargeWorkTask, int64, error)
	GetTaskByID(context.Context, GetTaskQuery) (*entity.LargeWorkTask, error)
	UpdateTask(context.Context, UpdateTaskInput) (*entity.LargeWorkTask, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context, q ListQuery) ([]entity.LargeWorkItem, int64, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	limit := q.Limit
	if limit < 1 || limit > 100 {
		limit = 50
	}

	dbq := r.db.WithContext(ctx).Model(&models.LargeWorkItem{}).Where("deleted_at IS NULL")
	if q.From != nil {
		dbq = dbq.Where("COALESCE(end_date, start_date) >= ?", *q.From)
	}
	if q.To != nil {
		dbq = dbq.Where("start_date <= ?", *q.To)
	}
	if len(q.Statuses) > 0 {
		dbq = dbq.Where("status IN ?", q.Statuses)
	}
	if q.TeamID != nil {
		dbq = dbq.Where("(owner_team_id = ? OR EXISTS (SELECT 1 FROM large_work_item_teams lwit WHERE lwit.large_work_item_id = large_work_items.id AND lwit.team_id = ?))", *q.TeamID, *q.TeamID)
	}

	var total int64
	if err := dbq.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.LargeWorkItem
	if err := dbq.Preload("Teams.Team").
		Order("start_date DESC, created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return mapItems(items), total, nil
}

func (r *repository) GetByID(ctx context.Context, q GetQuery) (*entity.LargeWorkItem, error) {
	var item models.LargeWorkItem
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Preload("Teams.Team").First(&item, q.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	out := mapItem(item)
	return &out, nil
}

func (r *repository) Create(ctx context.Context, input CreateInput) (*entity.LargeWorkItem, error) {
	now := time.Now()
	item := models.LargeWorkItem{
		OwnerTeamID:       input.OwnerTeamID,
		CreatedByUserID:   input.CreatedByUserID,
		Title:             input.Title,
		WorkType:          input.WorkType,
		StartDate:         input.StartDate,
		EndDate:           input.EndDate,
		WorkTime:          input.WorkTime,
		LocationText:      input.LocationText,
		PEAID:             input.PEAID,
		OperationCenterID: input.OperationCenterID,
		FeederID:          input.FeederID,
		StationID:         input.StationID,
		Notes:             input.Notes,
		Status:            input.Status,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	participantIDs := uniqueTeamIDs(append(input.ParticipantTeamIDs, input.OwnerTeamID))
	returnVal := entity.LargeWorkItem{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		teams := make([]models.LargeWorkItemTeam, 0, len(participantIDs))
		for _, teamID := range participantIDs {
			role := entity.LargeWorkTeamRoleParticipant
			if teamID == input.OwnerTeamID {
				role = entity.LargeWorkTeamRoleOwner
			}
			teams = append(teams, models.LargeWorkItemTeam{
				LargeWorkItemID:   item.ID,
				TeamID:            teamID,
				Role:              role,
				ParticipantStatus: entity.LargeWorkParticipantStatusAssigned,
			})
		}
		if len(teams) > 0 {
			if err := tx.Create(&teams).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Preload("Teams.Team").First(&item, item.ID).Error; err != nil {
		return nil, err
	}
	returnVal = mapItem(item)
	return &returnVal, nil
}

func (r *repository) Update(ctx context.Context, input UpdateInput) (*entity.LargeWorkItem, error) {
	var item models.LargeWorkItem
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Preload("Teams.Team").First(&item, input.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{"updated_at": time.Now()}
		if input.OwnerTeamID != nil {
			updates["owner_team_id"] = *input.OwnerTeamID
		}
		if input.Title != nil {
			updates["title"] = *input.Title
		}
		if input.WorkType != nil {
			updates["work_type"] = *input.WorkType
		}
		if input.StartDate != nil {
			updates["start_date"] = *input.StartDate
		}
		if input.EndDate != nil {
			updates["end_date"] = *input.EndDate
		}
		if input.WorkTime != nil {
			updates["work_time"] = *input.WorkTime
		}
		if input.LocationText != nil {
			updates["location_text"] = *input.LocationText
		}
		if input.PEAID != nil {
			updates["pea_id"] = *input.PEAID
		}
		if input.OperationCenterID != nil {
			updates["operation_center_id"] = *input.OperationCenterID
		}
		if input.FeederID != nil {
			updates["feeder_id"] = *input.FeederID
		}
		if input.StationID != nil {
			updates["station_id"] = *input.StationID
		}
		if input.Notes != nil {
			updates["notes"] = *input.Notes
		}
		if input.Status != nil {
			updates["status"] = *input.Status
		}
		if err := tx.Model(&item).Updates(updates).Error; err != nil {
			return err
		}
		if input.OwnerTeamID != nil || input.ParticipantTeamIDs != nil {
			if err := tx.Where("large_work_item_id = ?", item.ID).Delete(&models.LargeWorkItemTeam{}).Error; err != nil {
				return err
			}
			ownerID := item.OwnerTeamID
			if input.OwnerTeamID != nil {
				ownerID = *input.OwnerTeamID
			}
			participantIDs := uniqueTeamIDs(append(input.ParticipantTeamIDs, ownerID))
			teams := make([]models.LargeWorkItemTeam, 0, len(participantIDs))
			for _, teamID := range participantIDs {
				role := entity.LargeWorkTeamRoleParticipant
				if teamID == ownerID {
					role = entity.LargeWorkTeamRoleOwner
				}
				teams = append(teams, models.LargeWorkItemTeam{LargeWorkItemID: item.ID, TeamID: teamID, Role: role, ParticipantStatus: entity.LargeWorkParticipantStatusAssigned})
			}
			if len(teams) > 0 {
				if err := tx.Create(&teams).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Preload("Teams.Team").First(&item, input.ID).Error; err != nil {
		return nil, err
	}
	out := mapItem(item)
	return &out, nil
}

func (r *repository) ReplaceTasks(ctx context.Context, input ReplaceTasksInput) ([]entity.LargeWorkTask, error) {
	rows, err := buildTaskRows(input.LargeWorkItemID, input.Tasks, time.Now())
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return replaceTaskRows(tx, input.LargeWorkItemID, rows)
	})
	if err != nil {
		return nil, mapTaskSchemaError(err)
	}
	return r.ListTasksByPlan(ctx, ListTasksQuery{LargeWorkItemID: input.LargeWorkItemID})
}

func (r *repository) ReplaceTasksAndParticipants(ctx context.Context, input ReplaceTasksAndParticipantsInput) ([]entity.LargeWorkTask, error) {
	rows, err := buildTaskRows(input.LargeWorkItemID, input.Tasks, time.Now())
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("large_work_item_id = ?", input.LargeWorkItemID).Delete(&models.LargeWorkItemTeam{}).Error; err != nil {
			return err
		}
		participantIDs := uniqueTeamIDs(append(input.ParticipantTeamIDs, input.OwnerTeamID))
		teams := make([]models.LargeWorkItemTeam, 0, len(participantIDs))
		for _, teamID := range participantIDs {
			role := entity.LargeWorkTeamRoleParticipant
			if teamID == input.OwnerTeamID {
				role = entity.LargeWorkTeamRoleOwner
			}
			teams = append(teams, models.LargeWorkItemTeam{LargeWorkItemID: input.LargeWorkItemID, TeamID: teamID, Role: role, ParticipantStatus: entity.LargeWorkParticipantStatusAssigned})
		}
		if len(teams) > 0 {
			if err := tx.Create(&teams).Error; err != nil {
				return err
			}
		}
		return replaceTaskRows(tx, input.LargeWorkItemID, rows)
	})
	if err != nil {
		return nil, mapTaskSchemaError(err)
	}
	return r.ListTasksByPlan(ctx, ListTasksQuery{LargeWorkItemID: input.LargeWorkItemID})
}

func buildTaskRows(largeWorkItemID int64, tasks []TaskInput, now time.Time) ([]models.LargeWorkTask, error) {
	if len(tasks) == 0 {
		return nil, nil
	}
	rows := make([]models.LargeWorkTask, 0, len(tasks))
	for _, task := range tasks {
		row, err := taskInputToModel(largeWorkItemID, task, now)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func replaceTaskRows(tx *gorm.DB, largeWorkItemID int64, rows []models.LargeWorkTask) error {
	if err := tx.Where("large_work_item_id = ?", largeWorkItemID).Delete(&models.LargeWorkTask{}).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	return tx.Create(&rows).Error
}

func (r *repository) ListTasksByPlan(ctx context.Context, q ListTasksQuery) ([]entity.LargeWorkTask, error) {
	dbq := r.db.WithContext(ctx).Where("large_work_item_id = ?", q.LargeWorkItemID)
	if len(q.Statuses) > 0 {
		dbq = dbq.Where("status IN ?", q.Statuses)
	}
	var tasks []models.LargeWorkTask
	if err := dbq.Order("sequence ASC, id ASC").Find(&tasks).Error; err != nil {
		return nil, mapTaskSchemaError(err)
	}
	return mapTasks(tasks), nil
}

func (r *repository) ListAssignedTasks(ctx context.Context, q ListAssignedTasksQuery) ([]entity.LargeWorkTask, int64, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	limit := q.Limit
	if limit < 1 || limit > 100 {
		limit = 50
	}
	dbq := r.db.WithContext(ctx).
		Model(&models.LargeWorkTask{}).
		Joins("JOIN large_work_items ON large_work_items.id = large_work_tasks.large_work_item_id").
		Where("large_work_tasks.assigned_team_id = ?", q.AssignedTeamID).
		Where("large_work_items.deleted_at IS NULL").
		Where("large_work_items.status NOT IN ?", []string{entity.LargeWorkStatusCompleted, entity.LargeWorkStatusCancelled})
	if len(q.Statuses) > 0 {
		dbq = dbq.Where("large_work_tasks.status IN ?", q.Statuses)
	}
	var total int64
	if err := dbq.Count(&total).Error; err != nil {
		return nil, 0, mapTaskSchemaError(err)
	}
	var tasks []models.LargeWorkTask
	if err := dbq.Order("large_work_item_id ASC, sequence ASC, id ASC").Offset((page - 1) * limit).Limit(limit).Find(&tasks).Error; err != nil {
		return nil, 0, mapTaskSchemaError(err)
	}
	return mapTasks(tasks), total, nil
}

func (r *repository) GetTaskByID(ctx context.Context, q GetTaskQuery) (*entity.LargeWorkTask, error) {
	var task models.LargeWorkTask
	if err := r.db.WithContext(ctx).First(&task, q.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, mapTaskSchemaError(err)
	}
	out := mapTask(task)
	return &out, nil
}

func (r *repository) UpdateTask(ctx context.Context, input UpdateTaskInput) (*entity.LargeWorkTask, error) {
	var task models.LargeWorkTask
	if err := r.db.WithContext(ctx).First(&task, input.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, mapTaskSchemaError(err)
	}
	updates := map[string]any{"updated_at": time.Now()}
	if input.Status != nil {
		updates["status"] = *input.Status
	}
	if input.BeforePhotoURLs != nil {
		updates["before_photo_urls"] = models.StringArray(input.BeforePhotoURLs)
	}
	if input.AfterPhotoURLs != nil {
		updates["after_photo_urls"] = models.StringArray(input.AfterPhotoURLs)
	}
	if input.CompletionNote != nil {
		updates["completion_note"] = input.CompletionNote
	}
	if input.Notes != nil {
		updates["notes"] = input.Notes
	}
	if input.StartedByUserID != nil {
		updates["started_by_user_id"] = input.StartedByUserID
	}
	if input.StartedAt != nil {
		updates["started_at"] = input.StartedAt
	}
	if input.CompletedByUserID != nil {
		updates["completed_by_user_id"] = input.CompletedByUserID
	}
	if input.CompletedAt != nil {
		updates["completed_at"] = input.CompletedAt
	}
	if err := r.db.WithContext(ctx).Model(&task).Updates(updates).Error; err != nil {
		return nil, mapTaskSchemaError(err)
	}
	if err := r.db.WithContext(ctx).First(&task, input.ID).Error; err != nil {
		return nil, mapTaskSchemaError(err)
	}
	out := mapTask(task)
	return &out, nil
}

func taskInputToModel(largeWorkItemID int64, input TaskInput, now time.Time) (models.LargeWorkTask, error) {
	metadata := []byte(`{}`)
	if input.Metadata != nil {
		encoded, err := json.Marshal(input.Metadata)
		if err != nil {
			return models.LargeWorkTask{}, err
		}
		metadata = encoded
	}
	status := input.Status
	if status == "" {
		status = entity.LargeWorkTaskStatusTodo
	}
	return models.LargeWorkTask{
		LargeWorkItemID:   largeWorkItemID,
		AssignedTeamID:    input.AssignedTeamID,
		Sequence:          input.Sequence,
		PointLabel:        input.PointLabel,
		Latitude:          floatPtrToDecimal(input.Latitude),
		Longitude:         floatPtrToDecimal(input.Longitude),
		WorkType:          input.WorkType,
		WorkDetail:        input.WorkDetail,
		PointCount:        input.PointCount,
		TreeCount:         input.TreeCount,
		ItemCount:         input.ItemCount,
		Notes:             input.Notes,
		Status:            status,
		BeforePhotoURLs:   stringArrayOrEmpty(input.BeforePhotoURLs),
		AfterPhotoURLs:    stringArrayOrEmpty(input.AfterPhotoURLs),
		CompletionNote:    input.CompletionNote,
		StartedByUserID:   input.StartedByUserID,
		StartedAt:         input.StartedAt,
		CompletedByUserID: input.CompletedByUserID,
		CompletedAt:       input.CompletedAt,
		Metadata:          metadata,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func stringArrayOrEmpty(values []string) models.StringArray {
	if values == nil {
		return models.StringArray{}
	}
	return models.StringArray(values)
}

func floatPtrToDecimal(value *float64) *decimal.Decimal {
	if value == nil {
		return nil
	}
	out := decimal.NewFromFloat(*value)
	return &out
}

func uniqueTeamIDs(ids []int64) []int64 {
	seen := map[int64]struct{}{}
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func mapItems(items []models.LargeWorkItem) []entity.LargeWorkItem {
	out := make([]entity.LargeWorkItem, 0, len(items))
	for _, item := range items {
		out = append(out, mapItem(item))
	}
	return out
}

func mapTasks(tasks []models.LargeWorkTask) []entity.LargeWorkTask {
	out := make([]entity.LargeWorkTask, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, mapTask(task))
	}
	return out
}

func mapItem(item models.LargeWorkItem) entity.LargeWorkItem {
	out := entity.LargeWorkItem{
		ID:                item.ID,
		OwnerTeamID:       item.OwnerTeamID,
		CreatedByUserID:   item.CreatedByUserID,
		Title:             item.Title,
		WorkType:          item.WorkType,
		StartDate:         item.StartDate,
		EndDate:           item.EndDate,
		WorkTime:          item.WorkTime,
		LocationText:      item.LocationText,
		PEAID:             item.PEAID,
		OperationCenterID: item.OperationCenterID,
		FeederID:          item.FeederID,
		StationID:         item.StationID,
		Notes:             item.Notes,
		Status:            item.Status,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
		DeletedAt:         item.DeletedAt,
	}
	teams := make([]entity.LargeWorkTeam, 0, len(item.Teams))
	for _, team := range item.Teams {
		teamName := ""
		if team.Team != nil {
			teamName = team.Team.Name
		}
		teams = append(teams, entity.LargeWorkTeam{ID: team.TeamID, Name: teamName, Role: team.Role, ParticipantStatus: team.ParticipantStatus})
	}
	sort.SliceStable(teams, func(i, j int) bool {
		if teams[i].Role == teams[j].Role {
			return teams[i].ID < teams[j].ID
		}
		return teams[i].Role == entity.LargeWorkTeamRoleOwner
	})
	out.Teams = teams
	return out
}

func mapTaskSchemaError(err error) error {
	if err == nil {
		return nil
	}
	if isSchemaUnavailableError(err) {
		return errors.Join(ErrSchemaUnavailable, err)
	}
	return err
}

func isSchemaUnavailableError(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "sqlstate 42p01") || strings.Contains(lower, "sqlstate 42703") || strings.Contains(lower, "no such table") {
		return true
	}
	return strings.Contains(lower, "large_work_tasks") && (strings.Contains(lower, "does not exist") || strings.Contains(lower, "missing"))
}

func mapTask(task models.LargeWorkTask) entity.LargeWorkTask {
	metadata := map[string]any{}
	if len(task.Metadata) > 0 {
		_ = json.Unmarshal(task.Metadata, &metadata)
	}
	return entity.LargeWorkTask{
		ID:                task.ID,
		LargeWorkItemID:   task.LargeWorkItemID,
		AssignedTeamID:    task.AssignedTeamID,
		Sequence:          task.Sequence,
		PointLabel:        task.PointLabel,
		Latitude:          decimalToFloatPtr(task.Latitude),
		Longitude:         decimalToFloatPtr(task.Longitude),
		WorkType:          task.WorkType,
		WorkDetail:        task.WorkDetail,
		PointCount:        task.PointCount,
		TreeCount:         task.TreeCount,
		ItemCount:         task.ItemCount,
		Notes:             task.Notes,
		Status:            task.Status,
		BeforePhotoURLs:   []string(task.BeforePhotoURLs),
		AfterPhotoURLs:    []string(task.AfterPhotoURLs),
		CompletionNote:    task.CompletionNote,
		StartedByUserID:   task.StartedByUserID,
		StartedAt:         task.StartedAt,
		CompletedByUserID: task.CompletedByUserID,
		CompletedAt:       task.CompletedAt,
		Metadata:          metadata,
		CreatedAt:         task.CreatedAt,
		UpdatedAt:         task.UpdatedAt,
	}
}

func decimalToFloatPtr(v *decimal.Decimal) *float64 {
	if v == nil {
		return nil
	}
	f, _ := v.Float64()
	return &f
}
