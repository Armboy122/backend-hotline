package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type publicCacheEntry struct {
	status    int
	header    http.Header
	body      []byte
	expiresAt time.Time
}

var publicCache = struct {
	sync.RWMutex
	entries map[string]publicCacheEntry
}{entries: make(map[string]publicCacheEntry)}

// CachePublic sets Cache-Control header to allow CDN (e.g. Cloudflare) and the
// process-local cache to cache successful GET responses for the given max-age.
func CachePublic(seconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		cacheControl := fmt.Sprintf("public, max-age=%d", seconds)
		c.Header("Cache-Control", cacheControl)

		if seconds <= 0 || c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		key := c.Request.URL.RequestURI()
		now := time.Now()
		publicCache.RLock()
		entry, ok := publicCache.entries[key]
		publicCache.RUnlock()
		if ok && now.Before(entry.expiresAt) {
			for name, values := range entry.header {
				for _, value := range values {
					c.Writer.Header().Add(name, value)
				}
			}
			c.Header("Cache-Control", cacheControl)
			c.Data(entry.status, entry.header.Get("Content-Type"), entry.body)
			c.Abort()
			return
		}

		recorder := &cacheResponseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = recorder
		c.Next()

		status := recorder.Status()
		if status == http.StatusOK && !c.IsAborted() {
			header := recorder.Header().Clone()
			header.Set("Cache-Control", cacheControl)
			body := append([]byte(nil), recorder.body.Bytes()...)
			publicCache.Lock()
			publicCache.entries[key] = publicCacheEntry{status: status, header: header, body: body, expiresAt: now.Add(time.Duration(seconds) * time.Second)}
			publicCache.Unlock()
		}
	}
}

type cacheResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *cacheResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *cacheResponseWriter) WriteString(data string) (int, error) {
	w.body.WriteString(data)
	return w.ResponseWriter.WriteString(data)
}

// CachePrivate sets Cache-Control header to prevent CDN/proxy caching. When a
// positive max-age is provided, it also enables a process-local cache scoped by
// request URI and authenticated user context for repeat dashboard/status reads.
func CachePrivate(seconds ...int) gin.HandlerFunc {
	maxAge := 0
	if len(seconds) > 0 {
		maxAge = seconds[0]
	}
	return func(c *gin.Context) {
		if maxAge <= 0 || c.Request.Method != http.MethodGet {
			c.Header("Cache-Control", "private, no-store")
			c.Next()
			return
		}

		cacheControl := fmt.Sprintf("private, max-age=%d", maxAge)
		c.Header("Cache-Control", cacheControl)
		key := privateCacheKey(c)
		now := time.Now()
		publicCache.RLock()
		entry, ok := publicCache.entries[key]
		publicCache.RUnlock()
		if ok && now.Before(entry.expiresAt) {
			for name, values := range entry.header {
				for _, value := range values {
					c.Writer.Header().Add(name, value)
				}
			}
			c.Header("Cache-Control", cacheControl)
			c.Data(entry.status, entry.header.Get("Content-Type"), entry.body)
			c.Abort()
			return
		}

		recorder := &cacheResponseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = recorder
		c.Next()

		status := recorder.Status()
		if status == http.StatusOK && !c.IsAborted() {
			header := recorder.Header().Clone()
			header.Set("Cache-Control", cacheControl)
			body := append([]byte(nil), recorder.body.Bytes()...)
			publicCache.Lock()
			publicCache.entries[key] = publicCacheEntry{status: status, header: header, body: body, expiresAt: now.Add(time.Duration(maxAge) * time.Second)}
			publicCache.Unlock()
		}
	}
}

func privateCacheKey(c *gin.Context) string {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	teamID, _ := c.Get("team_id")
	return fmt.Sprintf("private:%v:%v:%v:%s", userID, role, teamID, c.Request.URL.RequestURI())
}
