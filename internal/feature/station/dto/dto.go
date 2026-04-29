package dto

import internaldto "backend-hotlines3/internal/dto"

type StandardResponse = internaldto.StandardResponse
type ErrorInfo = internaldto.ErrorInfo
type OperationCenterNested = internaldto.OperationCenterNested
type StationResponse = internaldto.StationResponse

type CreateRequest struct {
	Name        string `json:"name" binding:"required"`
	CodeName    string `json:"codeName" binding:"required"`
	OperationID int64  `json:"operationId" binding:"required"`
}

type UpdateRequest struct {
	Name        string `json:"name"`
	CodeName    string `json:"codeName"`
	OperationID int64  `json:"operationId"`
}
