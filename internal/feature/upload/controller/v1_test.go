package controller

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-hotlines3/internal/feature/auth/policy"

	"github.com/gin-gonic/gin"
)

func TestViewerCannotUseGenericUploadEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &Controller{}
	r.POST("/v1/upload/image", func(c *gin.Context) {
		c.Set("role", policy.RoleViewer)
		h.GetPresignedURL(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/upload/image", bytes.NewBufferString(`{"fileName":"photo.png","fileType":"image/png"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s, want viewer forbidden before storage use", rec.Code, rec.Body.String())
	}
}

func TestGenericUploadDeleteRequiresSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, role := range []string{policy.RoleViewer, policy.RoleUser, policy.RoleTeamLead} {
		t.Run(role, func(t *testing.T) {
			r := gin.New()
			h := &Controller{}
			r.DELETE("/v1/upload/*key", func(c *gin.Context) {
				c.Set("role", role)
				h.DeleteFile(c)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/v1/upload/images/a.png", nil)
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d body=%s, want delete forbidden for %s", rec.Code, rec.Body.String(), role)
			}
		})
	}
}
