package mapper

import (
	"backend-hotlines3/internal/feature/station/dto"
	"backend-hotlines3/internal/feature/station/entity"
	"backend-hotlines3/internal/models"
)

func FromModel(model models.Station) entity.Entity {
	out := entity.Entity{
		ID:          model.ID,
		Name:        model.Name,
		CodeName:    model.CodeName,
		OperationID: model.OperationID,
	}
	if model.OperationCenter != nil {
		out.OperationCenter = &entity.OperationCenter{
			ID:   model.OperationCenter.ID,
			Name: model.OperationCenter.Name,
		}
	}
	return out
}

func ToResponse(station entity.Entity) dto.StationResponse {
	resp := dto.StationResponse{
		ID:          station.ID,
		Name:        station.Name,
		CodeName:    station.CodeName,
		OperationID: station.OperationID,
	}
	if station.OperationCenter != nil {
		resp.OperationCenter = &dto.OperationCenterNested{
			ID:   station.OperationCenter.ID,
			Name: station.OperationCenter.Name,
		}
	}
	return resp
}
