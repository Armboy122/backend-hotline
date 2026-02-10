package v1

import (
	"log"
	"net/http"
	"strconv"

	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TeamHandler struct {
	db *gorm.DB
}

func NewTeamHandler(db *gorm.DB) *TeamHandler {
	return &TeamHandler{db: db}
}

// List retrieves all teams with their task counts.
func (h *TeamHandler) List(c *gin.Context) {
	var teams []models.Team
	if err := h.db.WithContext(c.Request.Context()).Find(&teams).Error; err != nil {
		log.Printf("Failed to fetch teams: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching teams",
			},
		})
		return
	}

	// Get task counts for each team
	var teamIDs []int64
	for _, t := range teams {
		teamIDs = append(teamIDs, t.ID)
	}
	countMap := models.CountTasksBy(h.db, models.TaskCol.TeamID, teamIDs)

	// Build response
	var response []dto.TeamResponse
	for _, t := range teams {
		response = append(response, dto.TeamResponse{
			ID:   t.ID,
			Name: t.Name,
			Count: &dto.Count{
				Tasks: countMap[t.ID],
			},
		})
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// GetByID retrieves a specific team by ID with its task count.
func (h *TeamHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid team ID",
			},
		})
		return
	}

	var team models.Team
	if err := h.db.WithContext(c.Request.Context()).First(&team, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Team not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch team %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the team",
			},
		})
		return
	}

	count := models.CountTasksFor(h.db, models.TaskCol.TeamID, id)

	response := dto.TeamResponse{
		ID:   team.ID,
		Name: team.Name,
		Count: &dto.Count{
			Tasks: count,
		},
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Create creates a new team with the provided name.
func (h *TeamHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	team := models.Team{Name: req.Name}
	if err := h.db.WithContext(c.Request.Context()).Create(&team).Error; err != nil {
		log.Printf("Failed to create team: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while creating the team",
			},
		})
		return
	}

	response := dto.TeamResponse{
		ID:   team.ID,
		Name: team.Name,
		Count: &dto.Count{
			Tasks: 0,
		},
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Update updates an existing team's name.
func (h *TeamHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid team ID",
			},
		})
		return
	}

	var team models.Team
	if err := h.db.WithContext(c.Request.Context()).First(&team, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Team not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch team %d for update: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the team",
			},
		})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	team.Name = req.Name
	if err := h.db.WithContext(c.Request.Context()).Save(&team).Error; err != nil {
		log.Printf("Failed to update team %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while updating the team",
			},
		})
		return
	}

	count := models.CountTasksFor(h.db, models.TaskCol.TeamID, id)

	response := dto.TeamResponse{
		ID:   team.ID,
		Name: team.Name,
		Count: &dto.Count{
			Tasks: count,
		},
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Delete removes a team by ID.
func (h *TeamHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid team ID",
			},
		})
		return
	}

	result := h.db.WithContext(c.Request.Context()).Delete(&models.Team{}, id)
	if result.Error != nil {
		log.Printf("Failed to delete team %d: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while deleting the team",
			},
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Team not found",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}
