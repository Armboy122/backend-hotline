package handlers

import (
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PEAHandler struct {
	db *gorm.DB
}

func NewPEAHandler(db *gorm.DB) *PEAHandler {
	return &PEAHandler{db: db}
}

func (h *PEAHandler) List(c *gin.Context) {
	var peas []models.PEA
	if err := h.db.Find(&peas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": peas})
}

func (h *PEAHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var pea models.PEA
	if err := h.db.First(&pea, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PEA not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": pea})
}

func (h *PEAHandler) Create(c *gin.Context) {
	var pea models.PEA
	if err := c.ShouldBindJSON(&pea); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&pea).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": pea})
}

func (h *PEAHandler) BulkCreate(c *gin.Context) {
	var peas []models.PEA
	if err := c.ShouldBindJSON(&peas); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Create(&peas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": peas})
}

func (h *PEAHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var pea models.PEA
	if err := h.db.First(&pea, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PEA not found"})
		return
	}
	if err := c.ShouldBindJSON(&pea); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Save(&pea).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": pea})
}

func (h *PEAHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.Delete(&models.PEA{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Deleted successfully"})
}
