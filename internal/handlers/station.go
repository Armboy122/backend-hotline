package handlers

import (
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StationHandler struct {
	db *gorm.DB
}

func NewStationHandler(db *gorm.DB) *StationHandler {
	return &StationHandler{db: db}
}

func (h *StationHandler) List(c *gin.Context) {
	var stations []models.Station
	if err := h.db.Preload("OperationCenter").Find(&stations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stations})
}

func (h *StationHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var station models.Station
	if err := h.db.Preload("OperationCenter").First(&station, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Station not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": station})
}

func (h *StationHandler) Create(c *gin.Context) {
	var station models.Station
	if err := c.ShouldBindJSON(&station); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&station).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.db.Preload("OperationCenter").First(&station, station.ID)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": station})
}

func (h *StationHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var station models.Station
	if err := h.db.First(&station, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Station not found"})
		return
	}
	if err := c.ShouldBindJSON(&station); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Save(&station).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.db.Preload("OperationCenter").First(&station, station.ID)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": station})
}

func (h *StationHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.Delete(&models.Station{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Deleted successfully"})
}
