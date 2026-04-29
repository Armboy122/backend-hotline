package mapper

import (
	"backend-hotlines3/internal/feature/operationcenter/dto"
	"backend-hotlines3/internal/feature/operationcenter/entity"
	"backend-hotlines3/internal/models"
)

func FromModel(m models.OperationCenter) entity.Entity {
	return entity.Entity{
		ID:   m.ID,
		Name: m.Name,
	}
}

func ToResponse(e entity.Entity) dto.OperationCenterResponse {
	return dto.OperationCenterResponse{
		ID:   e.ID,
		Name: e.Name,
	}
}
