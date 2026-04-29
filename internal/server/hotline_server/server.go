package hotlineserver

import (
	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/router"
	"backend-hotlines3/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// New builds the application router through the server package entrypoint.
//
// This package is the migration bridge toward internal/server/hotline_server.
func New(cfg *config.Config, db *gorm.DB, jwtManager *jwt.JWTManager) *gin.Engine {
	return router.SetupRouter(cfg, db, jwtManager)
}
