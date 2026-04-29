package mapper

import (
	"backend-hotlines3/internal/feature/jobtype/dto"
	"backend-hotlines3/internal/feature/jobtype/entity"
	"backend-hotlines3/internal/models"
)

func FromModel(model models.JobType, tasks int64) entity.Entity {
	return entity.Entity{
		ID:    model.ID,
		Name:  model.Name,
		Tasks: tasks,
	}
}

func ToResponse(jobType entity.Entity) dto.JobTypeResponse {
	return dto.JobTypeResponse{
		ID:   jobType.ID,
		Name: jobType.Name,
		Count: &dto.Count{
			Tasks: jobType.Tasks,
		},
	}
}
