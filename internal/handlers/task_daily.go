package handlers

import (
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskDailyHandler struct {
	db *gorm.DB
}

func NewTaskDailyHandler(db *gorm.DB) *TaskDailyHandler {
	return &TaskDailyHandler{db: db}
}

func (h *TaskDailyHandler) List(c *gin.Context) {
	var tasks []models.TaskDaily
	query := h.db.Preload("Team.PEA").
		Preload("Feeder.Station.OperationCenter").
		Preload("JobDetail.JobType")

	// Filter by year/month/team
	if year := c.Query("year"); year != "" {
		query = query.Where("EXTRACT(YEAR FROM task_date) = ?", year)
	}
	if month := c.Query("month"); month != "" {
		query = query.Where("EXTRACT(MONTH FROM task_date) = ?", month)
	}
	if teamId := c.Query("teamId"); teamId != "" {
		query = query.Where("team_id = ?", teamId)
	}

	if err := query.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tasks})
}

func (h *TaskDailyHandler) ListByTeam(c *gin.Context) {
	var tasks []models.TaskDaily
	query := h.db.Preload("Team.PEA").
		Preload("Feeder.Station.OperationCenter").
		Preload("JobDetail.JobType")

	// Filter by year/month
	if year := c.Query("year"); year != "" {
		query = query.Where("EXTRACT(YEAR FROM task_date) = ?", year)
	}
	if month := c.Query("month"); month != "" {
		query = query.Where("EXTRACT(MONTH FROM task_date) = ?", month)
	}

	if err := query.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Group by team
	groupedTasks := make(map[string]interface{})
	for _, task := range tasks {
		if task.Team != nil {
			teamName := task.Team.Name
			if _, exists := groupedTasks[teamName]; !exists {
				groupedTasks[teamName] = gin.H{
					"team":  task.Team,
					"tasks": []models.TaskDaily{},
				}
			}
			group := groupedTasks[teamName].(gin.H)
			group["tasks"] = append(group["tasks"].([]models.TaskDaily), task)
			groupedTasks[teamName] = group
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": groupedTasks})
}

func (h *TaskDailyHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var task models.TaskDaily
	if err := h.db.Preload("Team.PEA").
		Preload("Feeder.Station.OperationCenter").
		Preload("JobDetail.JobType").
		First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": task})
}

func (h *TaskDailyHandler) Create(c *gin.Context) {
	var task models.TaskDaily
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.db.Preload("Team.PEA").
		Preload("Feeder.Station.OperationCenter").
		Preload("JobDetail.JobType").
		First(&task, task.ID)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": task})
}

func (h *TaskDailyHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var task models.TaskDaily
	if err := h.db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.db.Preload("Team.PEA").
		Preload("Feeder.Station.OperationCenter").
		Preload("JobDetail.JobType").
		First(&task, task.ID)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": task})
}

func (h *TaskDailyHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.Delete(&models.TaskDaily{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Deleted successfully"})
}
