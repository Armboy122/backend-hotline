package entity

import "errors"

const CanUploadApprovedMonthlyPlan = "can_upload_approved_monthly_plan"

var (
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidCapability = errors.New("invalid capability")
)

type Capability struct {
	Code        string `json:"code"`
	Description string `json:"description"`
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
