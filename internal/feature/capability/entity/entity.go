package entity

import "errors"

const CanUploadApprovedMonthlyPlan = "can_upload_approved_monthly_plan"

var (
	ErrInvalidUserID        = errors.New("invalid user id")
	ErrInvalidCapability    = errors.New("invalid capability")
	ErrTargetUserNotFound   = errors.New("capability target user not found")
	ErrInactiveTargetUser   = errors.New("capability target user is inactive")
	ErrViewerTargetForbidden = errors.New("viewer cannot receive capabilities")
)

type Capability struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type TargetUser struct {
	ID       uint
	Role     string
	IsActive bool
}

type UserCapabilities struct {
	UserID       uint     `json:"userId"`
	Capabilities []string `json:"capabilities"`
}

func AvailableCapabilities() []Capability {
	return []Capability{{
		Code:        CanUploadApprovedMonthlyPlan,
		Description: "อัปโหลด/แทนที่ไฟล์ approved/master monthly plan",
	}}
}

func IsValidCapability(code string) bool {
	return code == CanUploadApprovedMonthlyPlan
}
