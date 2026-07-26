package mapper

import (
	"backend-hotlines3/internal/feature/team/dto"
	"backend-hotlines3/internal/feature/team/entity"
	"backend-hotlines3/internal/models"
)

func FromModel(model models.Team, tasks int64) entity.Entity {
	return entity.Entity{
		ID:                 model.ID,
		Name:               model.Name,
		Code:               model.Code,
		BaseArea:           model.BaseArea,
		CrewType:           model.CrewType,
		DisplayOrder:       model.DisplayOrder,
		MonthlyPlanVisible: model.MonthlyPlanVisible,
		Tasks:              tasks,
	}
}

func ToResponse(team entity.Entity) dto.TeamResponse {
	return dto.TeamResponse{
		ID:                 team.ID,
		Name:               team.Name,
		Code:               team.Code,
		BaseArea:           team.BaseArea,
		CrewType:           team.CrewType,
		DisplayOrder:       team.DisplayOrder,
		MonthlyPlanVisible: team.MonthlyPlanVisible,
		Count: &dto.Count{
			Tasks: team.Tasks,
		},
	}
}
