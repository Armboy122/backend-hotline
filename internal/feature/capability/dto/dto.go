package dto

import internaldto "backend-hotlines3/internal/dto"

type StandardResponse = internaldto.StandardResponse
type ErrorInfo = internaldto.ErrorInfo

type ReplaceCapabilitiesRequest struct {
	Capabilities []string `json:"capabilities"`
}
