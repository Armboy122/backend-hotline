package handlers

import (
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobTypeHandler struct {
	db *gorm.DB
}

func NewJobTypeHandler(db *gorm.DB) *JobTypeHandler {
	return &JobTypeHandler{db: db}
}

func (h *JobTypeHandler) List(c *gin.Context) {
	var jobTypes []models.JobType
	if err := h.db.Find(&jobTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": jobTypes})
}

func (h *JobTypeHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var jobType models.JobType
	if err := h.db.First(&jobType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job type not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": jobType})
}

func (h *JobTypeHandler) Create(c *gin.Context) {
	var jobType models.JobType
	if err := c.ShouldBindJSON(&jobType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&jobType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": jobType})
}

func (h *JobTypeHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var jobType models.JobType
	if err := h.db.First(&jobType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job type not found"})
		return
	}
	if err := c.ShouldBindJSON(&jobType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Save(&jobType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": jobType})
}

func (h *JobTypeHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.Delete(&models.JobType{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Deleted successfully"})
}
