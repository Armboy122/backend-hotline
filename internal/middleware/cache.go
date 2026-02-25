package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// CachePublic sets Cache-Control header to allow CDN (e.g. Cloudflare) to cache the response.
// seconds is the max-age in seconds.
func CachePublic(seconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", seconds))
		c.Next()
	}
}

// CachePrivate sets Cache-Control header to prevent CDN/proxy caching.
// Use this for user-specific or sensitive endpoints.
func CachePrivate() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-store")
		c.Next()
	}
}
