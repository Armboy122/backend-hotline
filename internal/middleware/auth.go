package middleware

import (
	"backend-hotlines3/internal/dto"
	"net/http"
	"strings"

	"backend-hotlines3/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtManager *jwt.JWTManager
}

func NewAuthMiddleware(jwtManager *jwt.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "MISSING_TOKEN",
					Message: "Authorization header required",
				},
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "INVALID_FORMAT",
					Message: "Invalid authorization format. Use: Bearer <token>",
				},
			})
			c.Abort()
			return
		}

		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "INVALID_TOKEN",
					Message: "Invalid or expired token",
				},
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "UNAUTHORIZED",
					Message: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "INVALID_TOKEN",
					Message: "Invalid role claim in token",
				},
			})
			c.Abort()
			return
		}

		hasPermission := false
		for _, role := range roles {
			if roleStr == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "FORBIDDEN",
					Message: "Insufficient permissions",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.Next()
			return
		}

		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err == nil {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
		}

		c.Next()
	}
}
