package dto

type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type PeriodResponse struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

type TeamResponse struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	Code               *string `json:"code"`
	BaseArea           *string `json:"baseArea"`
	CrewType           *string `json:"crewType"`
	DisplayOrder       int     `json:"displayOrder"`
	MonthlyPlanVisible bool    `json:"monthlyPlanVisible"`
}

type RevisionResponse struct {
	ID          int64   `json:"id"`
	RevisionNo  int     `json:"revisionNo"`
	Status      string  `json:"status"`
	PublishedAt *string `json:"publishedAt"`
	Checksum    *string `json:"checksum"`
}

type AssignmentResponse struct {
	ID             int64   `json:"id"`
	TeamID         int64   `json:"teamId"`
	AssignmentType string  `json:"assignmentType"`
	StartDate      string  `json:"startDate"`
	EndDate        string  `json:"endDate"`
	Destination    string  `json:"destination"`
	Note           *string `json:"note"`
	SourceType     string  `json:"sourceType"`
	SourceID       *int64  `json:"sourceId"`
}

type ScheduleResponse struct {
	Revision    RevisionResponse     `json:"revision"`
	Assignments []AssignmentResponse `json:"assignments"`
}

type WorkspaceResponse struct {
	Period    PeriodResponse    `json:"period"`
	Teams     []TeamResponse    `json:"teams"`
	Draft     *ScheduleResponse `json:"draft"`
	Published *ScheduleResponse `json:"published"`
}

type SaveDraftRequest struct {
	ExpectedRevisionNo *int                `json:"expectedRevisionNo"`
	Assignments        []AssignmentRequest `json:"assignments" binding:"required"`
}

type AssignmentRequest struct {
	TeamID         int64   `json:"teamId" binding:"required"`
	AssignmentType string  `json:"assignmentType" binding:"required"`
	StartDate      string  `json:"startDate" binding:"required"`
	EndDate        string  `json:"endDate" binding:"required"`
	Destination    string  `json:"destination" binding:"required"`
	Note           *string `json:"note"`
	SourceType     string  `json:"sourceType"`
	SourceID       *int64  `json:"sourceId"`
}

type PublishRequest struct {
	ExpectedRevisionNo int `json:"expectedRevisionNo" binding:"required,min=1"`
}

type ProjectionResponse struct {
	Period      PeriodResponse             `json:"period"`
	Revision    ProjectionRevisionResponse `json:"revision"`
	PublishedAt string                     `json:"publishedAt"`
	Checksum    string                     `json:"checksum"`
	Teams       []ProjectedTeamResponse    `json:"teams"`
}

type ProjectionRevisionResponse struct {
	ID int64 `json:"id"`
	No int   `json:"no"`
}

type ProjectedTeamResponse struct {
	ID           int64                      `json:"id"`
	Code         string                     `json:"code"`
	Name         string                     `json:"name"`
	BaseArea     string                     `json:"baseArea"`
	CrewType     string                     `json:"crewType"`
	DisplayOrder int                        `json:"displayOrder"`
	Segments     []ProjectedSegmentResponse `json:"segments"`
}

type ProjectedSegmentResponse struct {
	AssignmentID   *int64  `json:"assignmentId"`
	AssignmentType string  `json:"assignmentType"`
	StartDate      string  `json:"startDate"`
	EndDate        string  `json:"endDate"`
	Destination    string  `json:"destination"`
	Note           *string `json:"note"`
	SourceType     string  `json:"sourceType"`
	SourceID       *int64  `json:"sourceId"`
}
