package dto

import internaldto "backend-hotlines3/internal/dto"

type StandardResponse = internaldto.StandardResponse
type ErrorInfo = internaldto.ErrorInfo
type Meta = internaldto.Meta
type Count = internaldto.Count

type CreateTaskRequest struct {
	WorkDate    string   `json:"workDate" binding:"required"`
	TeamID      int64    `json:"teamId" binding:"required"`
	JobTypeID   int64    `json:"jobTypeId" binding:"required"`
	JobDetailID int64    `json:"jobDetailId" binding:"required"`
	FeederID    *int64   `json:"feederId"`
	NumPole     *string  `json:"numPole"`
	DeviceCode  *string  `json:"deviceCode"`
	Detail      *string  `json:"detail"`
	URLsBefore  []string `json:"urlsBefore"`
	URLsAfter   []string `json:"urlsAfter"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

type UpdateTaskRequest struct {
	WorkDate    *string  `json:"workDate"`
	TeamID      *int64   `json:"teamId"`
	JobTypeID   *int64   `json:"jobTypeId"`
	JobDetailID *int64   `json:"jobDetailId"`
	FeederID    *int64   `json:"feederId"`
	NumPole     *string  `json:"numPole"`
	DeviceCode  *string  `json:"deviceCode"`
	Detail      *string  `json:"detail"`
	URLsBefore  []string `json:"urlsBefore"`
	URLsAfter   []string `json:"urlsAfter"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

type TaskResponse struct {
	ID          int64                `json:"id"`
	WorkDate    string               `json:"workDate"`
	TeamID      int64                `json:"teamId"`
	JobTypeID   int64                `json:"jobTypeId"`
	JobDetailID int64                `json:"jobDetailId"`
	FeederID    *int64               `json:"feederId"`
	NumPole     *string              `json:"numPole"`
	DeviceCode  *string              `json:"deviceCode"`
	Detail      *string              `json:"detail"`
	URLsBefore  []string             `json:"urlsBefore"`
	URLsAfter   []string             `json:"urlsAfter"`
	Latitude    *float64             `json:"latitude"`
	Longitude   *float64             `json:"longitude"`
	Team        *TeamNested          `json:"team,omitempty"`
	JobType     *JobTypeNested       `json:"jobType,omitempty"`
	JobDetail   *JobDetailNested     `json:"jobDetail,omitempty"`
	Feeder      *FeederNestedForTask `json:"feeder,omitempty"`
	CreatedAt   string               `json:"createdAt"`
	UpdatedAt   string               `json:"updatedAt"`
	DeletedAt   *string              `json:"deletedAt"`
}

type TeamNested struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type JobTypeNested struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type JobDetailNested struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type FeederNestedForTask struct {
	ID      int64                `json:"id"`
	Code    string               `json:"code"`
	Station *StationNestedSimple `json:"station,omitempty"`
}

type StationNestedSimple struct {
	Name            string                 `json:"name"`
	OperationCenter *OperationCenterNested `json:"operationCenter,omitempty"`
}

type OperationCenterNested struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type TasksByTeamResponse struct {
	Team  TeamNested     `json:"team"`
	Tasks []TaskResponse `json:"tasks"`
}
