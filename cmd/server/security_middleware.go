package main

import (
	"net/http"
	"net/url"
	"strings"

	"ginMeterBox/internal/platform/response"

	"github.com/gin-gonic/gin"
)

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'")
		c.Next()
	}
}

// validateWriteOrigin reduces CSRF exposure for cookie-authenticated mutations.
// SameSite=Strict protects the normal browser path; this check also rejects any
// browser request whose Origin is neither the current host nor an explicit CORS allowlist entry.
func validateWriteOrigin(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if parsed, err := url.Parse(origin); err == nil && parsed.Scheme != "" && parsed.Host != "" && parsed.Path == "" {
			allowed[strings.TrimRight(origin, "/")] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		origin := strings.TrimRight(strings.TrimSpace(c.GetHeader("Origin")), "/")
		if origin == "" {
			// Non-browser API clients do not reliably send Origin. They still need a valid session.
			c.Next()
			return
		}
		if isSameOrigin(origin, c.Request) {
			c.Next()
			return
		}
		if _, ok := allowed[origin]; ok {
			c.Next()
			return
		}

		response.BadRequest(c, "请求来源不被允许")
		c.Abort()
	}
}

func isSameOrigin(origin string, request *http.Request) bool {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.Path != "" {
		return false
	}

	scheme := "http"
	if request.TLS != nil {
		scheme = "https"
	}
	return parsed.Scheme == scheme && strings.EqualFold(parsed.Host, request.Host)
}
