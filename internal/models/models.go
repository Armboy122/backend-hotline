// Package models defines the database models for the Hotline maintenance system.
// It includes 9 models for managing electrical network maintenance tasks.
package models

import (
	"database/sql/driver"
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// StringArray - Custom type for string array in PostgreSQL
type StringArray []string

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// PostgreSQL array format: {value1,value2}
		str := string(v)
		if str == "{}" || str == "" {
			*a = []string{}
			return nil
		}

		// Remove { and }
		if len(str) > 2 && str[0] == '{' && str[len(str)-1] == '}' {
			str = str[1 : len(str)-1]
		}

		// Split by comma
		if str == "" {
			*a = []string{}
		} else {
			*a = parsePostgresArray(str)
		}
		return nil
	case string:
		// Handle string input
		if v == "{}" || v == "" {
			*a = []string{}
			return nil
		}

		// Remove { and }
		if len(v) > 2 && v[0] == '{' && v[len(v)-1] == '}' {
			v = v[1 : len(v)-1]
		}

		if v == "" {
			*a = []string{}
		} else {
			*a = parsePostgresArray(v)
		}
		return nil
	default:
		return errors.New("failed to scan StringArray value")
	}
}

func parsePostgresArray(s string) []string {
	if s == "" {
		return []string{}
	}

	var result []string
	var current string
	inQuotes := false
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if escaped {
			current += string(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}

		if ch == ',' && !inQuotes {
			result = append(result, current)
			current = ""
			continue
		}

		current += string(ch)
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}

	// Convert to PostgreSQL array format
	var result string
	for i, v := range a {
		if i > 0 {
			result += ","
		}
		// Escape quotes and backslashes
		escaped := v
		escaped = strings.ReplaceAll(escaped, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")

		// Add quotes if contains comma or special characters
		if containsSpecialChar(escaped) {
			result += "\"" + escaped + "\""
		} else {
			result += escaped
		}
	}

	return "{" + result + "}", nil
}

func containsSpecialChar(s string) bool {
	for _, ch := range s {
		if ch == ',' || ch == '{' || ch == '}' || ch == '"' || ch == '\\' || ch == ' ' {
			return true
		}
	}
	return false
}

// OperationCenter - จุดรวมงาน
type OperationCenter struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name string `gorm:"not null;column:name" json:"name"`

	Peas     []PEA     `gorm:"foreignKey:OperationID" json:"peas,omitempty"`
	Stations []Station `gorm:"foreignKey:OperationID" json:"stations,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (OperationCenter) TableName() string {
	return "OperationCenter"
}

// PEA - การไฟฟ้า
type PEA struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Shortname   string `gorm:"not null;column:shortname" json:"shortname"`
	Fullname    string `gorm:"not null;column:fullname" json:"fullname"`
	OperationID int64  `gorm:"not null;column:operationId" json:"operationId"`

	OperationCenter *OperationCenter `gorm:"foreignKey:OperationID;references:ID" json:"operationCenter,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (PEA) TableName() string {
	return "Pea"
}

// Station - สถานี
type Station struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string `gorm:"not null;column:name" json:"name"`
	CodeName    string `gorm:"not null;unique;column:codeName" json:"codeName"`
	OperationID int64  `gorm:"not null;column:operationId" json:"operationId"`

	OperationCenter *OperationCenter `gorm:"foreignKey:OperationID;references:ID" json:"operationCenter,omitempty"`
	Feeders         []Feeder         `gorm:"foreignKey:StationID" json:"feeders,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (Station) TableName() string {
	return "Station"
}

// Feeder - ฟีดเดอร์
type Feeder struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Code      string `gorm:"not null;unique;column:code" json:"code"`
	StationID int64  `gorm:"not null;column:stationId;index:Feeder_stationId_idx" json:"stationId"`

	Station *Station    `gorm:"foreignKey:StationID;references:ID" json:"station,omitempty"`
	Tasks   []TaskDaily `gorm:"foreignKey:FeederID" json:"tasks,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (Feeder) TableName() string {
	return "Feeder"
}

// JobType - ประเภทงาน
type JobType struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name string `gorm:"not null;unique;column:name" json:"name"`

	Tasks      []TaskDaily `gorm:"foreignKey:JobTypeID" json:"tasks,omitempty"`
	JobDetails []JobDetail `gorm:"foreignKey:JobTypeID" json:"jobDetails,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (JobType) TableName() string {
	return "JobType"
}

// JobDetail - รายละเอียดงาน
type JobDetail struct {
	ID        int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string     `gorm:"not null;unique;column:name" json:"name"`
	CreatedAt time.Time  `gorm:"not null;type:timestamptz(6);column:createdAt;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"not null;type:timestamptz(6);column:updatedAt" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"type:timestamptz(6);column:deletedAt" json:"deletedAt,omitempty"`
	JobTypeID *int64     `gorm:"column:jobTypeId;index:JobDetail_jobTypeId_idx" json:"jobTypeId"`

	JobType *JobType    `gorm:"foreignKey:JobTypeID;references:ID" json:"jobType,omitempty"`
	Tasks   []TaskDaily `gorm:"foreignKey:JobDetailID" json:"tasks,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (JobDetail) TableName() string {
	return "JobDetail"
}

// Team - ทีมงาน
type Team struct {
	ID                 int64   `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name               string  `gorm:"not null;column:name" json:"name"`
	Code               *string `gorm:"column:code;uniqueIndex:team_code_unique_idx" json:"code,omitempty"`
	BaseArea           *string `gorm:"column:base_area" json:"baseArea,omitempty"`
	CrewType           *string `gorm:"column:crew_type" json:"crewType,omitempty"`
	DisplayOrder       int     `gorm:"not null;default:0;column:display_order;index:team_monthly_plan_order_idx,priority:2" json:"displayOrder"`
	MonthlyPlanVisible bool    `gorm:"not null;default:true;column:monthly_plan_visible;index:team_monthly_plan_order_idx,priority:1" json:"monthlyPlanVisible"`

	Tasks []TaskDaily `gorm:"foreignKey:TeamID" json:"tasks,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (Team) TableName() string {
	return "Team"
}

const (
	TeamPlanStatusDraft     = "draft"
	TeamPlanStatusPlanned   = "planned"
	TeamPlanStatusCancelled = "cancelled"
	TeamPlanStatusCompleted = "completed"

	LargeWorkStatusDraft      = "draft"
	LargeWorkStatusPlanned    = "planned"
	LargeWorkStatusInProgress = "in_progress"
	LargeWorkStatusCompleted  = "completed"
	LargeWorkStatusCancelled  = "cancelled"

	LargeWorkTeamRoleOwner       = "owner"
	LargeWorkTeamRoleParticipant = "participant"

	LargeWorkParticipantStatusAssigned     = "assigned"
	LargeWorkParticipantStatusAcknowledged = "acknowledged"

	LargeWorkTaskStatusTodo       = "todo"
	LargeWorkTaskStatusInProgress = "in_progress"
	LargeWorkTaskStatusDone       = "done"
	LargeWorkTaskStatusBlocked    = "blocked"
	LargeWorkTaskStatusCancelled  = "cancelled"
)

// TeamPlan - แผนงานในพื้นที่รับผิดชอบของทีม
// Additive planning table used as a source for the planning calendar.
type TeamPlan struct {
	ID                int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	TeamID            int64      `gorm:"not null;column:team_id;index:idx_team_plans_team_start_date,priority:1" json:"teamId"`
	CreatedByUserID   int64      `gorm:"not null;column:created_by_user_id;index:idx_team_plans_created_by" json:"createdByUserId"`
	Title             string     `gorm:"not null;column:title" json:"title"`
	WorkType          *string    `gorm:"column:work_type" json:"workType,omitempty"`
	StartDate         *time.Time `gorm:"type:date;column:start_date;index:idx_team_plans_team_start_date,priority:2;index:idx_team_plans_date_range,priority:1" json:"startDate,omitempty"`
	EndDate           *time.Time `gorm:"type:date;column:end_date;index:idx_team_plans_date_range,priority:2" json:"endDate,omitempty"`
	WorkTime          *string    `gorm:"column:work_time" json:"workTime,omitempty"`
	LocationText      string     `gorm:"not null;column:location_text" json:"locationText"`
	PEAID             *int64     `gorm:"column:pea_id" json:"peaId,omitempty"`
	OperationCenterID *int64     `gorm:"column:operation_center_id" json:"operationCenterId,omitempty"`
	FeederID          *int64     `gorm:"column:feeder_id" json:"feederId,omitempty"`
	StationID         *int64     `gorm:"column:station_id" json:"stationId,omitempty"`
	Notes             *string    `gorm:"column:notes" json:"notes,omitempty"`
	Status            string     `gorm:"not null;default:planned;column:status;index:idx_team_plans_status" json:"status"`
	DailyTaskID       *int64     `gorm:"column:daily_task_id" json:"dailyTaskId,omitempty"`
	CreatedAt         time.Time  `gorm:"not null;type:timestamptz(6);column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"not null;type:timestamptz(6);column:updated_at" json:"updatedAt"`
	DeletedAt         *time.Time `gorm:"type:timestamptz(6);column:deleted_at" json:"deletedAt,omitempty"`

	Team            *Team            `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
	CreatedBy       *User            `gorm:"foreignKey:CreatedByUserID;references:ID" json:"createdBy,omitempty"`
	PEA             *PEA             `gorm:"foreignKey:PEAID;references:ID" json:"pea,omitempty"`
	OperationCenter *OperationCenter `gorm:"foreignKey:OperationCenterID;references:ID" json:"operationCenter,omitempty"`
	Feeder          *Feeder          `gorm:"foreignKey:FeederID;references:ID" json:"feeder,omitempty"`
	Station         *Station         `gorm:"foreignKey:StationID;references:ID" json:"station,omitempty"`
	DailyTask       *TaskDaily       `gorm:"foreignKey:DailyTaskID;references:ID" json:"dailyTask,omitempty"`
}

func (TeamPlan) TableName() string {
	return "team_plans"
}

// LargeWorkItem - งานระดมทีมสำหรับงานขนาดใหญ่หลายทีม
type LargeWorkItem struct {
	ID                int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	OwnerTeamID       int64      `gorm:"not null;column:owner_team_id;index:idx_large_work_items_owner_start_date,priority:1" json:"ownerTeamId"`
	CreatedByUserID   int64      `gorm:"not null;column:created_by_user_id" json:"createdByUserId"`
	Title             string     `gorm:"not null;column:title" json:"title"`
	WorkType          *string    `gorm:"column:work_type" json:"workType,omitempty"`
	StartDate         time.Time  `gorm:"not null;type:date;column:start_date;index:idx_large_work_items_owner_start_date,priority:2;index:idx_large_work_items_date_range,priority:1" json:"startDate"`
	EndDate           *time.Time `gorm:"type:date;column:end_date;index:idx_large_work_items_date_range,priority:2" json:"endDate,omitempty"`
	WorkTime          *string    `gorm:"column:work_time" json:"workTime,omitempty"`
	LocationText      string     `gorm:"not null;column:location_text" json:"locationText"`
	PEAID             *int64     `gorm:"column:pea_id" json:"peaId,omitempty"`
	OperationCenterID *int64     `gorm:"column:operation_center_id" json:"operationCenterId,omitempty"`
	FeederID          *int64     `gorm:"column:feeder_id" json:"feederId,omitempty"`
	StationID         *int64     `gorm:"column:station_id" json:"stationId,omitempty"`
	Notes             *string    `gorm:"column:notes" json:"notes,omitempty"`
	Status            string     `gorm:"not null;default:planned;column:status;index:idx_large_work_items_status" json:"status"`
	CreatedAt         time.Time  `gorm:"not null;type:timestamptz(6);column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"not null;type:timestamptz(6);column:updated_at" json:"updatedAt"`
	DeletedAt         *time.Time `gorm:"type:timestamptz(6);column:deleted_at" json:"deletedAt,omitempty"`

	OwnerTeam       *Team               `gorm:"foreignKey:OwnerTeamID;references:ID" json:"ownerTeam,omitempty"`
	CreatedBy       *User               `gorm:"foreignKey:CreatedByUserID;references:ID" json:"createdBy,omitempty"`
	PEA             *PEA                `gorm:"foreignKey:PEAID;references:ID" json:"pea,omitempty"`
	OperationCenter *OperationCenter    `gorm:"foreignKey:OperationCenterID;references:ID" json:"operationCenter,omitempty"`
	Feeder          *Feeder             `gorm:"foreignKey:FeederID;references:ID" json:"feeder,omitempty"`
	Station         *Station            `gorm:"foreignKey:StationID;references:ID" json:"station,omitempty"`
	Teams           []LargeWorkItemTeam `gorm:"foreignKey:LargeWorkItemID" json:"teams,omitempty"`
}

func (LargeWorkItem) TableName() string {
	return "large_work_items"
}

// LargeWorkItemTeam stores lead/participant teams for a งานระดมทีม item.
type LargeWorkItemTeam struct {
	LargeWorkItemID      int64      `gorm:"primaryKey;not null;column:large_work_item_id" json:"largeWorkItemId"`
	TeamID               int64      `gorm:"primaryKey;not null;column:team_id;index:idx_large_work_item_teams_team" json:"teamId"`
	Role                 string     `gorm:"not null;column:role" json:"role"`
	ParticipantStatus    string     `gorm:"not null;default:assigned;column:participant_status" json:"participantStatus"`
	AcknowledgedByUserID *int64     `gorm:"column:acknowledged_by_user_id" json:"acknowledgedByUserId,omitempty"`
	AcknowledgedAt       *time.Time `gorm:"type:timestamptz(6);column:acknowledged_at" json:"acknowledgedAt,omitempty"`
	CreatedAt            time.Time  `gorm:"not null;type:timestamptz(6);column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`

	LargeWorkItem  *LargeWorkItem `gorm:"foreignKey:LargeWorkItemID;references:ID" json:"largeWorkItem,omitempty"`
	Team           *Team          `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
	AcknowledgedBy *User          `gorm:"foreignKey:AcknowledgedByUserID;references:ID" json:"acknowledgedBy,omitempty"`
}

func (LargeWorkItemTeam) TableName() string {
	return "large_work_item_teams"
}

// LargeWorkTask stores per-point execution work assigned to a team under a large work plan.
type LargeWorkTask struct {
	ID                int64            `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	LargeWorkItemID   int64            `gorm:"not null;column:large_work_item_id;uniqueIndex:idx_large_work_tasks_plan_team_sequence,priority:1" json:"largeWorkItemId"`
	AssignedTeamID    int64            `gorm:"not null;column:assigned_team_id;uniqueIndex:idx_large_work_tasks_plan_team_sequence,priority:2;index:idx_large_work_tasks_assigned_team_status,priority:1" json:"assignedTeamId"`
	Sequence          int              `gorm:"not null;column:sequence;uniqueIndex:idx_large_work_tasks_plan_team_sequence,priority:3" json:"sequence"`
	PointLabel        string           `gorm:"not null;column:point_label" json:"pointLabel"`
	Latitude          *decimal.Decimal `gorm:"type:decimal(9,6);column:latitude" json:"latitude,omitempty"`
	Longitude         *decimal.Decimal `gorm:"type:decimal(9,6);column:longitude" json:"longitude,omitempty"`
	WorkType          string           `gorm:"not null;column:work_type" json:"workType"`
	WorkDetail        *string          `gorm:"column:work_detail" json:"workDetail,omitempty"`
	PointCount        *int             `gorm:"column:point_count" json:"pointCount,omitempty"`
	TreeCount         *int             `gorm:"column:tree_count" json:"treeCount,omitempty"`
	ItemCount         *int             `gorm:"column:item_count" json:"itemCount,omitempty"`
	Notes             *string          `gorm:"column:notes" json:"notes,omitempty"`
	Status            string           `gorm:"not null;default:todo;column:status;index:idx_large_work_tasks_assigned_team_status,priority:2" json:"status"`
	BeforePhotoURLs   StringArray      `gorm:"type:text[];column:before_photo_urls" json:"beforePhotoUrls"`
	AfterPhotoURLs    StringArray      `gorm:"type:text[];column:after_photo_urls" json:"afterPhotoUrls"`
	CompletionNote    *string          `gorm:"column:completion_note" json:"completionNote,omitempty"`
	StartedByUserID   *int64           `gorm:"column:started_by_user_id" json:"startedByUserId,omitempty"`
	StartedAt         *time.Time       `gorm:"type:timestamptz(6);column:started_at" json:"startedAt,omitempty"`
	CompletedByUserID *int64           `gorm:"column:completed_by_user_id" json:"completedByUserId,omitempty"`
	CompletedAt       *time.Time       `gorm:"type:timestamptz(6);column:completed_at" json:"completedAt,omitempty"`
	Metadata          []byte           `gorm:"type:jsonb;column:metadata" json:"metadata,omitempty"`
	CreatedAt         time.Time        `gorm:"not null;type:timestamptz(6);column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt         time.Time        `gorm:"not null;type:timestamptz(6);column:updated_at" json:"updatedAt"`

	LargeWorkItem *LargeWorkItem `gorm:"foreignKey:LargeWorkItemID;references:ID" json:"largeWorkItem,omitempty"`
	AssignedTeam  *Team          `gorm:"foreignKey:AssignedTeamID;references:ID" json:"assignedTeam,omitempty"`
	StartedBy     *User          `gorm:"foreignKey:StartedByUserID;references:ID" json:"startedBy,omitempty"`
	CompletedBy   *User          `gorm:"foreignKey:CompletedByUserID;references:ID" json:"completedBy,omitempty"`
}

func (LargeWorkTask) TableName() string {
	return "large_work_tasks"
}

// TaskDaily - รายงานประจำวัน
type TaskDaily struct {
	ID              int64            `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	WorkDate        time.Time        `gorm:"not null;type:date;column:workdate;index:TaskDaily_workdate_idx" json:"workDate"`
	JobTypeID       int64            `gorm:"not null;column:jobTypeId;index:TaskDaily_jobTypeId_jobDetailId_idx" json:"jobTypeId"`
	JobDetailID     int64            `gorm:"not null;column:jobDetailId;index:TaskDaily_jobTypeId_jobDetailId_idx" json:"jobDetailId"`
	FeederID        *int64           `gorm:"column:feederId;index:TaskDaily_feederId_idx" json:"feederId"`
	NumPole         *string          `gorm:"column:numPole" json:"numPole,omitempty"`
	DeviceCode      *string          `gorm:"column:deviceCode" json:"deviceCode,omitempty"`
	URLsBefore      StringArray      `gorm:"type:text[];column:urlsBefore" json:"urlsBefore"`
	URLsAfter       StringArray      `gorm:"type:text[];column:urlsAfter" json:"urlsAfter"`
	CreatedAt       time.Time        `gorm:"not null;type:timestamptz(6);column:createdat;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt       time.Time        `gorm:"not null;type:timestamptz(6);column:updatedat" json:"updatedAt"`
	DeletedAt       *time.Time       `gorm:"type:timestamptz(6);column:deletedat" json:"deletedAt,omitempty"`
	Detail          *string          `gorm:"column:detail" json:"detail,omitempty"`
	TeamID          int64            `gorm:"not null;column:teamId" json:"teamId"`
	Latitude        *decimal.Decimal `gorm:"type:decimal(9,6);column:latitude;index:TaskDaily_latitude_longitude_idx" json:"latitude,omitempty"`
	Longitude       *decimal.Decimal `gorm:"type:decimal(9,6);column:longitude;index:TaskDaily_latitude_longitude_idx" json:"longitude,omitempty"`
	SourceType      *string          `gorm:"column:source_type;index" json:"sourceType,omitempty"`
	SourceID        *int64           `gorm:"column:source_id;index" json:"sourceId,omitempty"`
	LargeWorkTaskID *int64           `gorm:"column:large_work_task_id;uniqueIndex:taskdaily_large_work_task_unique_idx" json:"largeWorkTaskId,omitempty"`

	Team          *Team          `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
	JobType       *JobType       `gorm:"foreignKey:JobTypeID;references:ID" json:"jobType,omitempty"`
	JobDetail     *JobDetail     `gorm:"foreignKey:JobDetailID;references:ID" json:"jobDetail,omitempty"`
	Feeder        *Feeder        `gorm:"foreignKey:FeederID;references:ID" json:"feeder,omitempty"`
	LargeWorkTask *LargeWorkTask `gorm:"foreignKey:LargeWorkTaskID;references:ID" json:"largeWorkTask,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (TaskDaily) TableName() string {
	return "TaskDaily"
}

// ExternalContact stores non-user directory entries such as contractors, emergency numbers, and operation centers.
type ExternalContact struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Type         string     `gorm:"not null;column:type;index:idx_external_contacts_type" json:"type"`
	DisplayName  string     `gorm:"not null;column:display_name" json:"displayName"`
	PhoneNumber  string     `gorm:"not null;column:phone_number" json:"phoneNumber"`
	Organization *string    `gorm:"column:organization" json:"organization,omitempty"`
	Position     *string    `gorm:"column:position" json:"position,omitempty"`
	Notes        *string    `gorm:"column:notes" json:"notes,omitempty"`
	TeamID       *int64     `gorm:"column:team_id;index:idx_external_contacts_team" json:"teamId,omitempty"`
	IsActive     bool       `gorm:"not null;default:true;column:is_active" json:"isActive"`
	CreatedAt    time.Time  `gorm:"not null;type:timestamptz(6);column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"not null;type:timestamptz(6);column:updated_at" json:"updatedAt"`
	DeletedAt    *time.Time `gorm:"type:timestamptz(6);column:deleted_at" json:"deletedAt,omitempty"`

	Team *Team `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
}

func (ExternalContact) TableName() string {
	return "external_contacts"
}

// User - ผู้ใช้งานระบบ
type User struct {
	ID                 uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username           string     `gorm:"not null;unique;column:username" json:"username"`
	Password           string     `gorm:"not null;column:password" json:"-"`
	Role               string     `gorm:"not null;default:user;column:role" json:"role"`
	TeamID             *int64     `gorm:"column:teamId;index:User_teamId_idx" json:"teamId,omitempty"`
	DisplayName        *string    `gorm:"column:displayName" json:"displayName,omitempty"`
	Position           *string    `gorm:"column:position" json:"position,omitempty"`
	PhoneNumber        *string    `gorm:"column:phoneNumber" json:"phoneNumber,omitempty"`
	IsActive           bool       `gorm:"not null;default:true;column:isActive" json:"isActive"`
	MustChangePassword bool       `gorm:"not null;default:false;column:must_change_password" json:"mustChangePassword"`
	LastLogin          *time.Time `gorm:"column:lastLogin" json:"lastLogin,omitempty"`
	CreatedAt          time.Time  `gorm:"not null;type:timestamptz(6);column:createdAt;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt          time.Time  `gorm:"not null;type:timestamptz(6);column:updatedAt" json:"updatedAt"`
	DeletedAt          *time.Time `gorm:"type:timestamptz(6);column:deletedAt" json:"deletedAt,omitempty"`

	Team         *Team            `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
	Capabilities []UserCapability `gorm:"foreignKey:UserID;references:ID" json:"capabilities,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (User) TableName() string {
	return "User"
}

// UserCapability stores per-user feature capabilities granted by super_admin.
type UserCapability struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID          uint       `gorm:"not null;column:user_id;index:idx_user_capabilities_user_code,priority:1" json:"userId"`
	Code            string     `gorm:"not null;column:code;index:idx_user_capabilities_user_code,priority:2" json:"code"`
	GrantedByUserID *uint      `gorm:"column:granted_by_user_id" json:"grantedByUserId,omitempty"`
	CreatedAt       time.Time  `gorm:"not null;type:timestamptz(6);column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
	RevokedAt       *time.Time `gorm:"type:timestamptz(6);column:revoked_at;index" json:"revokedAt,omitempty"`

	User      *User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	GrantedBy *User `gorm:"foreignKey:GrantedByUserID;references:ID" json:"grantedBy,omitempty"`
}

func (UserCapability) TableName() string {
	return "user_capabilities"
}

// MonthlyPlan - แผนลงทะเบียนงานประจำเดือน
type MonthlyPlan struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Year      int       `gorm:"not null;column:year" json:"year"`
	Month     int       `gorm:"not null;column:month" json:"month"`
	IsLocked  bool      `gorm:"not null;default:false;column:isLocked" json:"isLocked"`
	CreatedAt time.Time `gorm:"not null;type:timestamptz(6);column:createdAt;default:CURRENT_TIMESTAMP" json:"createdAt"`

	Files []PlanFile `gorm:"foreignKey:MonthlyPlanID" json:"files,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (MonthlyPlan) TableName() string {
	return "MonthlyPlan"
}

const (
	MonthlyPlanScheduleStatusDraft      = "draft"
	MonthlyPlanScheduleStatusPublished  = "published"
	MonthlyPlanScheduleStatusSuperseded = "superseded"

	MonthlyPlanAssignmentTypeField   = "field"
	MonthlyPlanAssignmentTypeRemote  = "remote"
	MonthlyPlanAssignmentTypeSupport = "support"
	MonthlyPlanAssignmentTypeSpecial = "special"

	MonthlyPlanAssignmentSourceManual       = "manual"
	MonthlyPlanAssignmentSourceLargeWork    = "large_work"
	MonthlyPlanAssignmentSourceApprovedFile = "approved_file"
)

// MonthlyPlanScheduleRevision is an immutable publication unit for a monthly
// team schedule. Draft revisions may be replaced until they are published.
type MonthlyPlanScheduleRevision struct {
	ID                int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	MonthlyPlanID     int64      `gorm:"not null;column:monthly_plan_id;uniqueIndex:monthly_plan_schedule_revision_no_idx,priority:1;index:monthly_plan_schedule_status_idx,priority:1" json:"monthlyPlanId"`
	RevisionNo        int        `gorm:"not null;column:revision_no;uniqueIndex:monthly_plan_schedule_revision_no_idx,priority:2" json:"revisionNo"`
	Status            string     `gorm:"not null;column:status;index:monthly_plan_schedule_status_idx,priority:2" json:"status"`
	CreatedByUserID   int64      `gorm:"not null;column:created_by_user_id" json:"createdByUserId"`
	PublishedByUserID *int64     `gorm:"column:published_by_user_id" json:"publishedByUserId,omitempty"`
	PublishedAt       *time.Time `gorm:"type:timestamptz(6);column:published_at" json:"publishedAt,omitempty"`
	Checksum          *string    `gorm:"column:checksum" json:"checksum,omitempty"`
	Projection        []byte     `gorm:"type:jsonb;column:projection" json:"projection,omitempty"`
	CreatedAt         time.Time  `gorm:"not null;type:timestamptz(6);column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"not null;type:timestamptz(6);column:updated_at" json:"updatedAt"`

	MonthlyPlan *MonthlyPlan                `gorm:"foreignKey:MonthlyPlanID;references:ID" json:"monthlyPlan,omitempty"`
	Assignments []MonthlyPlanTeamAssignment `gorm:"foreignKey:RevisionID" json:"assignments,omitempty"`
}

func (MonthlyPlanScheduleRevision) TableName() string {
	return "monthly_plan_schedule_revisions"
}

// MonthlyPlanTeamAssignment stores one continuous away/support segment for a
// team. Home periods are deliberately derived from gaps at projection time.
type MonthlyPlanTeamAssignment struct {
	ID             int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RevisionID     int64     `gorm:"not null;column:revision_id;index:monthly_plan_assignment_team_date_idx,priority:1" json:"revisionId"`
	TeamID         int64     `gorm:"not null;column:team_id;index:monthly_plan_assignment_team_date_idx,priority:2" json:"teamId"`
	AssignmentType string    `gorm:"not null;column:assignment_type" json:"assignmentType"`
	StartDate      time.Time `gorm:"not null;type:date;column:start_date;index:monthly_plan_assignment_team_date_idx,priority:3" json:"startDate"`
	EndDate        time.Time `gorm:"not null;type:date;column:end_date" json:"endDate"`
	Destination    string    `gorm:"not null;column:destination" json:"destination"`
	Note           *string   `gorm:"column:note" json:"note,omitempty"`
	SourceType     string    `gorm:"not null;default:manual;column:source_type" json:"sourceType"`
	SourceID       *int64    `gorm:"column:source_id" json:"sourceId,omitempty"`
	CreatedAt      time.Time `gorm:"not null;type:timestamptz(6);column:created_at;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"not null;type:timestamptz(6);column:updated_at" json:"updatedAt"`

	Revision *MonthlyPlanScheduleRevision `gorm:"foreignKey:RevisionID;references:ID" json:"revision,omitempty"`
	Team     *Team                        `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
}

func (MonthlyPlanTeamAssignment) TableName() string {
	return "monthly_plan_team_assignments"
}

// PlanFile - ไฟล์แผนงานที่อัพโหลดเข้าเดือน
type PlanFile struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	MonthlyPlanID int64      `gorm:"not null;column:monthlyPlanId;index:PlanFile_monthlyPlanId_idx" json:"monthlyPlanId"`
	TeamID        *int64     `gorm:"column:teamId;index:PlanFile_teamId_idx" json:"teamId"`
	UploadedByID  int64      `gorm:"not null;column:uploadedById" json:"uploadedById"`
	FileKey       string     `gorm:"not null;column:fileKey" json:"fileKey"`
	FileURL       string     `gorm:"not null;column:fileURL" json:"fileURL"`
	FileName      string     `gorm:"not null;column:fileName" json:"fileName"`
	FileSizeBytes int64      `gorm:"not null;default:0;column:fileSizeBytes" json:"fileSizeBytes"`
	Description   *string    `gorm:"column:description" json:"description,omitempty"`
	WorkStartDate *time.Time `gorm:"type:date;column:workStartDate" json:"workStartDate,omitempty"`
	WorkEndDate   *time.Time `gorm:"type:date;column:workEndDate" json:"workEndDate,omitempty"`
	Destination   *string    `gorm:"column:destination" json:"destination,omitempty"`
	Remarks       *string    `gorm:"column:remarks" json:"remarks,omitempty"`
	IsMasterPlan  bool       `gorm:"not null;default:false;column:isMasterPlan" json:"isMasterPlan"`
	IsDeleted     bool       `gorm:"not null;default:false;column:isDeleted" json:"isDeleted"`
	DeletedAt     *time.Time `gorm:"type:timestamptz(6);column:deletedAt" json:"deletedAt,omitempty"`
	CreatedAt     time.Time  `gorm:"not null;type:timestamptz(6);column:createdAt;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"not null;type:timestamptz(6);column:updatedAt" json:"updatedAt"`

	MonthlyPlan *MonthlyPlan `gorm:"foreignKey:MonthlyPlanID;references:ID" json:"monthlyPlan,omitempty"`
	Team        *Team        `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
	UploadedBy  *User        `gorm:"foreignKey:UploadedByID;references:ID" json:"uploadedBy,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (PlanFile) TableName() string {
	return "PlanFile"
}

// FileSizeLog - log ขนาดไฟล์ทุกครั้งที่อัพโหลด (F-13)
type FileSizeLog struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PlanFileID    int64     `gorm:"not null;column:planFileId;index:FileSizeLog_planFileId_idx" json:"planFileId"`
	UploadedByID  int64     `gorm:"not null;column:uploadedById" json:"uploadedById"`
	FileSizeBytes int64     `gorm:"not null;column:fileSizeBytes" json:"fileSizeBytes"`
	CreatedAt     time.Time `gorm:"not null;type:timestamptz(6);column:createdAt;default:CURRENT_TIMESTAMP" json:"createdAt"`

	PlanFile   *PlanFile `gorm:"foreignKey:PlanFileID;references:ID" json:"planFile,omitempty"`
	UploadedBy *User     `gorm:"foreignKey:UploadedByID;references:ID" json:"uploadedBy,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (FileSizeLog) TableName() string {
	return "FileSizeLog"
}

// MonthlyPlanSetting - การตั้งค่า Monthly Plan (single-row, เก็บใน DB เพื่อใช้งานกับ Cloud Run ที่ไม่มี persistent filesystem)
type MonthlyPlanSetting struct {
	ID                      int64       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	LockDay                 int         `gorm:"not null;default:23;column:lockDay" json:"lockDay"`
	AutoCreateDay           int         `gorm:"not null;default:1;column:autoCreateDay" json:"autoCreateDay"`
	AutoCreateTarget        string      `gorm:"not null;default:'next_month';column:autoCreateTarget" json:"autoCreateTarget"`
	AllowedFileTypes        StringArray `gorm:"type:text[];column:allowedFileTypes" json:"allowedFileTypes"`
	MaxFileSizeMB           *int        `gorm:"column:maxFileSizeMB" json:"maxFileSizeMB"`
	ReminderStartDay        int         `gorm:"not null;default:20;column:reminderStartDay" json:"reminderStartDay"`
	AdminCanUploadAfterLock bool        `gorm:"not null;default:true;column:adminCanUploadAfterLock" json:"adminCanUploadAfterLock"`
	UpdatedAt               time.Time   `gorm:"not null;type:timestamptz(6);column:updatedAt" json:"updatedAt"`
}

// TableName กำหนดชื่อตารางใน database
func (MonthlyPlanSetting) TableName() string {
	return "MonthlyPlanSetting"
}
