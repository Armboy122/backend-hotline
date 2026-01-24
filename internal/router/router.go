package router

import (
	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/handlers"

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
			"status": "ok",
			"message": "Server is running",
		})
	})

	// API routes
	api := r.Group("/api")
	{
		// Master Data Routes
		operationCenters := api.Group("/operation-centers")
		{
			handler := handlers.NewOperationCenterHandler(db)
			operationCenters.GET("", handler.List)
			operationCenters.GET("/:id", handler.GetByID)
			operationCenters.POST("", handler.Create)
			operationCenters.PUT("/:id", handler.Update)
			operationCenters.DELETE("/:id", handler.Delete)
		}

		peas := api.Group("/peas")
		{
			handler := handlers.NewPEAHandler(db)
			peas.GET("", handler.List)
			peas.GET("/:id", handler.GetByID)
			peas.POST("", handler.Create)
			peas.POST("/bulk", handler.BulkCreate)
			peas.PUT("/:id", handler.Update)
			peas.DELETE("/:id", handler.Delete)
		}

		stations := api.Group("/stations")
		{
			handler := handlers.NewStationHandler(db)
			stations.GET("", handler.List)
			stations.GET("/:id", handler.GetByID)
			stations.POST("", handler.Create)
			stations.PUT("/:id", handler.Update)
			stations.DELETE("/:id", handler.Delete)
		}

		feeders := api.Group("/feeders")
		{
			handler := handlers.NewFeederHandler(db)
			feeders.GET("", handler.List)
			feeders.GET("/:id", handler.GetByID)
			feeders.POST("", handler.Create)
			feeders.PUT("/:id", handler.Update)
			feeders.DELETE("/:id", handler.Delete)
		}

		jobTypes := api.Group("/job-types")
		{
			handler := handlers.NewJobTypeHandler(db)
			jobTypes.GET("", handler.List)
			jobTypes.GET("/:id", handler.GetByID)
			jobTypes.POST("", handler.Create)
			jobTypes.PUT("/:id", handler.Update)
			jobTypes.DELETE("/:id", handler.Delete)
		}

		jobDetails := api.Group("/job-details")
		{
			handler := handlers.NewJobDetailHandler(db)
			jobDetails.GET("", handler.List)
			jobDetails.GET("/:id", handler.GetByID)
			jobDetails.POST("", handler.Create)
			jobDetails.PUT("/:id", handler.Update)
			jobDetails.DELETE("/:id", handler.Delete)
		}

		teams := api.Group("/teams")
		{
			handler := handlers.NewTeamHandler(db)
			teams.GET("", handler.List)
			teams.GET("/:id", handler.GetByID)
			teams.POST("", handler.Create)
			teams.PUT("/:id", handler.Update)
			teams.DELETE("/:id", handler.Delete)
		}

		// Task Daily Routes
		tasks := api.Group("/tasks")
		{
			handler := handlers.NewTaskDailyHandler(db)
			tasks.GET("", handler.List)
			tasks.GET("/by-team", handler.ListByTeam)
			tasks.GET("/:id", handler.GetByID)
			tasks.POST("", handler.Create)
			tasks.PUT("/:id", handler.Update)
			tasks.DELETE("/:id", handler.Delete)
		}

		// Dashboard Routes
		dashboard := api.Group("/dashboard")
		{
			handler := handlers.NewDashboardHandler(db)
			dashboard.GET("/summary", handler.Summary)
			dashboard.GET("/top-jobs", handler.TopJobs)
			dashboard.GET("/top-feeders", handler.TopFeeders)
			dashboard.GET("/stats", handler.Stats)
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
