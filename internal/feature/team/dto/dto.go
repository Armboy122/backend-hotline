package dto

import internaldto "backend-hotlines3/internal/dto"

type StandardResponse = internaldto.StandardResponse
type ErrorInfo = internaldto.ErrorInfo
type Count = internaldto.Count
type TeamResponse = internaldto.TeamResponse

type UpsertRequest struct {
	Name string `json:"name" binding:"required"`
}
