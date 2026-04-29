package mapper

import (
	"backend-hotlines3/internal/feature/team/dto"
	"backend-hotlines3/internal/feature/team/entity"
	"backend-hotlines3/internal/models"
)

func FromModel(model models.Team, tasks int64) entity.Entity {
	return entity.Entity{
		ID:    model.ID,
		Name:  model.Name,
		Tasks: tasks,
	}
}

func ToResponse(team entity.Entity) dto.TeamResponse {
	return dto.TeamResponse{
		ID:   team.ID,
		Name: team.Name,
		Count: &dto.Count{
			Tasks: team.Tasks,
		},
	}
}
