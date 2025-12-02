package main

import (
	"context"
	"elearning/pkg/grpcclient"
	"elearning/pkg/logger"
	"elearning/pkg/metrics"
	"elearning/pkg/storage"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"elearning/internal/config"
	"elearning/internal/handler"
	"elearning/internal/repository"
	"elearning/internal/router"
	"elearning/internal/service"
	"elearning/pkg/token"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	// Initialize Zap logger
	zapLogger, err := logger.NewLogger(
		cfg.LogConfig.Level,
		cfg.LogConfig.File,
		cfg.LogConfig.MaxSize,
		cfg.LogConfig.MaxBackups,
		cfg.LogConfig.MaxAge,
	)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	// Defer logger sync with proper error handling
	defer func() {
		// Sync can fail on stdout/stderr, which is normal and can be ignored
		if err := zapLogger.Sync(); err != nil {
			// Only log if it's not the common "invalid argument" error for stdout/stderr
			if !errors.Is(err, syscall.EINVAL) {
				zapLogger.Sugar().Errorf("failed to sync zap logger: %v", err)
			}
		}
	}()

	// Replace standard log with zap logger
	zap.ReplaceGlobals(zapLogger)
	zapLogger.Info("Logger initialized successfully",
		zap.String("level", cfg.LogConfig.Level),
		zap.String("file", cfg.LogConfig.File),
	)

	db, err := config.NewDatabase(cfg.Database)
	if err != nil {
		zapLogger.Fatal("database connection error", zap.Error(err))
	}
	zapLogger.Info("Database connected successfully")

	// Start database metrics collector
	sqlDB, err := db.DB()
	if err != nil {
		zapLogger.Fatal("failed to get sql.DB", zap.Error(err))
	}
	metrics.StartDBMetricsCollector(sqlDB, zapLogger)
	zapLogger.Info("Database metrics collector started")

	tokenMaker, err := token.NewJWTMaker(cfg.JWT.Secret)
	if err != nil {
		zapLogger.Fatal("jwt maker error", zap.Error(err))
	}

	// Initialize token blacklist (cleanup every 1 hour)
	tokenBlacklist := token.NewInMemoryBlacklist(1 * time.Hour)
	zapLogger.Info("Token blacklist initialized")

	// Initialize notification client (optional, don't fail if it's unavailable)
	var notifClient *grpcclient.NotificationClient
	notifClient, err = grpcclient.NewNotificationClient(cfg.NotificationGRPC)
	if err != nil {
		zapLogger.Warn("Notification service not available", zap.Error(err))
		notifClient = nil
	} else {
		zapLogger.Info("Notification client connected", zap.String("address", cfg.NotificationGRPC))
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	courseRepo := repository.NewCourseRepository(db)
	enrollmentRepo := repository.NewEnrollmentRepository(db)
	progressRepo := repository.NewProgressRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, tokenMaker, tokenBlacklist, cfg.JWT.Expiration)
	lessonService := service.NewLessonService(lessonRepo)
	courseService := service.NewCourseService(courseRepo, lessonRepo)
	enrollmentService := service.NewEnrollmentService(enrollmentRepo, courseRepo, userRepo, notifClient)
	progressService := service.NewProgressService(progressRepo, enrollmentRepo, lessonRepo, courseRepo, notifClient)
	userService := service.NewUserService(userRepo)
	dashboardService := service.NewDashboardService(dashboardRepo, notifClient, userRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	lessonHandler := handler.NewLessonHandler(lessonService)
	courseHandler := handler.NewCourseHandler(courseService)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentService)
	progressHandler := handler.NewProgressHandler(progressService)
	notificationHandler := handler.NewNotificationHandler(notifClient)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	adminHandler := handler.NewAdminHandler(userService, notifClient)
	reportsHandler := handler.NewReportsHandler(db, zapLogger)

	// Initialize GCS uploader (optional)
	var gcsUploader *storage.GCSUploader
	if cfg.GCS.Enabled {
		gcsUploader, err = storage.NewGCSUploader(cfg.GCS.BucketName)
		if err != nil {
			zapLogger.Fatal("Failed to init GCS", zap.Error(err))
		}
		defer func() {
			if err := gcsUploader.Close(); err != nil {
				zapLogger.Error("Failed to close GCS uploader", zap.Error(err))
			}
		}()
		zapLogger.Info("GCS uploader initialized", zap.String("bucket", cfg.GCS.BucketName))
	}

	userHandler := handler.NewUserHandler(userService, enrollmentRepo, courseRepo, gcsUploader, cfg.GCS.Enabled)

	// Initialize router with logger
	r := router.New(
		cfg,
		db,
		tokenMaker,
		tokenBlacklist,
		authHandler,
		courseHandler,
		lessonHandler,
		enrollmentHandler,
		progressHandler,
		notificationHandler,
		userHandler,
		dashboardHandler,
		adminHandler,
		reportsHandler,
		courseService,
		lessonService,
		zapLogger,
	)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		zapLogger.Info("Server starting",
			zap.String("port", cfg.Server.Port),
			zap.String("mode", cfg.Server.GinMode),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zapLogger.Fatal("server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Error("Graceful shutdown failed", zap.Error(err))
	}

	// Close database connection
	if err := config.CloseDatabase(db); err != nil {
		zapLogger.Error("Failed to close database", zap.Error(err))
	} else {
		zapLogger.Info("Database connection closed")
	}

	// Close notification client
	if notifClient != nil {
		if err := notifClient.Close(); err != nil {
			zapLogger.Error("Failed to close notification client", zap.Error(err))
		}
	}

	zapLogger.Info("Server exited gracefully")
}
