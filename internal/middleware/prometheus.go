package middleware

import (
	"strconv"
	"time"

	"elearning/pkg/metrics"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware records HTTP metrics
func PrometheusMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		// Use route pattern instead of actual path
		if path == "" {
			path = c.Request.URL.Path // Use actual path if route pattern is empty
		}

		logger.Debug("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("url", c.Request.URL.Path),
		)

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// Record metrics
		metrics.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Inc()

		metrics.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Observe(duration)

		metrics.HttpResponseSize.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(float64(c.Writer.Size()))

		if len(c.Errors) > 0 {
			metrics.ErrorsTotal.WithLabelValues(
				"gin_error",
				path,
			).Inc()
		}

		logger.Debug("HTTP Response",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("status", status),
			zap.Float64("duration_seconds", duration),
		)
	}
}
