package dto

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int64 `json:"total"`
}

type StandardResponse struct {
	Success bool        `json:"success"`
	Data    any         `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
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
	OperationCenterID *int64  `json:"operationCenterId"`
	FeederID          *int64  `json:"feederId"`
	StationID         *int64  `json:"stationId"`
	Notes             *string `json:"notes"`
}

type UpdateRequest struct {
	OwnerTeamID        *int64   `json:"ownerTeamId"`
	ParticipantTeamIDs []int64  `json:"participantTeamIds"`
	Title              *string  `json:"title"`
	WorkType           *string  `json:"workType"`
	StartDate          *string  `json:"startDate"`
	EndDate            *string  `json:"endDate"`
	WorkTime           *string  `json:"workTime"`
	LocationText       *string  `json:"locationText"`
	PEAID              *int64   `json:"peaId"`
	OperationCenterID *int64   `json:"operationCenterId"`
	FeederID          *int64   `json:"feederId"`
	StationID         *int64   `json:"stationId"`
	Notes             *string  `json:"notes"`
	Status            *string  `json:"status"`
}

type TeamRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type LargeWorkResponse struct {
	ID                int64      `json:"id"`
	OwnerTeamID       int64      `json:"ownerTeamId"`
	CreatedByUserID   int64      `json:"createdByUserId"`
	Title             string     `json:"title"`
	WorkType          *string    `json:"workType,omitempty"`
	StartDate         string     `json:"startDate"`
	EndDate           *string    `json:"endDate,omitempty"`
	WorkTime          *string    `json:"workTime,omitempty"`
	LocationText      string     `json:"locationText"`
	PEAID             *int64     `json:"peaId,omitempty"`
	OperationCenterID *int64     `json:"operationCenterId,omitempty"`
	FeederID          *int64     `json:"feederId,omitempty"`
	StationID         *int64     `json:"stationId,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
	Status            string     `json:"status"`
	Teams             []TeamRef  `json:"teams"`
	Actions           struct {
		CanEdit           bool `json:"canEdit"`
		CanCancel         bool `json:"canCancel"`
		CanStartDailyReport bool `json:"canStartDailyReport"`
	} `json:"actions"`
	CreatedAt         string     `json:"createdAt"`
	UpdatedAt         string     `json:"updatedAt"`
	DeletedAt         *string    `json:"deletedAt"`
}
