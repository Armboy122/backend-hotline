package repository

import (
	"context"
	"errors"
	"sort"
	"time"


	"backend-hotlines3/internal/feature/largework/entity"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

type ListQuery struct {
	Page    int
	Limit   int
	From    *time.Time
	To      *time.Time
	TeamID  *int64
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

type Repository interface {
	List(context.Context, ListQuery) ([]entity.LargeWorkItem, int64, error)
	GetByID(context.Context, GetQuery) (*entity.LargeWorkItem, error)
	Create(context.Context, CreateInput) (*entity.LargeWorkItem, error)
	Update(context.Context, UpdateInput) (*entity.LargeWorkItem, error)
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
		Offset((page-1)*limit).
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
		CreatedByUserID:    input.CreatedByUserID,
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

func mapItem(item models.LargeWorkItem) entity.LargeWorkItem {
	out := entity.LargeWorkItem{
		ID:                item.ID,
		OwnerTeamID:       item.OwnerTeamID,
		CreatedByUserID:    item.CreatedByUserID,
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
