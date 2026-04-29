package mapper

import (
	"backend-hotlines3/internal/feature/dashboard/dto"
	"backend-hotlines3/internal/feature/dashboard/entity"
)

func ToSummaryResponse(e entity.SummaryResult) dto.DashboardSummaryResponse {
	var topTeam *dto.TopTeam
	if e.TopTeam != nil {
		topTeam = &dto.TopTeam{ID: e.TopTeam.ID, Name: e.TopTeam.Name, Count: e.TopTeam.Count}
	}
	return dto.DashboardSummaryResponse{
		TotalTasks:    e.TotalTasks,
		TotalJobTypes: e.TotalJobTypes,
		TotalFeeders:  e.TotalFeeders,
		TopTeam:       topTeam,
	}
}

func ToTopJobResponses(items []entity.TopJobResult) []dto.TopJobResponse {
	out := make([]dto.TopJobResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.TopJobResponse{ID: item.ID, Name: item.Name, Count: item.Count, JobTypeName: item.JobTypeName})
	}
	return out
}

func ToTopFeederResponses(items []entity.TopFeederResult) []dto.TopFeederResponse {
	out := make([]dto.TopFeederResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.TopFeederResponse{ID: item.ID, Code: item.Code, StationName: item.StationName, Count: item.Count})
	}
	return out
}

func ToFeederMatrixResponse(e entity.FeederMatrixResult) dto.FeederMatrixResponse {
	jds := make([]dto.JobDetailInMatrix, 0, len(e.JobDetails))
	for _, jd := range e.JobDetails {
		jds = append(jds, dto.JobDetailInMatrix{ID: jd.ID, Name: jd.Name, Count: jd.Count, JobTypeName: jd.JobTypeName})
	}
	return dto.FeederMatrixResponse{
		FeederID: e.FeederID, FeederCode: e.FeederCode,
		StationName: e.StationName, TotalCount: e.TotalCount, JobDetails: jds,
	}
}

func ToStatsResponse(e entity.StatsResult) dto.DashboardStatsResponse {
	tasksByFeeder := make([]dto.ChartItem, 0, len(e.Charts.TasksByFeeder))
	for _, c := range e.Charts.TasksByFeeder {
		tasksByFeeder = append(tasksByFeeder, dto.ChartItem{Name: c.Name, Value: c.Value})
	}
	tasksByJobType := make([]dto.ChartItem, 0, len(e.Charts.TasksByJobType))
	for _, c := range e.Charts.TasksByJobType {
		tasksByJobType = append(tasksByJobType, dto.ChartItem{Name: c.Name, Value: c.Value})
	}
	tasksByTeam := make([]dto.ChartItem, 0, len(e.Charts.TasksByTeam))
	for _, c := range e.Charts.TasksByTeam {
		tasksByTeam = append(tasksByTeam, dto.ChartItem{Name: c.Name, Value: c.Value})
	}
	tasksByDate := make([]dto.DateChartItem, 0, len(e.Charts.TasksByDate))
	for _, c := range e.Charts.TasksByDate {
		tasksByDate = append(tasksByDate, dto.DateChartItem{Date: c.Date, Count: c.Count})
	}
	return dto.DashboardStatsResponse{
		Summary: dto.DashboardStatsSummary{
			TotalTasks:  e.Summary.TotalTasks,
			ActiveTeams: e.Summary.ActiveTeams,
			TopJobType:  e.Summary.TopJobType,
			TopFeeder:   e.Summary.TopFeeder,
		},
		Charts: dto.DashboardCharts{
			TasksByFeeder:  tasksByFeeder,
			TasksByJobType: tasksByJobType,
			TasksByTeam:    tasksByTeam,
			TasksByDate:    tasksByDate,
		},
	}
}
