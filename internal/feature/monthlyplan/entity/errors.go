package entity

import "errors"

var (
	ErrPeriodNotFound   = errors.New("period not found")
	ErrInvalidPeriod    = errors.New("invalid period (year/month out of bounds)")
	ErrPeriodLocked     = errors.New("period is locked")
	ErrFileNotFound     = errors.New("file not found")
	ErrInvalidFileID    = errors.New("invalid file ID")
	ErrInvalidFileType  = errors.New("file type not allowed")
	ErrFileNotDeleted   = errors.New("file is not deleted")
	ErrForbiddenAction  = errors.New("forbidden: insufficient permissions")
	ErrSettingsLoadFail = errors.New("failed to load settings")
	ErrSettingsSaveFail = errors.New("failed to save settings")
)
