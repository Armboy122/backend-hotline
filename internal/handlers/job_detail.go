package handlers

import (
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobDetailHandler struct {
	db *gorm.DB
}

func NewJobDetailHandler(db *gorm.DB) *JobDetailHandler {
	return &JobDetailHandler{db: db}
}

func (h *JobDetailHandler) List(c *gin.Context) {
	var jobDetails []models.JobDetail
	if err := h.db.Preload("JobType").Find(&jobDetails).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": jobDetails})
}

func (h *JobDetailHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var jobDetail models.JobDetail
	if err := h.db.Preload("JobType").First(&jobDetail, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job detail not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": jobDetail})
}

func (h *JobDetailHandler) Create(c *gin.Context) {
	var jobDetail models.JobDetail
	if err := c.ShouldBindJSON(&jobDetail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&jobDetail).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.db.Preload("JobType").First(&jobDetail, jobDetail.ID)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": jobDetail})
}

func (h *JobDetailHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var jobDetail models.JobDetail
	if err := h.db.First(&jobDetail, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job detail not found"})
		return
	}
	if err := c.ShouldBindJSON(&jobDetail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Save(&jobDetail).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.db.Preload("JobType").First(&jobDetail, jobDetail.ID)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": jobDetail})
}

func (h *JobDetailHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.Delete(&models.JobDetail{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Deleted successfully"})
}
