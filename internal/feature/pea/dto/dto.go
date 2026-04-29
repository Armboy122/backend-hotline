package dto

import internaldto "backend-hotlines3/internal/dto"

type StandardResponse = internaldto.StandardResponse
type ErrorInfo = internaldto.ErrorInfo
type PEAResponse = internaldto.PEAResponse
type OperationCenterNested = internaldto.OperationCenterNested

type CreateRequest struct {
	Shortname   string `json:"shortname" binding:"required"`
	Fullname    string `json:"fullname" binding:"required"`
	OperationID int64  `json:"operationId" binding:"required"`
}

type UpdateRequest struct {
	Shortname   string `json:"shortname"`
	Fullname    string `json:"fullname"`
	OperationID int64  `json:"operationId"`
}
