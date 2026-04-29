package dto

import internaldto "backend-hotlines3/internal/dto"

type StandardResponse = internaldto.StandardResponse
type ErrorInfo = internaldto.ErrorInfo
type OperationCenterResponse = internaldto.OperationCenterResponse

type UpsertRequest struct {
	Name string `json:"name" binding:"required"`
}
