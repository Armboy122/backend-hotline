package handlers

import (
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OperationCenterHandler struct {
	db *gorm.DB
}

func NewOperationCenterHandler(db *gorm.DB) *OperationCenterHandler {
	return &OperationCenterHandler{db: db}
}

func (h *OperationCenterHandler) List(c *gin.Context) {
	var centers []models.OperationCenter
	if err := h.db.Find(&centers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": centers})
}

func (h *OperationCenterHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var center models.OperationCenter
	if err := h.db.First(&center, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Operation center not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": center})
}

func (h *OperationCenterHandler) Create(c *gin.Context) {
	var center models.OperationCenter
	if err := c.ShouldBindJSON(&center); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&center).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": center})
}

func (h *OperationCenterHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var center models.OperationCenter
	if err := h.db.First(&center, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Operation center not found"})
		return
	}
	if err := c.ShouldBindJSON(&center); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Save(&center).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": center})
}

func (h *OperationCenterHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.Delete(&models.OperationCenter{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Deleted successfully"})
}
