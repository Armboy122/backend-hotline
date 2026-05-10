package models

import (
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

// TaskNotDeleted filters out soft-deleted TaskDaily records.
func TaskNotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where(TaskCol.DeletedAt + " IS NULL")
}

// JobDetailNotDeleted filters out soft-deleted JobDetail records.
func JobDetailNotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where(JobDetailCol.DeletedAt + " IS NULL")
}

// UserNotDeleted filters out soft-deleted User records.
func UserNotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where(UserCol.DeletedAt + " IS NULL")
}

// TaskByYear filters tasks by an index-friendly half-open workdate range.
func TaskByYear(year string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if year == "" {
			return db
		}
		y, err := strconv.Atoi(year)
		if err != nil || y < 1 {
			return db.Where("1 = 0")
		}
		return db.Where(TaskCol.WorkDate+" >= ? AND "+TaskCol.WorkDate+" < ?", fmt.Sprintf("%04d-01-01", y), fmt.Sprintf("%04d-01-01", y+1))
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

// PlanFileNotDeleted filters out soft-deleted PlanFile records.
func PlanFileNotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where(PlanFileCol.IsDeleted + " = false")
}

// PlanFileByPlan filters plan files by monthly plan ID.
func PlanFileByPlan(monthlyPlanID int64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(PlanFileCol.MonthlyPlanID+" = ?", monthlyPlanID)
	}
}

// PlanFileByTeam filters plan files by team ID. Skips if teamID is 0.
func PlanFileByTeam(teamID int64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if teamID == 0 {
			return db
		}
		return db.Where(PlanFileCol.TeamID+" = ?", teamID)
	}
}
