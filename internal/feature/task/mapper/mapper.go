package mapper

import (
	"time"

	taskdto "backend-hotlines3/internal/feature/task/dto"
	"backend-hotlines3/internal/feature/task/entity"
	"backend-hotlines3/internal/models"
)

func FromModel(task models.TaskDaily) entity.Task {
	out := entity.Task{
		ID:          task.ID,
		WorkDate:    task.WorkDate,
		TeamID:      task.TeamID,
		JobTypeID:   task.JobTypeID,
		JobDetailID: task.JobDetailID,
		FeederID:    task.FeederID,
		NumPole:     task.NumPole,
		DeviceCode:  task.DeviceCode,
		Detail:      task.Detail,
		URLsBefore:  []string(task.URLsBefore),
		URLsAfter:   []string(task.URLsAfter),
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		DeletedAt:   task.DeletedAt,
	}

	if task.Latitude != nil {
		lat, _ := task.Latitude.Float64()
		out.Latitude = &lat
	}
	if task.Longitude != nil {
		lng, _ := task.Longitude.Float64()
		out.Longitude = &lng
	}
	if task.Team != nil {
		out.TeamName = &task.Team.Name
	}
	if task.JobType != nil {
		out.JobTypeName = &task.JobType.Name
	}
	if task.JobDetail != nil {
		out.JobDetailName = &task.JobDetail.Name
	}
	if task.Feeder != nil {
		out.FeederCode = &task.Feeder.Code
		if task.Feeder.Station != nil {
			out.StationName = &task.Feeder.Station.Name
			if task.Feeder.Station.OperationCenter != nil {
				out.OperationCenterID = &task.Feeder.Station.OperationCenter.ID
				out.OperationCenterName = &task.Feeder.Station.OperationCenter.Name
			}
		}
	}

	return out
}

func ToResponse(task entity.Task) taskdto.TaskResponse {
	response := taskdto.TaskResponse{
		ID:          task.ID,
		WorkDate:    task.WorkDate.Format("2006-01-02"),
		TeamID:      task.TeamID,
		JobTypeID:   task.JobTypeID,
		JobDetailID: task.JobDetailID,
		FeederID:    task.FeederID,
		NumPole:     task.NumPole,
		DeviceCode:  task.DeviceCode,
		Detail:      task.Detail,
		URLsBefore:  task.URLsBefore,
		URLsAfter:   task.URLsAfter,
		Latitude:    task.Latitude,
		Longitude:   task.Longitude,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}

	if task.DeletedAt != nil {
		formatted := task.DeletedAt.Format(time.RFC3339)
		response.DeletedAt = &formatted
	}
	if task.TeamName != nil {
		response.Team = &taskdto.TeamNested{ID: task.TeamID, Name: *task.TeamName}
	}
	if task.JobTypeName != nil {
		response.JobType = &taskdto.JobTypeNested{ID: task.JobTypeID, Name: *task.JobTypeName}
	}
	if task.JobDetailName != nil {
		response.JobDetail = &taskdto.JobDetailNested{ID: task.JobDetailID, Name: *task.JobDetailName}
	}
	if task.FeederID != nil && task.FeederCode != nil {
		response.Feeder = &taskdto.FeederNestedForTask{
			ID:   *task.FeederID,
			Code: *task.FeederCode,
		}
		if task.StationName != nil {
			response.Feeder.Station = &taskdto.StationNestedSimple{
				Name: *task.StationName,
			}
			if task.OperationCenterID != nil && task.OperationCenterName != nil {
				response.Feeder.Station.OperationCenter = &taskdto.OperationCenterNested{
					ID:   *task.OperationCenterID,
					Name: *task.OperationCenterName,
				}
			}
		}
	}

	return response
}

func GroupByTeam(tasks []entity.Task) []taskdto.TasksByTeamResponse {
	teamMap := make(map[string]taskdto.TasksByTeamResponse)
	order := make([]string, 0, len(tasks))

	for _, task := range tasks {
		teamName := "Unknown"
		teamID := int64(0)
		if task.TeamName != nil {
			teamName = *task.TeamName
			teamID = task.TeamID
		}
		if _, exists := teamMap[teamName]; !exists {
			teamMap[teamName] = taskdto.TasksByTeamResponse{
				Team:  taskdto.TeamNested{ID: teamID, Name: teamName},
				Tasks: []taskdto.TaskResponse{},
			}
			order = append(order, teamName)
		}
		entry := teamMap[teamName]
		entry.Tasks = append(entry.Tasks, ToResponse(task))
		teamMap[teamName] = entry
	}

	response := make([]taskdto.TasksByTeamResponse, 0, len(order))
	for _, name := range order {
		response = append(response, teamMap[name])
	}
	return response
}
