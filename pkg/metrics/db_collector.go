package metrics

import (
	"database/sql"
	"time"

	"go.uber.org/zap"
)

// StartDBMetricsCollector starts collecting database metrics
func StartDBMetricsCollector(db *sql.DB, logger *zap.Logger) {
	ticker := time.NewTicker(5 * time.Second) // Collect every 5 seconds instead of 15
	go func() {
		for range ticker.C {
			stats := db.Stats()

			// Update connection metrics
			DbConnectionsActive.Set(float64(stats.InUse))
			DbConnectionsIdle.Set(float64(stats.Idle))

			// Log with info level to see in production
			logger.Info("Database metrics collected",
				zap.Int("active_connections", stats.InUse),
				zap.Int("idle_connections", stats.Idle),
				zap.Int("max_open_connections", stats.MaxOpenConnections),
				zap.Int64("wait_count", stats.WaitCount),
				zap.Int64("wait_duration_ns", int64(stats.WaitDuration)),
				zap.Int64("max_idle_closed", stats.MaxIdleClosed),
				zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
			)
		}
	}()
}
