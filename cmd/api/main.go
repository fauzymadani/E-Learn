package main

import (
	"context"
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
	enrollmentService := service.NewEnrollmentService(enrollmentRepo, courseRepo, userRepo)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentService)

	progressRepo := repository.NewProgressRepository(db)
	progressService := service.NewProgressService(progressRepo, enrollmentRepo, lessonRepo)
	progressHandler := handler.NewProgressHandler(progressService)

	r := router.New(cfg, db, tokenMaker, tokenBlacklist, authHandler, courseHandler, lessonHandler, enrollmentHandler, progressHandler)

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

	log.Println("server exited")
}
