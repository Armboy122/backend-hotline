package mapper

import (
	"time"

	"backend-hotlines3/internal/feature/monthlyschedule/dto"
	"backend-hotlines3/internal/feature/monthlyschedule/entity"
)

func WorkspaceToResponse(value entity.Workspace) dto.WorkspaceResponse {
	out := dto.WorkspaceResponse{
		Period: dto.PeriodResponse{Year: value.Period.Year, Month: value.Period.Month},
		Teams:  make([]dto.TeamResponse, 0, len(value.Teams)),
	}
	for _, team := range value.Teams {
		out.Teams = append(out.Teams, dto.TeamResponse{
			ID:                 team.ID,
			Name:               team.Name,
			Code:               team.Code,
			BaseArea:           team.BaseArea,
			CrewType:           team.CrewType,
			DisplayOrder:       team.DisplayOrder,
			MonthlyPlanVisible: team.MonthlyPlanVisible,
		})
	}
	out.Draft = scheduleToResponse(value.Draft)
	out.Published = scheduleToResponse(value.Published)
	return out
}

func ProjectionToResponse(value entity.Projection) dto.ProjectionResponse {
	out := dto.ProjectionResponse{
		Period:      dto.PeriodResponse{Year: value.Period.Year, Month: value.Period.Month},
		Revision:    dto.ProjectionRevisionResponse{ID: value.Revision.ID, No: value.Revision.No},
		PublishedAt: value.PublishedAt.Format(time.RFC3339),
		Checksum:    value.Checksum,
		Teams:       make([]dto.ProjectedTeamResponse, 0, len(value.Teams)),
	}
	for _, team := range value.Teams {
		mapped := dto.ProjectedTeamResponse{
			ID:           team.ID,
			Code:         team.Code,
			Name:         team.Name,
			BaseArea:     team.BaseArea,
			CrewType:     team.CrewType,
			DisplayOrder: team.DisplayOrder,
			Segments:     make([]dto.ProjectedSegmentResponse, 0, len(team.Segments)),
		}
		for _, segment := range team.Segments {
			mapped.Segments = append(mapped.Segments, dto.ProjectedSegmentResponse{
				AssignmentID:   segment.AssignmentID,
				AssignmentType: segment.AssignmentType,
				StartDate:      segment.StartDate,
				EndDate:        segment.EndDate,
				Destination:    segment.Destination,
				Note:           segment.Note,
				SourceType:     segment.SourceType,
				SourceID:       segment.SourceID,
			})
		}
		out.Teams = append(out.Teams, mapped)
	}
	return out
}

func scheduleToResponse(value *entity.Schedule) *dto.ScheduleResponse {
	if value == nil || value.Revision == nil {
		return nil
	}
	revision := dto.RevisionResponse{
		ID:         value.Revision.ID,
		RevisionNo: value.Revision.RevisionNo,
		Status:     value.Revision.Status,
		Checksum:   value.Revision.Checksum,
	}
	if value.Revision.PublishedAt != nil {
		formatted := value.Revision.PublishedAt.Format(time.RFC3339)
		revision.PublishedAt = &formatted
	}
	out := &dto.ScheduleResponse{
		Revision:    revision,
		Assignments: make([]dto.AssignmentResponse, 0, len(value.Assignments)),
	}
	for _, assignment := range value.Assignments {
		out.Assignments = append(out.Assignments, dto.AssignmentResponse{
			ID:             assignment.ID,
			TeamID:         assignment.TeamID,
			AssignmentType: assignment.AssignmentType,
			StartDate:      assignment.StartDate.Format(time.DateOnly),
			EndDate:        assignment.EndDate.Format(time.DateOnly),
			Destination:    assignment.Destination,
			Note:           assignment.Note,
			SourceType:     assignment.SourceType,
			SourceID:       assignment.SourceID,
		})
	}
	return out
}
