package dto

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type StandardResponse struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Meta    *Meta      `json:"meta,omitempty"`
	Error   *ErrorInfo `json:"error,omitempty"`
}

type CreateRequest struct {
	OwnerTeamID        int64   `json:"ownerTeamId" binding:"required"`
	ParticipantTeamIDs []int64 `json:"participantTeamIds" binding:"required"`
	Title              string  `json:"title" binding:"required"`
	WorkType           *string `json:"workType"`
	StartDate          string  `json:"startDate" binding:"required"`
	EndDate            *string `json:"endDate"`
	WorkTime           *string `json:"workTime"`
	LocationText       string  `json:"locationText" binding:"required"`
	PEAID              *int64  `json:"peaId"`
	OperationCenterID  *int64  `json:"operationCenterId"`
	FeederID           *int64  `json:"feederId"`
	StationID          *int64  `json:"stationId"`
	Notes              *string `json:"notes"`
}

type UpdateRequest struct {
	OwnerTeamID        *int64  `json:"ownerTeamId"`
	ParticipantTeamIDs []int64 `json:"participantTeamIds"`
	Title              *string `json:"title"`
	WorkType           *string `json:"workType"`
	StartDate          *string `json:"startDate"`
	EndDate            *string `json:"endDate"`
	WorkTime           *string `json:"workTime"`
	LocationText       *string `json:"locationText"`
	PEAID              *int64  `json:"peaId"`
	OperationCenterID  *int64  `json:"operationCenterId"`
	FeederID           *int64  `json:"feederId"`
	StationID          *int64  `json:"stationId"`
	Notes              *string `json:"notes"`
	Status             *string `json:"status"`
}

type TeamRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type LargeWorkResponse struct {
	ID                int64     `json:"id"`
	OwnerTeamID       int64     `json:"ownerTeamId"`
	CreatedByUserID   int64     `json:"createdByUserId"`
	Title             string    `json:"title"`
	WorkType          *string   `json:"workType,omitempty"`
	StartDate         string    `json:"startDate"`
	EndDate           *string   `json:"endDate,omitempty"`
	WorkTime          *string   `json:"workTime,omitempty"`
	LocationText      string    `json:"locationText"`
	PEAID             *int64    `json:"peaId,omitempty"`
	OperationCenterID *int64    `json:"operationCenterId,omitempty"`
	FeederID          *int64    `json:"feederId,omitempty"`
	StationID         *int64    `json:"stationId,omitempty"`
	Notes             *string   `json:"notes,omitempty"`
	Status            string    `json:"status"`
	Teams             []TeamRef `json:"teams"`
	Actions           struct {
		CanEdit             bool `json:"canEdit"`
		CanCancel           bool `json:"canCancel"`
		CanAssignTasks      bool `json:"canAssignTasks"`
		CanStartDailyReport bool `json:"canStartDailyReport"`
	} `json:"actions"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	DeletedAt *string `json:"deletedAt"`
}

type TaskPointRequest struct {
	AssignedTeamID int64          `json:"assignedTeamId" binding:"required"`
	Sequence       int            `json:"sequence"`
	PointLabel     string         `json:"pointLabel" binding:"required"`
	Latitude       *float64       `json:"latitude"`
	Longitude      *float64       `json:"longitude"`
	WorkType       string         `json:"workType" binding:"required"`
	WorkDetail     *string        `json:"workDetail"`
	PointCount     *int           `json:"pointCount"`
	TreeCount      *int           `json:"treeCount"`
	ItemCount      *int           `json:"itemCount"`
	Notes          *string        `json:"notes"`
	Metadata       map[string]any `json:"metadata"`
}

type ReplaceTasksRequest struct {
	Tasks []TaskPointRequest `json:"tasks" binding:"required"`
}

type TaskExecutionPhotoRequest struct {
	Kind string `json:"kind" binding:"required"`
	URL  string `json:"url" binding:"required"`
}

type CompleteTaskRequest struct {
	AfterPhotoURLs []string `json:"afterPhotoUrls"`
	CompletionNote string   `json:"completionNote" binding:"required"`
}

type BlockTaskRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type OverviewProgressResponse struct {
	Total      int `json:"total"`
	Todo       int `json:"todo"`
	InProgress int `json:"inProgress"`
	Done       int `json:"done"`
	Blocked    int `json:"blocked"`
	Cancelled  int `json:"cancelled"`
}

type TeamProgressResponse struct {
	AssignedTeamID int64 `json:"assignedTeamId"`
	Total          int   `json:"total"`
	Todo           int   `json:"todo"`
	InProgress     int   `json:"inProgress"`
	Done           int   `json:"done"`
	Blocked        int   `json:"blocked"`
	Cancelled      int   `json:"cancelled"`
}

type OverviewResponse struct {
	Plan         LargeWorkResponse        `json:"plan"`
	Progress     OverviewProgressResponse `json:"progress"`
	TeamProgress []TeamProgressResponse   `json:"teamProgress"`
}

type LargeWorkTaskResponse struct {
	ID                int64          `json:"id"`
	LargeWorkItemID   int64          `json:"largeWorkItemId"`
	AssignedTeamID    int64          `json:"assignedTeamId"`
	Sequence          int            `json:"sequence"`
	PointLabel        string         `json:"pointLabel"`
	Latitude          *float64       `json:"latitude,omitempty"`
	Longitude         *float64       `json:"longitude,omitempty"`
	WorkType          string         `json:"workType"`
	WorkDetail        *string        `json:"workDetail,omitempty"`
	PointCount        *int           `json:"pointCount,omitempty"`
	TreeCount         *int           `json:"treeCount,omitempty"`
	ItemCount         *int           `json:"itemCount,omitempty"`
	Notes             *string        `json:"notes,omitempty"`
	Status            string         `json:"status"`
	BeforePhotoURLs   []string       `json:"beforePhotoUrls"`
	AfterPhotoURLs    []string       `json:"afterPhotoUrls"`
	CompletionNote    *string        `json:"completionNote,omitempty"`
	StartedByUserID   *int64         `json:"startedByUserId,omitempty"`
	StartedAt         *string        `json:"startedAt,omitempty"`
	CompletedByUserID *int64         `json:"completedByUserId,omitempty"`
	CompletedAt       *string        `json:"completedAt,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}
