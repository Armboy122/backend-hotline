package middleware

import (
	"context"
	"net/http"
	"time"

	"backend-hotlines3/internal/dto"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware creates a middleware that sets a timeout for each request
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context with timeout context
		c.Request = c.Request.WithContext(ctx)

		// Use a channel to detect if handler completed
		done := make(chan struct{})

		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// Handler completed normally
			return
		case <-ctx.Done():
			// Timeout occurred
			c.JSON(http.StatusGatewayTimeout, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "REQUEST_TIMEOUT",
					Message: "Request timeout - please try again later",
				},
			})
			c.Abort()
		}
	}
}
