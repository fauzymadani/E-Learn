package main

import (
	"context"
	"elearning/pkg/grpcclient"
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
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	db, err := config.NewDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}

	tokenMaker, err := token.NewJWTMaker(cfg.JWT.Secret)
	if err != nil {
		log.Fatalf("jwt maker error: %v", err)
	}

	// Initialize token blacklist (cleanup every 1 hour)
	tokenBlacklist := token.NewInMemoryBlacklist(1 * time.Hour)
	log.Println("Token blacklist initialized")

	// Initialize notification client (optional, don't fail if it's unavailable)
	var notifClient *grpcclient.NotificationClient
	notifClient, err = grpcclient.NewNotificationClient(cfg.NotificationGRPC)
	if err != nil {
		log.Printf("Warning: notification service not available: %v", err)
		notifClient = nil
	}

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, tokenMaker, tokenBlacklist, cfg.JWT.Expiration)
	authHandler := handler.NewAuthHandler(authService)

	courseRepo := repository.NewCourseRepository(db)
	courseService := service.NewCourseService(courseRepo)
	courseHandler := handler.NewCourseHandler(courseService)

	lessonRepo := repository.NewLessonRepository(db)
	lessonService := service.NewLessonService(lessonRepo)
	lessonHandler := handler.NewLessonHandler(lessonService)

	enrollmentRepo := repository.NewEnrollmentRepository(db)
	enrollmentService := service.NewEnrollmentService(enrollmentRepo, courseRepo, userRepo, notifClient)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentService)

	progressRepo := repository.NewProgressRepository(db)
	progressService := service.NewProgressService(progressRepo, enrollmentRepo, lessonRepo, courseRepo, notifClient)
	progressHandler := handler.NewProgressHandler(progressService)

	notificationHandler := handler.NewNotificationHandler(notifClient)

	userService := service.NewUserService(userRepo)
	var gcsUploader *storage.GCSUploader
	if cfg.GCS.Enabled {
		gcsUploader, err = storage.NewGCSUploader(cfg.GCS.BucketName)
		if err != nil {
			log.Fatal("Failed to init GCS:", err)
		}
		defer func(gcsUploader *storage.GCSUploader) {
			err := gcsUploader.Close()
			if err != nil {
				log.Printf("Failed to close GCS uploader: %v", err)
			}
		}(gcsUploader)
	}
	userHandler := handler.NewUserHandler(userService, enrollmentRepo, courseRepo, gcsUploader, cfg.GCS.Enabled)

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
		userHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	_ = config.CloseDatabase(db)

	// Close notification client
	if notifClient != nil {
		if err := notifClient.Close(); err != nil {
			log.Printf("failed to close notification client: %v", err)
		}
	}

	log.Println("server exited")
}
