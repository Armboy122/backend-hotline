package mapper

import (
	"time"

	"backend-hotlines3/internal/feature/monthlyplan/dto"
	"backend-hotlines3/internal/feature/monthlyplan/entity"
)

func ToPlanFileResponse(f entity.PlanFileEntity) dto.PlanFileResponse {
	resp := dto.PlanFileResponse{
		ID:            f.ID,
		MonthlyPlanID: f.MonthlyPlanID,
		TeamID:        f.TeamID,
		UploadedByID:  f.UploadedByID,
		FileKey:       f.FileKey,
		FileURL:       f.FileURL,
		FileName:      f.FileName,
		FileSizeBytes: f.FileSizeBytes,
		Description:   f.Description,
		Destination:   f.Destination,
		Remarks:       f.Remarks,
		IsMasterPlan:  f.IsMasterPlan,
		IsDeleted:     f.IsDeleted,
		CreatedAt:     f.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     f.UpdatedAt.Format(time.RFC3339),
	}
	if f.WorkStartDate != nil {
		workStartDate := f.WorkStartDate.Format("2006-01-02")
		resp.WorkStartDate = &workStartDate
	}
	if f.WorkEndDate != nil {
		workEndDate := f.WorkEndDate.Format("2006-01-02")
		resp.WorkEndDate = &workEndDate
	}
	if f.DeletedAt != nil {
		deletedAt := f.DeletedAt.Format(time.RFC3339)
		resp.DeletedAt = &deletedAt
	}
	if f.TeamName != nil {
		teamID := int64(0)
		if f.TeamID != nil {
			teamID = *f.TeamID
		}
		resp.Team = &dto.TeamNested{ID: teamID, Name: *f.TeamName}
	}
	if f.UploaderName != nil {
		resp.UploadedBy = &dto.PlanFileUploaderNested{
			ID:       uint(f.UploadedByID),
			Username: *f.UploaderName,
		}
	}
	return resp
}
