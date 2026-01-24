package handlers

import (
	"backend-hotlines3/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	var totalTasks int64
	var activeTeams int64

	h.db.Model(&models.TaskDaily{}).Count(&totalTasks)
	h.db.Model(&models.Team{}).Count(&activeTeams)

	// Top job detail
	type JobCount struct {
		JobDetailID int64
		Count       int64
	}
	var topJob JobCount
	h.db.Table("TaskDaily").
		Select("\"jobDetailId\", COUNT(*) as count").
		Group("\"jobDetailId\"").
		Order("count DESC").
		Limit(1).
		Scan(&topJob)

	var jobDetail models.JobDetail
	h.db.First(&jobDetail, topJob.JobDetailID)

	// Top feeder
	type FeederCount struct {
		FeederID int64
		Count    int64
	}
	var topFeeder FeederCount
	h.db.Table("TaskDaily").
		Select("\"feederId\", COUNT(*) as count").
		Where("\"feederId\" IS NOT NULL").
		Group("\"feederId\"").
		Order("count DESC").
		Limit(1).
		Scan(&topFeeder)

	var feeder models.Feeder
	h.db.First(&feeder, topFeeder.FeederID)

	summary := gin.H{
		"totalTasks":  totalTasks,
		"activeTeams": activeTeams,
		"topJobType":  jobDetail.Name,
		"topFeeder":   feeder.Code,
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": summary})
}

func (h *DashboardHandler) TopJobs(c *gin.Context) {
	type JobCount struct {
		JobDetailID int64  `json:"jobDetailId"`
		Count       int64  `json:"count"`
		Name        string `json:"name"`
	}
	var topJobs []JobCount
	h.db.Table("TaskDaily").
		Select("\"TaskDaily\".\"jobDetailId\", \"JobDetail\".name, COUNT(*) as count").
		Joins("JOIN \"JobDetail\" ON \"TaskDaily\".\"jobDetailId\" = \"JobDetail\".id").
		Group("\"TaskDaily\".\"jobDetailId\", \"JobDetail\".name").
		Order("count DESC").
		Limit(10).
		Scan(&topJobs)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": topJobs})
}

func (h *DashboardHandler) TopFeeders(c *gin.Context) {
	type FeederCount struct {
		FeederID int64  `json:"feederId"`
		Count    int64  `json:"count"`
		Code     string `json:"code"`
	}
	var topFeeders []FeederCount
	h.db.Table("TaskDaily").
		Select("\"TaskDaily\".\"feederId\", \"Feeder\".code, COUNT(*) as count").
		Joins("JOIN \"Feeder\" ON \"TaskDaily\".\"feederId\" = \"Feeder\".id").
		Where("\"TaskDaily\".\"feederId\" IS NOT NULL").
		Group("\"TaskDaily\".\"feederId\", \"Feeder\".code").
		Order("count DESC").
		Limit(10).
		Scan(&topFeeders)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": topFeeders})
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	// Tasks by feeder
	type FeederStat struct {
		Name  string `json:"name"`
		Value int64  `json:"value"`
	}
	var tasksByFeeder []FeederStat
	h.db.Table("TaskDaily").
		Select("\"Feeder\".code as name, COUNT(*) as value").
		Joins("JOIN \"Feeder\" ON \"TaskDaily\".\"feederId\" = \"Feeder\".id").
		Where("\"TaskDaily\".\"feederId\" IS NOT NULL").
		Group("\"Feeder\".code").
		Order("value DESC").
		Limit(10).
		Scan(&tasksByFeeder)

	// Tasks by job type
	type JobTypeStat struct {
		Name  string `json:"name"`
		Value int64  `json:"value"`
	}
	var tasksByJobType []JobTypeStat
	h.db.Table("TaskDaily").
		Select("\"JobDetail\".name, COUNT(*) as value").
		Joins("JOIN \"JobDetail\" ON \"TaskDaily\".\"jobDetailId\" = \"JobDetail\".id").
		Group("\"JobDetail\".name").
		Order("value DESC").
		Limit(10).
		Scan(&tasksByJobType)

	// Tasks by team
	type TeamStat struct {
		Name  string `json:"name"`
		Value int64  `json:"value"`
	}
	var tasksByTeam []TeamStat
	h.db.Table("TaskDaily").
		Select("\"Team\".name, COUNT(*) as value").
		Joins("JOIN \"Team\" ON \"TaskDaily\".\"teamId\" = \"Team\".id").
		Group("\"Team\".name").
		Order("value DESC").
		Scan(&tasksByTeam)

	// Tasks by date (last 30 days)
	type DateStat struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}
	var tasksByDate []DateStat
	h.db.Table("TaskDaily").
		Select("TO_CHAR(\"workDate\", 'YYYY-MM-DD') as date, COUNT(*) as count").
		Group("date").
		Order("date DESC").
		Limit(30).
		Scan(&tasksByDate)

	charts := gin.H{
		"tasksByFeeder":  tasksByFeeder,
		"tasksByJobType": tasksByJobType,
		"tasksByTeam":    tasksByTeam,
		"tasksByDate":    tasksByDate,
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": charts})
}
