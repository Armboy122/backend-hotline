package dto

import "time"

type CreateRequest struct {
	TeamID            int64   `json:"teamId" binding:"required"`
	Title             string  `json:"title" binding:"required"`
	WorkType          *string `json:"workType"`
	StartDate         string  `json:"startDate" binding:"required"`
	EndDate           *string `json:"endDate"`
	WorkTime          *string `json:"workTime"`
	LocationText      string  `json:"locationText" binding:"required"`
	PEAID             *int64  `json:"peaId"`
	OperationCenterID *int64  `json:"operationCenterId"`
	FeederID          *int64  `json:"feederId"`
	StationID         *int64  `json:"stationId"`
	Notes             *string `json:"notes"`
}

type UpdateRequest struct {
	TeamID            *int64  `json:"teamId"`
	Title             *string `json:"title"`
	WorkType          *string `json:"workType"`
	StartDate         *string `json:"startDate"`
	EndDate           *string `json:"endDate"`
	WorkTime          *string `json:"workTime"`
	LocationText      *string `json:"locationText"`
	PEAID             *int64  `json:"peaId"`
	OperationCenterID *int64  `json:"operationCenterId"`
	FeederID          *int64  `json:"feederId"`
	StationID         *int64  `json:"stationId"`
	Notes             *string `json:"notes"`
	Status            *string `json:"status"`
}

type TeamPlanResponse struct {
	ID                int64              `json:"id"`
	TeamID            int64              `json:"teamId"`
	CreatedByUserID   int64              `json:"createdByUserId"`
	Title             string             `json:"title"`
	WorkType          *string            `json:"workType,omitempty"`
	StartDate         string             `json:"startDate"`
	EndDate           *string            `json:"endDate,omitempty"`
	WorkTime          *string            `json:"workTime,omitempty"`
	LocationText      string             `json:"locationText"`
	PEAID             *int64             `json:"peaId,omitempty"`
	OperationCenterID *int64             `json:"operationCenterId,omitempty"`
	FeederID          *int64             `json:"feederId,omitempty"`
	StationID         *int64             `json:"stationId,omitempty"`
	Notes             *string            `json:"notes,omitempty"`
	Status            string             `json:"status"`
	DailyTaskID       *int64             `json:"dailyTaskId,omitempty"`
	CreatedAt         string             `json:"createdAt"`
	UpdatedAt         string             `json:"updatedAt"`
	DeletedAt         *string            `json:"deletedAt,omitempty"`
	Team              *TeamSummaryRef    `json:"team,omitempty"`
	CreatedBy         *UserSummaryRef    `json:"createdBy,omitempty"`
	Actions           TeamPlanActionRef  `json:"actions"`
}

type TeamSummaryRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type UserSummaryRef struct {
	ID          int64   `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"displayName,omitempty"`
}

type TeamPlanActionRef struct {
	CanEdit   bool `json:"canEdit"`
	CanDelete bool `json:"canDelete"`
}

func DateString(t time.Time) string { return t.UTC().Format(time.RFC3339) }
