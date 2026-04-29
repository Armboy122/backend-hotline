package mapper

import (
	"backend-hotlines3/internal/feature/feeder/dto"
	"backend-hotlines3/internal/feature/feeder/entity"
	"backend-hotlines3/internal/models"
)

func FromModel(m models.Feeder, tasks int64) entity.Entity {
	e := entity.Entity{
		ID:        m.ID,
		Code:      m.Code,
		StationID: m.StationID,
		Tasks:     tasks,
	}
	if m.Station != nil {
		nested := &entity.NestedStation{
			ID:       m.Station.ID,
			Name:     m.Station.Name,
			CodeName: m.Station.CodeName,
		}
		if m.Station.OperationCenter != nil {
			nested.OperationCenter = &entity.NestedOperationCenter{
				ID:   m.Station.OperationCenter.ID,
				Name: m.Station.OperationCenter.Name,
			}
		}
		e.Station = nested
	}
	return e
}

func ToResponse(e entity.Entity) dto.FeederResponse {
	resp := dto.FeederResponse{
		ID:        e.ID,
		Code:      e.Code,
		StationID: e.StationID,
		Count: &dto.Count{
			Tasks: e.Tasks,
		},
	}
	if e.Station != nil {
		nested := &dto.StationNested{
			ID:       e.Station.ID,
			Name:     e.Station.Name,
			CodeName: e.Station.CodeName,
		}
		if e.Station.OperationCenter != nil {
			nested.OperationCenter = &dto.OperationCenterNested{
				ID:   e.Station.OperationCenter.ID,
				Name: e.Station.OperationCenter.Name,
			}
		}
		resp.Station = nested
	}
	return resp
}
