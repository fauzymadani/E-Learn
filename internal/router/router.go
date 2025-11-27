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
	progressHandler *handler.ProgressHandler,
	notificationHandler *handler.NotificationHandler,
	userHandler *handler.UserHandler,
	dashboardHandler *handler.DashboardHandler,
) *gin.Engine {

	gin.SetMode(cfg.Server.GinMode)
	r := gin.New()

	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	r.Use(middleware.CORS())

	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

	// Serve static files (uploaded videos and PDFs)
	r.Static("/uploads", "./uploads")

	// Handle 404 - Not Found
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "endpoint not found",
			"path":  c.Request.URL.Path,
		})
	})

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
		courses.POST("", // ← Hapus "/" jadi ""
			middleware.AuthMiddleware(tokenMaker, tokenBlacklist),
			middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin),
			courseHandler.Create,
		)

		// GET ALL COURSES (public)
		courses.GET("", courseHandler.GetList) // ← Hapus "/" jadi ""

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

	progress := v1.Group("/progress")
	progress.Use(middleware.AuthMiddleware(tokenMaker, tokenBlacklist))
	{
		// Mark lesson as completed
		progress.POST("/lessons/:lesson_id/complete", progressHandler.MarkCompleted)

		// Unmark lesson
		progress.DELETE("/lessons/:lesson_id/complete", progressHandler.UnmarkCompleted)

		// Get lesson progress
		progress.GET("/lessons/:lesson_id", progressHandler.GetLessonProgress)

		// Get course progress
		progress.GET("/courses/:course_id", progressHandler.GetCourseProgress)
	}

	// NOTIFICATION ROUTES
	notifications := v1.Group("/notifications")
	notifications.Use(middleware.AuthMiddleware(tokenMaker, tokenBlacklist))
	{
		// Get all notifications for the current user
		notifications.GET("", notificationHandler.GetNotifications)

		// Get unread count
		notifications.GET("/unread-count", notificationHandler.GetUnreadCount)

		// Mark specific notification as read
		notifications.PUT("/:id/read", notificationHandler.MarkAsRead)

		// Mark all notifications as read
		notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
	}

	users := v1.Group("/users")
	users.Use(middleware.AuthMiddleware(tokenMaker, tokenBlacklist))
	{
		// Get user profile
		users.GET("/profile", userHandler.GetProfile)

		// Edit user profile
		users.PUT("/profile", userHandler.UpdateProfile)

		// change user password
		users.PUT("/change-password", userHandler.ChangePassword)

		// Get courses the user is enrolled in (students and admins)
		users.GET("/enrolled-courses",
			middleware.RequireRole(domain.RoleStudent, domain.RoleAdmin),
			userHandler.GetEnrolledCourses,
		)

		// Get courses the user is teaching (teachers and admins)
		users.GET("/taught-courses",
			middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin),
			userHandler.GetTaughtCourses,
		)
	}

	// DASHBOARD ROUTES
	dashboard := v1.Group("/dashboard")
	dashboard.Use(middleware.AuthMiddleware(tokenMaker, tokenBlacklist))
	{
		// Student dashboard
		dashboard.GET("/student",
			middleware.RequireRole(domain.RoleStudent),
			dashboardHandler.GetStudentDashboard,
		)

		// Teacher dashboard
		dashboard.GET("/teacher",
			middleware.RequireRole(domain.RoleTeacher),
			dashboardHandler.GetTeacherDashboard,
		)

		// Admin dashboard
		dashboard.GET("/admin",
			middleware.RequireRole(domain.RoleAdmin),
			dashboardHandler.GetAdminDashboard,
		)
	}

	return r
}
