package monthlyplan

import mpdomain "backend-hotlines3/internal/domain/monthlyplan"

var (
	ErrPeriodNotFound    = mpdomain.ErrPeriodNotFound
	ErrInvalidPeriod     = mpdomain.ErrInvalidPeriod
	ErrPeriodLocked      = mpdomain.ErrPeriodLocked
	ErrFileNotFound      = mpdomain.ErrFileNotFound
	ErrInvalidFileID     = mpdomain.ErrInvalidFileID
	ErrInvalidFileType   = mpdomain.ErrInvalidFileType
	ErrFileNotDeleted    = mpdomain.ErrFileNotDeleted
	ErrForbiddenAction   = mpdomain.ErrForbiddenAction
	ErrSettingsLoadFail  = mpdomain.ErrSettingsLoadFail
	ErrSettingsSaveFail  = mpdomain.ErrSettingsSaveFail
)
