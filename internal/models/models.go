package models

import (
	"database/sql/driver"
	"errors"
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
		escaped = replaceAll(escaped, "\\", "\\\\")
		escaped = replaceAll(escaped, "\"", "\\\"")

		// Add quotes if contains comma or special characters
		if containsSpecialChar(escaped) {
			result += "\"" + escaped + "\""
		} else {
			result += escaped
		}
	}

	return "{" + result + "}", nil
}

func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := indexOfString(s, old)
		if idx == -1 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
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
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name string `gorm:"not null;column:name" json:"name"`

	Tasks []TaskDaily `gorm:"foreignKey:TeamID" json:"tasks,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (Team) TableName() string {
	return "Team"
}

// TaskDaily - รายงานประจำวัน
type TaskDaily struct {
	ID          int64            `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	WorkDate    time.Time        `gorm:"not null;type:date;column:workdate;index:TaskDaily_workdate_idx" json:"workDate"`
	JobTypeID   int64            `gorm:"not null;column:jobTypeId;index:TaskDaily_jobTypeId_jobDetailId_idx" json:"jobTypeId"`
	JobDetailID int64            `gorm:"not null;column:jobDetailId;index:TaskDaily_jobTypeId_jobDetailId_idx" json:"jobDetailId"`
	FeederID    *int64           `gorm:"column:feederId;index:TaskDaily_feederId_idx" json:"feederId"`
	NumPole     *string          `gorm:"column:numPole" json:"numPole,omitempty"`
	DeviceCode  *string          `gorm:"column:deviceCode" json:"deviceCode,omitempty"`
	URLsBefore  StringArray      `gorm:"type:text[];column:urlsBefore" json:"urlsBefore"`
	URLsAfter   StringArray      `gorm:"type:text[];column:urlsAfter" json:"urlsAfter"`
	CreatedAt   time.Time        `gorm:"not null;type:timestamptz(6);column:createdat;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time        `gorm:"not null;type:timestamptz(6);column:updatedat" json:"updatedAt"`
	DeletedAt   *time.Time       `gorm:"type:timestamptz(6);column:deletedat" json:"deletedAt,omitempty"`
	Detail      *string          `gorm:"column:detail" json:"detail,omitempty"`
	TeamID      int64            `gorm:"not null;column:teamId" json:"teamId"`
	Latitude    *decimal.Decimal `gorm:"type:decimal(9,6);column:latitude;index:TaskDaily_latitude_longitude_idx" json:"latitude,omitempty"`
	Longitude   *decimal.Decimal `gorm:"type:decimal(9,6);column:longitude;index:TaskDaily_latitude_longitude_idx" json:"longitude,omitempty"`

	Team      *Team      `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
	JobType   *JobType   `gorm:"foreignKey:JobTypeID;references:ID" json:"jobType,omitempty"`
	JobDetail *JobDetail `gorm:"foreignKey:JobDetailID;references:ID" json:"jobDetail,omitempty"`
	Feeder    *Feeder    `gorm:"foreignKey:FeederID;references:ID" json:"feeder,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (TaskDaily) TableName() string {
	return "TaskDaily"
}

// User - ผู้ใช้งานระบบ
type User struct {
	ID        uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username  string     `gorm:"not null;unique;column:username" json:"username"`
	Password  string     `gorm:"not null;column:password" json:"-"`
	Role      string     `gorm:"not null;default:user;column:role" json:"role"`
	TeamID    *int64     `gorm:"column:teamId;index:User_teamId_idx" json:"teamId,omitempty"`
	IsActive  bool       `gorm:"not null;default:true;column:isActive" json:"isActive"`
	LastLogin *time.Time `gorm:"column:lastLogin" json:"lastLogin,omitempty"`
	CreatedAt time.Time  `gorm:"not null;type:timestamptz(6);column:createdAt;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"not null;type:timestamptz(6);column:updatedAt" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"type:timestamptz(6);column:deletedAt" json:"deletedAt,omitempty"`

	Team *Team `gorm:"foreignKey:TeamID;references:ID" json:"team,omitempty"`
}

// TableName กำหนดชื่อตารางใน database
func (User) TableName() string {
	return "User"
}
