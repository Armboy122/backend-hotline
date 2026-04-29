package dto

import internaldto "backend-hotlines3/internal/dto"

type StandardResponse = internaldto.StandardResponse
type ErrorInfo = internaldto.ErrorInfo
type Count = internaldto.Count
type FeederResponse = internaldto.FeederResponse
type StationNested = internaldto.StationNested
type OperationCenterNested = internaldto.OperationCenterNested

type CreateRequest struct {
	Code      string `json:"code" binding:"required"`
	StationID int64  `json:"stationId" binding:"required"`
}

type UpdateRequest struct {
	Code      string `json:"code"`
	StationID int64  `json:"stationId"`
}
