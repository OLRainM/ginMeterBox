package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidateWriteOriginAllowsSameOriginAndConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/write", validateWriteOrigin([]string{"https://console.example.com"}), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for _, origin := range []string{"http://example.test", "https://console.example.com"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "http://example.test/write", nil)
		request.Header.Set("Origin", origin)
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("Origin %q status = %d, want %d", origin, recorder.Code, http.StatusNoContent)
		}
	}
}

func TestValidateWriteOriginRejectsForeignBrowserOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/write", validateWriteOrigin(nil), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "http://example.test/write", nil)
	request.Header.Set("Origin", "https://attacker.example")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("foreign Origin status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestSecurityHeadersAreSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(securityHeaders())
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	for _, header := range []string{"Content-Security-Policy", "X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy", "Permissions-Policy"} {
		if recorder.Header().Get(header) == "" {
			t.Errorf("expected response header %s", header)
		}
	}
}
