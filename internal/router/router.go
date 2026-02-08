package router

import (
	"backend-hotlines3/internal/config"
	v1 "backend-hotlines3/internal/handlers/v1"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(CORSMiddleware(cfg))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// ============================================
	// API v1 Routes (New Standard Format)
	// ============================================
	apiV1 := r.Group("/v1")
	{
		// Teams
		teamsV1 := apiV1.Group("/teams")
		{
			handler := v1.NewTeamHandler(db)
			teamsV1.GET("", handler.List)
			teamsV1.GET("/:id", handler.GetByID)
			teamsV1.POST("", handler.Create)
			teamsV1.PUT("/:id", handler.Update)
			teamsV1.DELETE("/:id", handler.Delete)
		}

		// Job Types
		jobTypesV1 := apiV1.Group("/job-types")
		{
			handler := v1.NewJobTypeHandler(db)
			jobTypesV1.GET("", handler.List)
			jobTypesV1.GET("/:id", handler.GetByID)
			jobTypesV1.POST("", handler.Create)
			jobTypesV1.PUT("/:id", handler.Update)
			jobTypesV1.DELETE("/:id", handler.Delete)
		}

		// Job Details
		jobDetailsV1 := apiV1.Group("/job-details")
		{
			handler := v1.NewJobDetailHandler(db)
			jobDetailsV1.GET("", handler.List)
			jobDetailsV1.GET("/:id", handler.GetByID)
			jobDetailsV1.POST("", handler.Create)
			jobDetailsV1.PUT("/:id", handler.Update)
			jobDetailsV1.DELETE("/:id", handler.Delete)
			jobDetailsV1.POST("/:id/restore", handler.Restore)
		}

		// Feeders
		feedersV1 := apiV1.Group("/feeders")
		{
			handler := v1.NewFeederHandler(db)
			feedersV1.GET("", handler.List)
			feedersV1.GET("/:id", handler.GetByID)
			feedersV1.POST("", handler.Create)
			feedersV1.PUT("/:id", handler.Update)
			feedersV1.DELETE("/:id", handler.Delete)
		}

		// Stations
		stationsV1 := apiV1.Group("/stations")
		{
			handler := v1.NewStationHandler(db)
			stationsV1.GET("", handler.List)
			stationsV1.GET("/:id", handler.GetByID)
			stationsV1.POST("", handler.Create)
			stationsV1.PUT("/:id", handler.Update)
			stationsV1.DELETE("/:id", handler.Delete)
		}

		// PEAs
		peasV1 := apiV1.Group("/peas")
		{
			handler := v1.NewPEAHandler(db)
			peasV1.GET("", handler.List)
			peasV1.GET("/:id", handler.GetByID)
			peasV1.POST("", handler.Create)
			peasV1.POST("/bulk", handler.BulkCreate)
			peasV1.PUT("/:id", handler.Update)
			peasV1.DELETE("/:id", handler.Delete)
		}

		// Operation Centers
		operationCentersV1 := apiV1.Group("/operation-centers")
		{
			handler := v1.NewOperationCenterHandler(db)
			operationCentersV1.GET("", handler.List)
			operationCentersV1.GET("/:id", handler.GetByID)
			operationCentersV1.POST("", handler.Create)
			operationCentersV1.PUT("/:id", handler.Update)
			operationCentersV1.DELETE("/:id", handler.Delete)
		}

		// Tasks
		tasksV1 := apiV1.Group("/tasks")
		{
			handler := v1.NewTaskHandler(db)
			tasksV1.GET("", handler.List)
			tasksV1.GET("/by-team", handler.ListByTeam)
			tasksV1.GET("/by-filter", handler.ListByFilter)
			tasksV1.GET("/:id", handler.GetByID)
			tasksV1.POST("", handler.Create)
			tasksV1.PUT("/:id", handler.Update)
			tasksV1.DELETE("/:id", handler.Delete)
		}

		// Upload
		uploadHandler, err := v1.NewUploadHandler(cfg)
		if err != nil {
			log.Printf("Warning: Upload handler initialization failed: %v", err)
		} else {
			uploadV1 := apiV1.Group("/upload")
			{
				uploadV1.POST("/image", uploadHandler.GetPresignedURL)
				uploadV1.DELETE("/*key", uploadHandler.DeleteFile)
			}
		}

		// Dashboard
		dashboardV1 := apiV1.Group("/dashboard")
		{
			handler := v1.NewDashboardHandler(db)
			dashboardV1.GET("/summary", handler.Summary)
			dashboardV1.GET("/top-jobs", handler.TopJobs)
			dashboardV1.GET("/top-feeders", handler.TopFeeders)
			dashboardV1.GET("/feeder-matrix", handler.FeederMatrix)
			dashboardV1.GET("/stats", handler.Stats)
		}
	}

	return r
}

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
