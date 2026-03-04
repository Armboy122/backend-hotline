package middleware

import (
	"net/http"
	"sync"
	"time"

	"backend-hotlines3/internal/dto"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory rate limiter using token bucket algorithm
type RateLimiter struct {
	requests map[string]*clientInfo
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type clientInfo struct {
	count     int
	resetTime time.Time
}

// NewRateLimiter creates a new rate limiter with specified requests per window
func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*clientInfo),
		limit:    requestsPerWindow,
		window:   window,
	}
}

// Middleware returns the gin middleware function
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		rl.mu.Lock()
		defer rl.mu.Unlock()
		
		now := time.Now()
		info, exists := rl.requests[clientIP]
		
		// Reset if window has passed
		if !exists || now.After(info.resetTime) {
			rl.requests[clientIP] = &clientInfo{
				count:     1,
				resetTime: now.Add(rl.window),
			}
			c.Next()
			return
		}
		
		// Check limit
		if info.count >= rl.limit {
			c.JSON(http.StatusTooManyRequests, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Too many requests. Please try again later.",
				},
			})
			c.Abort()
			return
		}
		
		// Increment count
		info.count++
		c.Next()
	}
}

// Cleanup removes expired entries (should be called periodically)
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	for ip, info := range rl.requests {
		if now.After(info.resetTime) {
			delete(rl.requests, ip)
		}
	}
}
