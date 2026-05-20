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
	ID                 uint        `json:"id"`
	Username           string      `json:"username"`
	Role               string      `json:"role"`
	TeamID             *int64      `json:"teamId,omitempty"`
	Team               *TeamNested `json:"team,omitempty"`
	Capabilities       []string    `json:"capabilities,omitempty"`
	DisplayName        *string     `json:"displayName,omitempty"`
	Position           *string     `json:"position,omitempty"`
	PhoneNumber        *string     `json:"phoneNumber,omitempty"`
	IsActive           bool        `json:"isActive"`
	MustChangePassword bool        `json:"mustChangePassword"`
	LastLogin          *string     `json:"lastLogin,omitempty"`
	CreatedAt          string      `json:"createdAt"`
}

type ContactDirectoryActions struct {
	CanEdit           bool `json:"canEdit"`
	CanEditRoleOrTeam bool `json:"canEditRoleOrTeam"`
}

type ContactDirectoryResponse struct {
	ID          uint                    `json:"id"`
	Username    string                  `json:"username"`
	DisplayName *string                 `json:"displayName"`
	Position    *string                 `json:"position"`
	PhoneNumber *string                 `json:"phoneNumber"`
	Role        string                  `json:"role"`
	TeamID      *int64                  `json:"teamId"`
	Team        *TeamNested             `json:"team,omitempty"`
	IsActive    bool                    `json:"isActive"`
	Actions     ContactDirectoryActions `json:"actions"`
	UpdatedAt   string                  `json:"updatedAt"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,len=6,numeric"`
	Password string `json:"password" binding:"omitempty,min=6"`
	Role     string `json:"role" binding:"required,oneof=super_admin team_lead user viewer"`
	TeamID   *int64 `json:"teamId"`
	IsActive *bool  `json:"isActive"`
}

type UpdateUserRequest struct {
	Username *string `json:"username" binding:"omitempty,len=6,numeric"`
	Role     *string `json:"role" binding:"omitempty,oneof=super_admin team_lead user viewer"`
	TeamID   *int64  `json:"teamId"`
	IsActive *bool   `json:"isActive"`
}

type UpdateContactRequest struct {
	DisplayName *string `json:"displayName" binding:"omitempty,max=120"`
	Position    *string `json:"position" binding:"omitempty,max=120"`
	PhoneNumber *string `json:"phoneNumber" binding:"omitempty,max=40"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"omitempty,min=6"`
}

// === Monthly Plan DTOs ===

// MonthlyPlanResponse — period ประจำเดือน
type MonthlyPlanResponse struct {
	ID        int64  `json:"id"`
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	IsLocked  bool   `json:"isLocked"`
	CreatedAt string `json:"createdAt"`
}

// PlanFileUploaderNested — compact uploader info embedded in file responses
type PlanFileUploaderNested struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// PlanFileResponse — single file entry
type PlanFileResponse struct {
	ID            int64                   `json:"id"`
	MonthlyPlanID int64                   `json:"monthlyPlanId"`
	TeamID        *int64                  `json:"teamId"`
	UploadedByID  int64                   `json:"uploadedById"`
	FileKey       string                  `json:"fileKey"`
	FileURL       string                  `json:"fileURL"`
	FileName      string                  `json:"fileName"`
	FileSizeBytes int64                   `json:"fileSizeBytes"`
	Description   *string                 `json:"description"`
	WorkStartDate *string                 `json:"workStartDate"`
	WorkEndDate   *string                 `json:"workEndDate"`
	Destination   *string                 `json:"destination"`
	Remarks       *string                 `json:"remarks"`
	IsMasterPlan  bool                    `json:"isMasterPlan"`
	IsDeleted     bool                    `json:"isDeleted"`
	DeletedAt     *string                 `json:"deletedAt"`
	CreatedAt     string                  `json:"createdAt"`
	UpdatedAt     string                  `json:"updatedAt"`
	Team          *TeamNested             `json:"team,omitempty"`
	UploadedBy    *PlanFileUploaderNested `json:"uploadedBy,omitempty"`
}

// PresignPlanFileRequest — ขอ presigned PUT URL สำหรับ PDF upload
type PresignPlanFileRequest struct {
	FileName string `json:"fileName" binding:"required"`
	FileType string `json:"fileType" binding:"required"`
}

// ConfirmPlanFileRequest — ยืนยันหลังอัพโหลดเสร็จ + save metadata
type ConfirmPlanFileRequest struct {
	FileKey       string  `json:"fileKey" binding:"required"`
	FileURL       string  `json:"fileURL" binding:"required"`
	FileName      string  `json:"fileName" binding:"required"`
	FileSizeBytes int64   `json:"fileSizeBytes" binding:"required,min=1"`
	Description   *string `json:"description"`
	WorkStartDate *string `json:"workStartDate"`
	WorkEndDate   *string `json:"workEndDate"`
	Destination   *string `json:"destination"`
	Remarks       *string `json:"remarks"`
	IsMasterPlan  bool    `json:"isMasterPlan"`
	// TeamID — optional, admin เท่านั้นที่ระบุได้ เพื่ออัพโหลดแทนทีมอื่น
	// ถ้าไม่ส่งมา → ใช้ teamId ของ user ที่ login
	// ถ้า isMasterPlan = true → teamId จะถูก ignore
	TeamID *int64 `json:"teamId"`
}

type MonthlyPlanConversionRequest struct {
	Year            int     `json:"year" binding:"required"`
	Month           int     `json:"month" binding:"required"`
	ApprovedFileID  int64   `json:"approvedFileId" binding:"required,min=1"`
	SelectedTeamIDs []int64 `json:"selectedTeamIds" binding:"required,min=1"`
}

type MonthlyPlanConversionResponse struct {
	PlanningItemsCreated int    `json:"planningItemsCreated"`
	SourceFileID         int64  `json:"sourceFileId"`
	ConvertedAt          string `json:"convertedAt"`
}

// TeamSubmissionStatus — สถานะการส่งของแต่ละทีม
type TeamSubmissionStatus struct {
	Team      TeamNested `json:"team"`
	Status    string     `json:"status"`    // "submitted" | "pending" | "missed"
	FileCount int        `json:"fileCount"` // จำนวนไฟล์ที่อัพโหลด
}

// SubmissionStatusResponse — overview สถานะทุกทีม
type SubmissionStatusResponse struct {
	Period   MonthlyPlanResponse    `json:"period"`
	Deadline string                 `json:"deadline"` // YYYY-MM-DD
	Teams    []TeamSubmissionStatus `json:"teams"`
}

// MonthlyPlanActionResponse — actor-specific available actions for one month.
type MonthlyPlanActionResponse struct {
	CanUpload bool `json:"canUpload"`
}

// MonthlyPlanOverviewMonthResponse — yearly overview entry for one month.
type MonthlyPlanOverviewMonthResponse struct {
	Period   MonthlyPlanResponse       `json:"period"`
	Month    int                       `json:"month"`
	Deadline string                    `json:"deadline"`
	IsLocked bool                      `json:"isLocked"`
	Status   string                    `json:"status"`
	Actions  MonthlyPlanActionResponse `json:"actions"`
	Files    []PlanFileResponse        `json:"files"`
}

// MonthlyPlanYearOverviewResponse — all 12 monthly plan buckets for a year.
type MonthlyPlanYearOverviewResponse struct {
	Year   int                                `json:"year"`
	Months []MonthlyPlanOverviewMonthResponse `json:"months"`
}

// MonthlyPlanSettingsResponse — response สำหรับ round-1 Admin settings
// Round 1 exposes only Monthly Plan lock day and whether privileged upload after lock is allowed.
type MonthlyPlanSettingsResponse struct {
	LockDay                 int  `json:"lockDay"`
	AdminCanUploadAfterLock bool `json:"adminCanUploadAfterLock"`
}

// UpdateMonthlyPlanSettingsRequest — request สำหรับ round-1 Admin settings
type UpdateMonthlyPlanSettingsRequest struct {
	LockDay                 *int  `json:"lockDay"`
	AdminCanUploadAfterLock *bool `json:"adminCanUploadAfterLock"`
}
