package mapper

import (
	"time"

	"backend-hotlines3/internal/feature/jobdetail/dto"
	"backend-hotlines3/internal/feature/jobdetail/entity"
	"backend-hotlines3/internal/models"
)

func FromModel(m models.JobDetail, tasks int64) entity.Entity {
	return entity.Entity{
		ID:        m.ID,
		Name:      m.Name,
		JobTypeID: m.JobTypeID,
		Tasks:     tasks,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
	}
}

func ToResponse(e entity.Entity) dto.JobDetailResponse {
	var deletedAt *string
	if e.DeletedAt != nil {
		formatted := e.DeletedAt.Format(time.RFC3339)
		deletedAt = &formatted
	}
	return dto.JobDetailResponse{
		ID:        e.ID,
		Name:      e.Name,
		JobTypeID: e.JobTypeID,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
		DeletedAt: deletedAt,
		Count: &dto.Count{
			Tasks: e.Tasks,
		},
	}
}
