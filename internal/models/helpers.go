package models

import "gorm.io/gorm"

// CountTasksBy returns a map of id -> task count for the given column.
// Used by List() handlers for team, job_type, job_detail, feeder.
func CountTasksBy(db *gorm.DB, colName string, ids []int64) map[int64]int64 {
	countMap := make(map[int64]int64)
	if len(ids) == 0 {
		return countMap
	}

	type row struct {
		ID    int64
		Count int64
	}
	var rows []row

	db.Model(&TaskDaily{}).
		Select(colName+" as id, count(*) as count").
		Where(colName+" IN ?", ids).
		Scopes(TaskNotDeleted).
		Group(colName).
		Find(&rows)

	for _, r := range rows {
		countMap[r.ID] = r.Count
	}
	return countMap
}

// CountTasksFor returns the task count for a single id on the given column.
// Used by GetByID() and Update() handlers.
func CountTasksFor(db *gorm.DB, colName string, id int64) int64 {
	var count int64
	db.Model(&TaskDaily{}).
		Where(colName+" = ?", id).
		Scopes(TaskNotDeleted).
		Count(&count)
	return count
}
