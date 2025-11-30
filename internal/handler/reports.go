package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"elearning/internal/domain"
)

type ReportsHandler struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewReportsHandler(db *gorm.DB, logger *zap.Logger) *ReportsHandler {
	return &ReportsHandler{
		db:     db,
		logger: logger,
	}
}

// GetOverviewReport returns platform-wide statistics
func (h *ReportsHandler) GetOverviewReport(c *gin.Context) {
	var stats struct {
		TotalUsers       int64 `json:"total_users"`
		TotalCourses     int64 `json:"total_courses"`
		TotalEnrollments int64 `json:"total_enrollments"`
		TotalLessons     int64 `json:"total_lessons"`

		ActiveStudents   int64 `json:"active_students"`
		ActiveTeachers   int64 `json:"active_teachers"`
		PublishedCourses int64 `json:"published_courses"`

		EnrollmentsThisMonth int64 `json:"enrollments_this_month"`
		NewUsersThisMonth    int64 `json:"new_users_this_month"`
	}

	// Use transactions for consistency
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// Total counts
		if err := tx.Model(&domain.User{}).Count(&stats.TotalUsers).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.Course{}).Count(&stats.TotalCourses).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.Enrollment{}).Count(&stats.TotalEnrollments).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.Lesson{}).Count(&stats.TotalLessons).Error; err != nil {
			return err
		}

		// Active users by role
		if err := tx.Model(&domain.User{}).
			Where("role = ?", domain.RoleStudent).
			Count(&stats.ActiveStudents).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.User{}).
			Where("role = ?", domain.RoleTeacher).
			Count(&stats.ActiveTeachers).Error; err != nil {
			return err
		}

		// Published courses
		if err := tx.Model(&domain.Course{}).
			Where("is_published = ?", true).
			Count(&stats.PublishedCourses).Error; err != nil {
			return err
		}

		// This month stats
		startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1).Truncate(24 * time.Hour)

		if err := tx.Model(&domain.Enrollment{}).
			Where("enrolled_at >= ?", startOfMonth).
			Count(&stats.EnrollmentsThisMonth).Error; err != nil {
			return err
		}

		if err := tx.Model(&domain.User{}).
			Where("created_at >= ?", startOfMonth).
			Count(&stats.NewUsersThisMonth).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		h.logger.Error("Failed to fetch overview report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch overview statistics",
		})
		return
	}

	h.logger.Info("Overview report fetched successfully")
	c.JSON(http.StatusOK, stats)
}

// GetEnrollmentReport returns enrollment statistics
func (h *ReportsHandler) GetEnrollmentReport(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		days = 7
	}

	type DailyEnrollment struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	type CourseEnrollment struct {
		CourseID    uint   `json:"course_id"`
		CourseTitle string `json:"course_title"`
		Enrollments int64  `json:"enrollments"`
	}

	var enrollmentTrends []DailyEnrollment
	var topCourses []CourseEnrollment

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// SAFE: Using parameterized query with DATE function
		startDate := time.Now().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)

		// Using GORM's Select with proper parameterization
		if err := tx.Model(&domain.Enrollment{}).
			Select("DATE(enrolled_at) as date, COUNT(*) as count").
			Where("enrolled_at >= ?", startDate).
			Group("DATE(enrolled_at)").
			Order("date ASC").
			Scan(&enrollmentTrends).Error; err != nil {
			return err
		}

		// Top enrolled courses with proper joins
		if err := tx.Model(&domain.Enrollment{}).
			Select("courses.id as course_id, courses.title as course_title, COUNT(enrollments.id) as enrollments").
			Joins("JOIN courses ON courses.id = enrollments.course_id").
			Group("courses.id, courses.title").
			Order("enrollments DESC").
			Limit(10).
			Scan(&topCourses).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		h.logger.Error("Failed to fetch enrollment report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch enrollment statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enrollment_trends": enrollmentTrends,
		"top_courses":       topCourses,
	})
}

// GetUserReport returns user statistics
func (h *ReportsHandler) GetUserReport(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		days = 30
	}

	type DailyUser struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	type RoleDistribution struct {
		Role  string `json:"role"`
		Count int64  `json:"count"`
	}

	type ActiveUser struct {
		UserID      uint   `json:"user_id"`
		Name        string `json:"username"` // Frontend expects 'username'
		Email       string `json:"email"`
		Enrollments int64  `json:"enrollments"`
	}

	var userGrowth []DailyUser
	var roleDistribution []RoleDistribution
	var activeUsers []ActiveUser

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// SAFE: Using GORM query builder with parameterization
		startDate := time.Now().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)

		// User growth with DATE aggregation
		if err := tx.Model(&domain.User{}).
			Select("DATE(created_at) as date, COUNT(*) as count").
			Where("created_at >= ?", startDate).
			Group("DATE(created_at)").
			Order("date ASC").
			Scan(&userGrowth).Error; err != nil {
			return err
		}

		// User distribution by role
		if err := tx.Model(&domain.User{}).
			Select("role, COUNT(*) as count").
			Group("role").
			Scan(&roleDistribution).Error; err != nil {
			return err
		}

		// SAFE: Most active users with parameterized joins
		if err := tx.Model(&domain.User{}).
			Select("users.id as user_id, users.name, users.email, COUNT(enrollments.id) as enrollments").
			Joins("LEFT JOIN enrollments ON enrollments.user_id = users.id").
			Where("users.role = ?", domain.RoleStudent).
			Group("users.id, users.name, users.email").
			Order("enrollments DESC").
			Limit(10).
			Scan(&activeUsers).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		h.logger.Error("Failed to fetch user report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_growth":       userGrowth,
		"role_distribution": roleDistribution,
		"most_active_users": activeUsers,
	})
}

// GetCourseReport returns course statistics
func (h *ReportsHandler) GetCourseReport(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	type CourseCompletion struct {
		CourseID       uint    `json:"course_id"`
		CourseTitle    string  `json:"course_title"`
		TotalStudents  int64   `json:"total_students"`
		CompletedCount int64   `json:"completed_count"`
		CompletionRate float64 `json:"completion_rate"`
	}

	type CategoryStat struct {
		CategoryID   *uint  `json:"category_id"`
		CategoryName string `json:"category_name"`
		CourseCount  int64  `json:"course_count"`
	}

	var courseCompletions []CourseCompletion
	var categoryStats []CategoryStat

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// SAFE: Using GORM query builder with CASE expressions
		// Fixed: Use completed_at instead of progress column
		if err := tx.Model(&domain.Course{}).
			Select(`
				courses.id as course_id,
				courses.title as course_title,
				COUNT(DISTINCT enrollments.id) as total_students,
				COUNT(DISTINCT CASE WHEN enrollments.completed_at IS NOT NULL THEN enrollments.id END) as completed_count,
				COALESCE(
					(COUNT(DISTINCT CASE WHEN enrollments.completed_at IS NOT NULL THEN enrollments.id END)::float / 
					NULLIF(COUNT(DISTINCT enrollments.id), 0)) * 100,
					0
				) as completion_rate
			`).
			Joins("LEFT JOIN enrollments ON enrollments.course_id = courses.id").
			Group("courses.id, courses.title").
			Having("COUNT(DISTINCT enrollments.id) > ?", 0).
			Order("total_students DESC").
			Limit(limit).
			Scan(&courseCompletions).Error; err != nil {
			return err
		}

		// Courses by category
		if err := tx.Model(&domain.Course{}).
			Select("category_id, COUNT(*) as course_count").
			Group("category_id").
			Scan(&categoryStats).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		h.logger.Error("Failed to fetch course report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch course statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"course_completions": courseCompletions,
		"category_stats":     categoryStats,
	})
}

// GetRevenueReport (placeholder for future implementation)
func (h *ReportsHandler) GetRevenueReport(c *gin.Context) {
	h.logger.Info("Revenue report requested (not implemented)")
	c.JSON(http.StatusOK, gin.H{
		"message":         "Revenue reporting not implemented yet",
		"total_revenue":   0,
		"monthly_revenue": []gin.H{},
	})
}
