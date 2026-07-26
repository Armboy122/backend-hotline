package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend-hotlines3/internal/feature/monthlyschedule/entity"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormRepository struct {
	db *gorm.DB
}

func NewGORM(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) FindOrCreatePeriod(ctx context.Context, year, month int) (entity.Period, error) {
	var model models.MonthlyPlan
	err := r.db.WithContext(ctx).
		Where(`year = ? AND month = ?`, year, month).
		FirstOrCreate(&model, models.MonthlyPlan{Year: year, Month: month}).
		Error
	if err != nil {
		return entity.Period{}, fmt.Errorf("find or create monthly plan period: %w", err)
	}
	return mapPeriod(model), nil
}

func (r *gormRepository) FindPeriod(ctx context.Context, year, month int) (entity.Period, error) {
	var model models.MonthlyPlan
	err := r.db.WithContext(ctx).Where(`year = ? AND month = ?`, year, month).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.Period{}, entity.ErrPublishedNotFound
	}
	if err != nil {
		return entity.Period{}, fmt.Errorf("find monthly plan period: %w", err)
	}
	return mapPeriod(model), nil
}

func (r *gormRepository) ListVisibleTeams(ctx context.Context) ([]entity.Team, error) {
	var teams []models.Team
	err := r.db.WithContext(ctx).
		Where("monthly_plan_visible = ?", true).
		Order("display_order ASC, id ASC").
		Find(&teams).Error
	if err != nil {
		return nil, fmt.Errorf("list monthly plan teams: %w", err)
	}
	out := make([]entity.Team, 0, len(teams))
	for _, team := range teams {
		out = append(out, mapTeam(team))
	}
	return out, nil
}

func (r *gormRepository) GetSchedule(ctx context.Context, monthlyPlanID int64, status string) (*entity.Schedule, error) {
	var revision models.MonthlyPlanScheduleRevision
	err := r.db.WithContext(ctx).
		Where("monthly_plan_id = ? AND status = ?", monthlyPlanID, status).
		First(&revision).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if status == models.MonthlyPlanScheduleStatusDraft {
			return nil, entity.ErrDraftNotFound
		}
		return nil, entity.ErrPublishedNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get monthly schedule revision: %w", err)
	}
	return r.loadSchedule(ctx, r.db, revision)
}

func (r *gormRepository) ReplaceDraft(
	ctx context.Context,
	period entity.Period,
	actor entity.Actor,
	expectedRevisionNo *int,
	assignments []entity.Assignment,
) (*entity.Schedule, error) {
	var result *entity.Schedule
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var draft models.MonthlyPlanScheduleRevision
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("monthly_plan_id = ? AND status = ?", period.ID, models.MonthlyPlanScheduleStatusDraft).
			First(&draft).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			if expectedRevisionNo != nil && *expectedRevisionNo != 0 {
				return entity.ErrRevisionConflict
			}
			var maxRevision int
			if err := tx.Model(&models.MonthlyPlanScheduleRevision{}).
				Where("monthly_plan_id = ?", period.ID).
				Select("COALESCE(MAX(revision_no), 0)").
				Scan(&maxRevision).Error; err != nil {
				return fmt.Errorf("get next monthly schedule revision: %w", err)
			}
			draft = models.MonthlyPlanScheduleRevision{
				MonthlyPlanID:   period.ID,
				RevisionNo:      maxRevision + 1,
				Status:          models.MonthlyPlanScheduleStatusDraft,
				CreatedByUserID: actor.UserID,
			}
			if err := tx.Create(&draft).Error; err != nil {
				return fmt.Errorf("create monthly schedule draft: %w", err)
			}
		case err != nil:
			return fmt.Errorf("lock monthly schedule draft: %w", err)
		default:
			if expectedRevisionNo != nil && draft.RevisionNo != *expectedRevisionNo {
				return entity.ErrRevisionConflict
			}
		}

		if err := tx.Where("revision_id = ?", draft.ID).Delete(&models.MonthlyPlanTeamAssignment{}).Error; err != nil {
			return fmt.Errorf("replace monthly schedule assignments: %w", err)
		}
		modelsIn := make([]models.MonthlyPlanTeamAssignment, 0, len(assignments))
		for _, assignment := range assignments {
			modelsIn = append(modelsIn, models.MonthlyPlanTeamAssignment{
				RevisionID:     draft.ID,
				TeamID:         assignment.TeamID,
				AssignmentType: assignment.AssignmentType,
				StartDate:      assignment.StartDate,
				EndDate:        assignment.EndDate,
				Destination:    assignment.Destination,
				Note:           assignment.Note,
				SourceType:     assignment.SourceType,
				SourceID:       assignment.SourceID,
			})
		}
		if len(modelsIn) > 0 {
			if err := tx.Create(&modelsIn).Error; err != nil {
				return fmt.Errorf("create monthly schedule assignments: %w", err)
			}
		}
		if err := tx.Model(&draft).Update("updated_at", time.Now()).Error; err != nil {
			return fmt.Errorf("touch monthly schedule draft: %w", err)
		}
		loaded, err := r.loadSchedule(ctx, tx, draft)
		if err != nil {
			return err
		}
		result = loaded
		return nil
	})
	return result, err
}

func (r *gormRepository) PublishDraft(
	ctx context.Context,
	period entity.Period,
	actor entity.Actor,
	draftRevisionID int64,
	checksum string,
	projection []byte,
) (*entity.Schedule, error) {
	var result *entity.Schedule
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var draft models.MonthlyPlanScheduleRevision
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND monthly_plan_id = ? AND status = ?", draftRevisionID, period.ID, models.MonthlyPlanScheduleStatusDraft).
			First(&draft).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.ErrRevisionConflict
		}
		if err != nil {
			return fmt.Errorf("lock monthly schedule draft for publish: %w", err)
		}

		if err := tx.Model(&models.MonthlyPlanScheduleRevision{}).
			Where("monthly_plan_id = ? AND status = ?", period.ID, models.MonthlyPlanScheduleStatusPublished).
			Updates(map[string]any{
				"status":     models.MonthlyPlanScheduleStatusSuperseded,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return fmt.Errorf("supersede monthly schedule: %w", err)
		}

		now := time.Now()
		if err := tx.Model(&draft).Updates(map[string]any{
			"status":               models.MonthlyPlanScheduleStatusPublished,
			"published_by_user_id": actor.UserID,
			"published_at":         now,
			"checksum":             checksum,
			"projection":           projection,
			"updated_at":           now,
		}).Error; err != nil {
			return fmt.Errorf("publish monthly schedule: %w", err)
		}
		draft.Status = models.MonthlyPlanScheduleStatusPublished
		draft.PublishedByUserID = &actor.UserID
		draft.PublishedAt = &now
		draft.Checksum = &checksum
		draft.Projection = append([]byte(nil), projection...)
		loaded, err := r.loadSchedule(ctx, tx, draft)
		if err != nil {
			return err
		}
		result = loaded
		return nil
	})
	return result, err
}

func (r *gormRepository) loadSchedule(ctx context.Context, db *gorm.DB, revision models.MonthlyPlanScheduleRevision) (*entity.Schedule, error) {
	var assignments []models.MonthlyPlanTeamAssignment
	err := db.WithContext(ctx).
		Where("revision_id = ?", revision.ID).
		Order("team_id ASC, start_date ASC, id ASC").
		Find(&assignments).Error
	if err != nil {
		return nil, fmt.Errorf("load monthly schedule assignments: %w", err)
	}
	out := &entity.Schedule{Revision: mapRevision(revision), Assignments: make([]entity.Assignment, 0, len(assignments))}
	for _, assignment := range assignments {
		out.Assignments = append(out.Assignments, mapAssignment(assignment))
	}
	return out, nil
}

func mapPeriod(model models.MonthlyPlan) entity.Period {
	return entity.Period{ID: model.ID, Year: model.Year, Month: model.Month}
}

func mapTeam(model models.Team) entity.Team {
	return entity.Team{
		ID:                 model.ID,
		Name:               model.Name,
		Code:               model.Code,
		BaseArea:           model.BaseArea,
		CrewType:           model.CrewType,
		DisplayOrder:       model.DisplayOrder,
		MonthlyPlanVisible: model.MonthlyPlanVisible,
	}
}

func mapRevision(model models.MonthlyPlanScheduleRevision) *entity.Revision {
	return &entity.Revision{
		ID:                model.ID,
		MonthlyPlanID:     model.MonthlyPlanID,
		RevisionNo:        model.RevisionNo,
		Status:            model.Status,
		CreatedByUserID:   model.CreatedByUserID,
		PublishedByUserID: model.PublishedByUserID,
		PublishedAt:       model.PublishedAt,
		Checksum:          model.Checksum,
		Projection:        append([]byte(nil), model.Projection...),
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}

func mapAssignment(model models.MonthlyPlanTeamAssignment) entity.Assignment {
	return entity.Assignment{
		ID:             model.ID,
		RevisionID:     model.RevisionID,
		TeamID:         model.TeamID,
		AssignmentType: model.AssignmentType,
		StartDate:      model.StartDate,
		EndDate:        model.EndDate,
		Destination:    model.Destination,
		Note:           model.Note,
		SourceType:     model.SourceType,
		SourceID:       model.SourceID,
	}
}
