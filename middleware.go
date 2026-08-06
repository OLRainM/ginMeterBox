package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func requestIDAndAccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		startedAt := time.Now()
		c.Next()
		log.Printf("request_id=%s method=%s path=%s status=%d duration_ms=%d client_ip=%s", requestID, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(startedAt).Milliseconds(), c.ClientIP())
	}
}

func newRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(bytes)
}
