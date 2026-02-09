package models

import "gorm.io/gorm"

// TaskNotDeleted filters out soft-deleted TaskDaily records.
func TaskNotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where(TaskCol.DeletedAt + " IS NULL")
}

// JobDetailNotDeleted filters out soft-deleted JobDetail records.
func JobDetailNotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where(JobDetailCol.DeletedAt + " IS NULL")
}

// TaskByYear filters tasks by year extracted from workdate.
func TaskByYear(year string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if year == "" {
			return db
		}
		return db.Where("EXTRACT(YEAR FROM "+TaskCol.WorkDate+") = ?", year)
	}
}

// TaskByMonth filters tasks by month extracted from workdate.
func TaskByMonth(month string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if month == "" {
			return db
		}
		return db.Where("EXTRACT(MONTH FROM "+TaskCol.WorkDate+") = ?", month)
	}
}

// TaskByTeam filters tasks by teamId. Skips if empty or "all".
func TaskByTeam(teamID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if teamID == "" || teamID == "all" {
			return db
		}
		return db.Where(TaskCol.TeamID+" = ?", teamID)
	}
}

// TaskByJobType filters tasks by jobTypeId. Skips if empty or "all".
func TaskByJobType(jobTypeID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if jobTypeID == "" || jobTypeID == "all" {
			return db
		}
		return db.Where(TaskCol.JobTypeID+" = ?", jobTypeID)
	}
}

// TaskByFeeder filters tasks by feederId. Skips if empty or "all".
func TaskByFeeder(feederID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if feederID == "" || feederID == "all" {
			return db
		}
		return db.Where(TaskCol.FeederID+" = ?", feederID)
	}
}

// TaskByDateRange filters tasks between startDate and endDate.
func TaskByDateRange(startDate, endDate string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if startDate != "" {
			db = db.Where(TaskCol.WorkDate+" >= ?", startDate)
		}
		if endDate != "" {
			db = db.Where(TaskCol.WorkDate+" <= ?", endDate)
		}
		return db
	}
}

// TaskFeederNotNull filters tasks where feederId is not null.
func TaskFeederNotNull(db *gorm.DB) *gorm.DB {
	return db.Where(TaskCol.FeederID + " IS NOT NULL")
}

// ApplyDashboardFilters applies year, month, team, and jobType filters together.
func ApplyDashboardFilters(year, month, teamID, jobTypeID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Scopes(TaskByYear(year)).
			Scopes(TaskByMonth(month)).
			Scopes(TaskByTeam(teamID)).
			Scopes(TaskByJobType(jobTypeID))
	}
}
