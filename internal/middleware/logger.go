package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger is a custom logging middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Log request details
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		log.Printf(
			"[%s] %s %s | Status: %d | Duration: %v | IP: %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.Proto,
			statusCode,
			duration,
			c.ClientIP(),
		)
	}
}

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
