package dto

import internaldto "backend-hotlines3/internal/dto"

type StandardResponse = internaldto.StandardResponse
type ErrorInfo = internaldto.ErrorInfo
type Count = internaldto.Count
type JobDetailResponse = internaldto.JobDetailResponse

type CreateRequest struct {
	Name      string `json:"name" binding:"required"`
	JobTypeID *int64 `json:"jobTypeId"`
}

type UpdateRequest struct {
	Name      string `json:"name"`
	JobTypeID *int64 `json:"jobTypeId"`
}
