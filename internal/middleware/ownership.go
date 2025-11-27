package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"elearning/internal/domain"
	"elearning/internal/service"
)

// RequireCourseOwnership checks if the authenticated user owns the course
func RequireCourseOwnership(courseService service.CourseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID, err := strconv.ParseInt(c.Param("course_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course_id"})
			c.Abort()
			return
		}

		course, err := courseService.GetByID(courseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			c.Abort()
			return
		}

		claims, err := GetCurrentUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if claims.Role != string(domain.RoleAdmin) && course.TeacherID != int64(claims.UserID) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "you don't have permission to access this course",
			})
			c.Abort()
			return
		}

		c.Set("course", course)
		c.Next()
	}
}

// RequireLessonOwnership checks if the authenticated user owns the lesson's course
func RequireLessonOwnership(lessonService service.LessonServiceInterface, courseService service.CourseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		lessonID, err := strconv.ParseInt(c.Param("lesson_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson_id"})
			c.Abort()
			return
		}

		lesson, err := lessonService.GetLesson(lessonID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
			c.Abort()
			return
		}

		course, err := courseService.GetByID(int64(lesson.CourseID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			c.Abort()
			return
		}

		claims, err := GetCurrentUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if claims.Role != string(domain.RoleAdmin) && course.TeacherID != int64(claims.UserID) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "you don't have permission to access this lesson",
			})
			c.Abort()
			return
		}

		c.Set("lesson", lesson)
		c.Set("course", course)
		c.Next()
	}
}

// GetCourseFromContext Helper functions
func GetCourseFromContext(c *gin.Context) (*domain.Course, bool) {
	course, exists := c.Get("course")
	if !exists {
		return nil, false
	}
	courseObj, ok := course.(*domain.Course)
	return courseObj, ok
}

func GetLessonFromContext(c *gin.Context) (*domain.Lesson, bool) {
	lesson, exists := c.Get("lesson")
	if !exists {
		return nil, false
	}
	lessonObj, ok := lesson.(*domain.Lesson)
	return lessonObj, ok
}
