package dto

import internaldto "backend-hotlines3/internal/dto"

type StandardResponse = internaldto.StandardResponse
type ErrorInfo = internaldto.ErrorInfo
type Meta = internaldto.Meta

// Monthly Plan DTOs
type MonthlyPlanResponse = internaldto.MonthlyPlanResponse
type PlanFileResponse = internaldto.PlanFileResponse
type PlanFileUploaderNested = internaldto.PlanFileUploaderNested
type PresignPlanFileRequest = internaldto.PresignPlanFileRequest
type ConfirmPlanFileRequest = internaldto.ConfirmPlanFileRequest
type PresignedURLResponse = internaldto.PresignedURLResponse
type TeamSubmissionStatus = internaldto.TeamSubmissionStatus
type SubmissionStatusResponse = internaldto.SubmissionStatusResponse
type MonthlyPlanActionResponse = internaldto.MonthlyPlanActionResponse
type MonthlyPlanOverviewMonthResponse = internaldto.MonthlyPlanOverviewMonthResponse
type MonthlyPlanYearOverviewResponse = internaldto.MonthlyPlanYearOverviewResponse
type MonthlyPlanSettingsResponse = internaldto.MonthlyPlanSettingsResponse
type UpdateMonthlyPlanSettingsRequest = internaldto.UpdateMonthlyPlanSettingsRequest
type TeamNested = internaldto.TeamNested
