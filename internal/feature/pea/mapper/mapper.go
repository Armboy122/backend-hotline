package mapper

import (
	"backend-hotlines3/internal/feature/pea/dto"
	"backend-hotlines3/internal/feature/pea/entity"
	"backend-hotlines3/internal/models"
)

func FromModel(m models.PEA) entity.Entity {
	e := entity.Entity{
		ID:          m.ID,
		Shortname:   m.Shortname,
		Fullname:    m.Fullname,
		OperationID: m.OperationID,
	}
	if m.OperationCenter != nil {
		e.OperationCenter = &entity.NestedOperationCenter{
			ID:   m.OperationCenter.ID,
			Name: m.OperationCenter.Name,
		}
	}
	return e
}

func ToResponse(e entity.Entity) dto.PEAResponse {
	resp := dto.PEAResponse{
		ID:          e.ID,
		Shortname:   e.Shortname,
		Fullname:    e.Fullname,
		OperationID: e.OperationID,
	}
	if e.OperationCenter != nil {
		resp.OperationCenter = &dto.OperationCenterNested{
			ID:   e.OperationCenter.ID,
			Name: e.OperationCenter.Name,
		}
	}
	return resp
}
