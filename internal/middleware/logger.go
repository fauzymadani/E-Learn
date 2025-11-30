package middleware

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const RequestIDKey = "X-Request-ID"

var (
	requestCounter uint64
	randomPrefix   = rand.New(rand.NewSource(time.Now().UnixNano())).Uint32()
)

// generateRequestID generates a unique request ID
// Format: {randomPrefix}-{counter}
// Example: a3f2b1c4-1, a3f2b1c4-2, etc.
func generateRequestID() string {
	counter := atomic.AddUint64(&requestCounter, 1)
	return fmt.Sprintf("%x-%d", randomPrefix, counter)
}

// LoggerMiddleware creates a Gin middleware using Zap logger
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or get request ID
		requestID := c.GetHeader(RequestIDKey)
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Store request ID in context
		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDKey, requestID)

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Get client IP
		clientIP := c.ClientIP()

		// Get X-Forwarded-For header for proxy/load balancer
		forwardedFor := c.GetHeader("X-Forwarded-For")
		realIP := c.GetHeader("X-Real-IP")

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Skip logging for health check and 204 No Content
		if path == "/health" || statusCode == 204 {
			return
		}

		// Build full path with query
		if raw != "" {
			path = path + "?" + raw
		}

		// Prepare log fields
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("client_ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// Add forwarded headers if present
		if forwardedFor != "" {
			fields = append(fields, zap.String("x_forwarded_for", forwardedFor))
		}
		if realIP != "" {
			fields = append(fields, zap.String("x_real_ip", realIP))
		}

		// Add error if exists
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		// Log based on status code
		switch {
		case statusCode >= 500:
			logger.Error("Server error", fields...)
		case statusCode >= 400:
			logger.Warn("Client error", fields...)
		case statusCode >= 300:
			logger.Info("Redirection", fields...)
		default:
			logger.Info("Request completed", fields...)
		}
	}
}

// RequestIDMiddleware ensures every request has a request ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDKey)
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDKey, requestID)
		c.Next()
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
