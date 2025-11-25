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
	authHandler *handler.AuthHandler,
	courseHandler *handler.CourseHandler, // ADD THIS
) *gin.Engine {

	gin.SetMode(cfg.Server.GinMode)
	r := gin.New()

	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

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
		auth.GET("/me", middleware.AuthMiddleware(tokenMaker), authHandler.GetProfile)
	}

	// COURSE ROUTES
	courses := v1.Group("/courses")
	{
		// CREATE COURSE (Teacher/Admin only)
		courses.POST("/",
			middleware.AuthMiddleware(tokenMaker),
			middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin),
			courseHandler.Create,
		)

		// GET ALL COURSES (public)
		courses.GET("/", courseHandler.GetList)

		// GET COURSE BY ID (public)
		courses.GET("/:id", courseHandler.Get)

		// UPDATE COURSE (Teacher/Admin only)
		courses.PUT("/:id",
			middleware.AuthMiddleware(tokenMaker),
			middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin),
			courseHandler.Update,
		)

		courses.PUT("/:id/publish",
			middleware.AuthMiddleware(tokenMaker),
			middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin),
			courseHandler.Publish,
		)

		// DELETE COURSE (Teacher/Admin only)
		courses.DELETE("/:id",
			middleware.AuthMiddleware(tokenMaker),
			middleware.RequireRole(domain.RoleTeacher, domain.RoleAdmin),
			courseHandler.Delete,
		)
	}

	return r
}
