package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"elearning/internal/config"
	"elearning/internal/domain"
	"elearning/internal/handler"
	"elearning/internal/middleware"
	"elearning/pkg/token"
)

// New initializes the main router for the API.
func New(
	cfg *config.Config,
	db *gorm.DB,
	tokenMaker token.TokenMaker,
	tokenBlacklist token.TokenBlacklist,
	authHandler *handler.AuthHandler,
	courseHandler *handler.CourseHandler,
	lessonHandler *handler.LessonHandler,
	enrollmentHandler *handler.EnrollmentHandler,
) *gin.Engine {

	gin.SetMode(cfg.Server.GinMode)
	r := gin.New()

	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Serve static files (uploaded videos and PDFs)
	r.Static("/uploads", "./uploads")

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// DB Debug Endpoint
	r.GET("/debug/db-check", func(c *gin.Context) {
		var dbName string
		db.Raw("SELECT current_database()").Scan(&dbName)

		var userCount int64
		db.Model(&domain.User{}).Count(&userCount)

		var users []domain.User
		db.Find(&users)

		c.JSON(http.StatusOK, gin.H{
			"database": dbName,
			"users":    users,
			"count":    userCount,
		})
	})

	v1 := r.Group("/api/v1")

	// AUTH ROUTES
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", middleware.AuthMiddleware(tokenMaker, tokenBlacklist), authHandler.GetProfile)
		auth.POST("/logout", middleware.AuthMiddleware(tokenMaker, tokenBlacklist), authHandler.Logout)
	}

	// COURSE ROUTES
	courses := v1.Group("/courses")
	{
		// CREATE COURSE (Teacher/Admin only)
		courses.POST("/",
			middleware.AuthMiddleware(tokenMaker, tokenBlacklist),
			middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin),
			courseHandler.Create,
		)

		// GET ALL COURSES (public)
		courses.GET("/", courseHandler.GetList)

		// GET COURSE BY ID (public)
		courses.GET("/:course_id", courseHandler.Get)

		// UPDATE COURSE (Teacher/Admin only)
		courses.PUT("/:course_id",
			middleware.AuthMiddleware(tokenMaker, tokenBlacklist),
			middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin),
			courseHandler.Update,
		)

		courses.PUT("/:course_id/publish",
			middleware.AuthMiddleware(tokenMaker, tokenBlacklist),
			middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin),
			courseHandler.Publish,
		)

		// DELETE COURSE (Teacher/Admin only)
		courses.DELETE("/:course_id",
			middleware.AuthMiddleware(tokenMaker, tokenBlacklist),
			middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin),
			courseHandler.Delete,
		)
	}

	lessons := v1.Group("/courses/:course_id/lessons")
	{
		lessons.POST("", middleware.AuthMiddleware(tokenMaker, tokenBlacklist), middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin), lessonHandler.Create)
		lessons.GET("", lessonHandler.ListByCourse)
		lessons.GET("/:lesson_id", lessonHandler.Get)
		lessons.PUT("/:lesson_id", middleware.AuthMiddleware(tokenMaker, tokenBlacklist), middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin), lessonHandler.Update)
		lessons.DELETE("/:lesson_id", middleware.AuthMiddleware(tokenMaker, tokenBlacklist), middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin), lessonHandler.Delete)
		lessons.PUT("/reorder", middleware.AuthMiddleware(tokenMaker, tokenBlacklist), middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin), lessonHandler.Reorder)
	}

	// ENROLLMENT ROUTES
	enrollments := v1.Group("/enrollments")
	enrollments.Use(middleware.AuthMiddleware(tokenMaker, tokenBlacklist))
	{
		// Get my enrolled courses
		enrollments.GET("/my-courses", enrollmentHandler.GetMyEnrollments)

		// Update progress
		enrollments.PUT("/:enrollment_id/progress", enrollmentHandler.UpdateProgress)
	}

	// Course-specific enrollment routes
	courseEnrollments := v1.Group("/courses/:course_id")
	{
		// Enroll in a course (students)
		courseEnrollments.POST("/enroll",
			middleware.AuthMiddleware(tokenMaker, tokenBlacklist),
			middleware.RequireRole(domain.RoleStudent, domain.RoleAdmin),
			enrollmentHandler.Enroll,
		)

		// Unenroll from a course
		courseEnrollments.POST("/unenroll",
			middleware.AuthMiddleware(tokenMaker, tokenBlacklist),
			enrollmentHandler.Unenroll,
		)

		// Check enrollment status
		courseEnrollments.GET("/enrollment-status",
			middleware.AuthMiddleware(tokenMaker, tokenBlacklist),
			enrollmentHandler.GetEnrollmentStatus,
		)

		// Get list of enrolled students (teacher only)
		courseEnrollments.GET("/enrollments",
			middleware.AuthMiddleware(tokenMaker, tokenBlacklist),
			middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin),
			enrollmentHandler.GetCourseEnrollments,
		)
	}

	return r
}
