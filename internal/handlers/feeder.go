package handlers

import (
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FeederHandler struct {
	db *gorm.DB
}

func NewFeederHandler(db *gorm.DB) *FeederHandler {
	return &FeederHandler{db: db}
}

func (h *FeederHandler) List(c *gin.Context) {
	var feeders []models.Feeder
	if err := h.db.Preload("Station.OperationCenter").Find(&feeders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": feeders})
}

func (h *FeederHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var feeder models.Feeder
	if err := h.db.Preload("Station.OperationCenter").First(&feeder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feeder not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": feeder})
}

func (h *FeederHandler) Create(c *gin.Context) {
	var feeder models.Feeder
	if err := c.ShouldBindJSON(&feeder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&feeder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.db.Preload("Station.OperationCenter").First(&feeder, feeder.ID)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": feeder})
}

func (h *FeederHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var feeder models.Feeder
	if err := h.db.First(&feeder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feeder not found"})
		return
	}
	if err := c.ShouldBindJSON(&feeder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Save(&feeder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.db.Preload("Station.OperationCenter").First(&feeder, feeder.ID)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": feeder})
}

func (h *FeederHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.Delete(&models.Feeder{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Deleted successfully"})
}
