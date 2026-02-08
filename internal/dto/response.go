package dto

// StandardResponse - Standard API response format
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// Meta - Pagination metadata
type Meta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// ErrorInfo - Error details
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Count - For _count field in responses
type Count struct {
	Tasks int64 `json:"tasks"`
}

// === Team DTOs ===

type TeamResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count *Count `json:"_count,omitempty"`
}

// === JobType DTOs ===

type JobTypeResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count *Count `json:"_count,omitempty"`
}

// === JobDetail DTOs ===

type JobDetailResponse struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	JobTypeID *int64  `json:"jobTypeId"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	DeletedAt *string `json:"deletedAt"`
	Count     *Count  `json:"_count,omitempty"`
}

// === Feeder DTOs ===

type FeederResponse struct {
	ID        int64          `json:"id"`
	Code      string         `json:"code"`
	StationID int64          `json:"stationId"`
	Station   *StationNested `json:"station,omitempty"`
	Count     *Count         `json:"_count,omitempty"`
}

type StationNested struct {
	ID              int64                  `json:"id"`
	Name            string                 `json:"name"`
	CodeName        string                 `json:"codeName"`
	OperationCenter *OperationCenterNested `json:"operationCenter,omitempty"`
}

type OperationCenterNested struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// === Station DTOs ===

type StationResponse struct {
	ID              int64                  `json:"id"`
	Name            string                 `json:"name"`
	CodeName        string                 `json:"codeName"`
	OperationID     int64                  `json:"operationId"`
	OperationCenter *OperationCenterNested `json:"operationCenter,omitempty"`
}

// === PEA DTOs ===

type PEAResponse struct {
	ID              int64                  `json:"id"`
	Shortname       string                 `json:"shortname"`
	Fullname        string                 `json:"fullname"`
	OperationID     int64                  `json:"operationId"`
	OperationCenter *OperationCenterNested `json:"operationCenter,omitempty"`
}

// === OperationCenter DTOs ===

type OperationCenterResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// === Task DTOs ===

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

// === Upload DTOs ===

type UploadRequest struct {
	FileName string `json:"fileName" binding:"required"`
	FileType string `json:"fileType" binding:"required"`
}

type UploadResponse struct {
	URL          string `json:"url"`
	FileName     string `json:"fileName"`
	OriginalName string `json:"originalName"`
	Size         int64  `json:"size"`
	Type         string `json:"type"`
}

type PresignedURLResponse struct {
	UploadURL string `json:"uploadUrl"`
	FileURL   string `json:"fileUrl"`
	FileKey   string `json:"fileKey"`
}

// === Dashboard DTOs ===

type DashboardSummaryResponse struct {
	TotalTasks    int64    `json:"totalTasks"`
	TotalJobTypes int64    `json:"totalJobTypes"`
	TotalFeeders  int64    `json:"totalFeeders"`
	TopTeam       *TopTeam `json:"topTeam"`
}

type TopTeam struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type TopJobResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Count       int64  `json:"count"`
	JobTypeName string `json:"jobTypeName"`
}

type TopFeederResponse struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	StationName string `json:"stationName"`
	Count       int64  `json:"count"`
}

type FeederMatrixResponse struct {
	FeederID    int64               `json:"feederId"`
	FeederCode  string              `json:"feederCode"`
	StationName string              `json:"stationName"`
	TotalCount  int64               `json:"totalCount"`
	JobDetails  []JobDetailInMatrix `json:"jobDetails"`
}

type JobDetailInMatrix struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Count       int64  `json:"count"`
	JobTypeName string `json:"jobTypeName"`
}

type DashboardStatsResponse struct {
	Summary DashboardStatsSummary `json:"summary"`
	Charts  DashboardCharts       `json:"charts"`
}

type DashboardStatsSummary struct {
	TotalTasks  int64  `json:"totalTasks"`
	ActiveTeams int64  `json:"activeTeams"`
	TopJobType  string `json:"topJobType"`
	TopFeeder   string `json:"topFeeder"`
}

type DashboardCharts struct {
	TasksByFeeder  []ChartItem     `json:"tasksByFeeder"`
	TasksByJobType []ChartItem     `json:"tasksByJobType"`
	TasksByTeam    []ChartItem     `json:"tasksByTeam"`
	TasksByDate    []DateChartItem `json:"tasksByDate"`
}

type ChartItem struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type DateChartItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// === Tasks By Team/Filter ===

type TasksByTeamResponse struct {
	Team  TeamNested     `json:"team"`
	Tasks []TaskResponse `json:"tasks"`
}

// === Auth DTOs ===

type LoginRequest struct {
	Username string `json:"username" binding:"required,len=6,numeric"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type RefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type UserResponse struct {
	ID        uint    `json:"id"`
	Username  string  `json:"username"`
	Role      string  `json:"role"`
	TeamID    *int64  `json:"teamId,omitempty"`
	IsActive  bool    `json:"isActive"`
	LastLogin *string `json:"lastLogin,omitempty"`
	CreatedAt string  `json:"createdAt"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,len=6,numeric"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin user viewer"`
	TeamID   *int64 `json:"teamId"`
	IsActive *bool  `json:"isActive"`
}

type UpdateUserRequest struct {
	Username *string `json:"username" binding:"omitempty,len=6,numeric"`
	Role     *string `json:"role" binding:"omitempty,oneof=admin user viewer"`
	TeamID   *int64  `json:"teamId"`
	IsActive *bool   `json:"isActive"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}
