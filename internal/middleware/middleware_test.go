package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-hotlines3/internal/dto"
	"backend-hotlines3/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func performRequest(t *testing.T, engine *gin.Engine, method, path string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestCachePublicSetsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/cache", CachePublic(120), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	rec := performRequest(t, r, http.MethodGet, "/cache", nil, nil)
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=120" {
		t.Fatalf("Cache-Control = %q, want public, max-age=120", got)
	}
}

func TestCachePublicCachesSuccessfulGETResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	calls := 0
	r.GET("/cache", CachePublic(120), func(c *gin.Context) {
		calls++
		c.JSON(http.StatusOK, gin.H{"calls": calls})
	})

	first := performRequest(t, r, http.MethodGet, "/cache", nil, nil)
	second := performRequest(t, r, http.MethodGet, "/cache", nil, nil)

	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("statuses = %d/%d, want 200/200", first.Code, second.Code)
	}
	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1 cached backend execution", calls)
	}
	if first.Body.String() != second.Body.String() {
		t.Fatalf("cached response body changed: first=%q second=%q", first.Body.String(), second.Body.String())
	}
}

func TestCachePrivateSetsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/cache", CachePrivate(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	rec := performRequest(t, r, http.MethodGet, "/cache", nil, nil)
	if got := rec.Header().Get("Cache-Control"); got != "private, no-store" {
		t.Fatalf("Cache-Control = %q, want private, no-store", got)
	}
}

func TestCachePrivateCanCacheGETResponsesByUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	calls := 0
	r.GET("/dashboard", func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Set("role", "super_admin")
	}, CachePrivate(60), func(c *gin.Context) {
		calls++
		c.JSON(http.StatusOK, gin.H{"calls": calls})
	})

	first := performRequest(t, r, http.MethodGet, "/dashboard?year=2026", nil, nil)
	second := performRequest(t, r, http.MethodGet, "/dashboard?year=2026", nil, nil)

	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("statuses = %d/%d, want 200/200", first.Code, second.Code)
	}
	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1 cached private execution", calls)
	}
	if first.Body.String() != second.Body.String() {
		t.Fatalf("cached response body changed: first=%q second=%q", first.Body.String(), second.Body.String())
	}
	if got := second.Header().Get("Cache-Control"); got != "private, max-age=60" {
		t.Fatalf("Cache-Control = %q, want private, max-age=60", got)
	}
}

func TestRecoveryMiddlewareReturnsStandardError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RecoveryMiddleware())
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	rec := performRequest(t, r, http.MethodGet, "/panic", nil, nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp dto.StandardResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Success {
		t.Fatal("Success = true, want false")
	}
	if resp.Error == nil || resp.Error.Code != "INTERNAL_SERVER_ERROR" {
		t.Fatalf("Error = %#v, want INTERNAL_SERVER_ERROR", resp.Error)
	}
}

func TestTimeoutMiddlewarePropagatesDeadlineWithoutConcurrentGinWrites(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TimeoutMiddleware(10 * time.Millisecond))
	r.GET("/slow", func(c *gin.Context) {
		select {
		case <-time.After(100 * time.Millisecond):
			c.JSON(http.StatusOK, gin.H{"ok": true})
		case <-c.Request.Context().Done():
			c.Status(http.StatusNoContent)
		}
	})

	rec := performRequest(t, r, http.MethodGet, "/slow", nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s, want handler-controlled timeout response", rec.Code, rec.Body.String())
	}
}

func TestRequireAuthAllowsValidTokenAndSetsClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	teamID := int64(7)
	accessToken, _, err := jwtManager.GenerateTokenPair(42, "alice", "team_lead", &teamID)
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	auth := NewAuthMiddleware(jwtManager)
	r := gin.New()
	r.GET("/secure", auth.RequireAuth(), func(c *gin.Context) {
		if got, ok := c.Get("user_id"); !ok || got.(uint) != 42 {
			t.Fatalf("user_id = %#v, want 42", got)
		}
		if got, ok := c.Get("username"); !ok || got.(string) != "alice" {
			t.Fatalf("username = %#v, want alice", got)
		}
		if got, ok := c.Get("role"); !ok || got.(string) != "team_lead" {
			t.Fatalf("role = %#v, want team_lead", got)
		}
		if got, ok := c.Get("team_id"); !ok || got.(int64) != teamID {
			t.Fatalf("team_id = %#v, want %d", got, teamID)
		}
		c.Status(http.StatusNoContent)
	})

	rec := performRequest(t, r, http.MethodGet, "/secure", nil, map[string]string{"Authorization": "Bearer " + accessToken})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestRequireAuthMissingTokenReturnsUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtManager := jwt.NewJWTManager("secret", time.Hour, 24*time.Hour)
	auth := NewAuthMiddleware(jwtManager)
	r := gin.New()
	r.GET("/secure", auth.RequireAuth(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	rec := performRequest(t, r, http.MethodGet, "/secure", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var resp dto.StandardResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != "MISSING_TOKEN" {
		t.Fatalf("Error = %#v, want MISSING_TOKEN", resp.Error)
	}
}

func TestRequireRoleRejectsInsufficientPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := NewAuthMiddleware(jwt.NewJWTManager("secret", time.Hour, 24*time.Hour))
	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set("role", "staff")
		c.Next()
	}, auth.RequireRole("admin"), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	rec := performRequest(t, r, http.MethodGet, "/admin", nil, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	var resp dto.StandardResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != "FORBIDDEN" {
		t.Fatalf("Error = %#v, want FORBIDDEN", resp.Error)
	}
}
