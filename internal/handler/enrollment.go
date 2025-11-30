package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"elearning/internal/middleware"
	"elearning/internal/repository"
	"elearning/internal/service"
)

// EnrollmentHandler handles enrollment requests
type EnrollmentHandler struct {
	service *service.EnrollmentService
}

// NewEnrollmentHandler creates a new enrollment handler
func NewEnrollmentHandler(service *service.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{service: service}
}

// Enroll handles student enrollment in a course
// @Summary Enroll in a course
// @Description Student enrolls in a course
// @Tags enrollments
// @Accept json
// @Produce json
// @Param course_id path int true "Course ID"
// @Security BearerAuth
// @Success 201 {object} domain.Enrollment
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /courses/{course_id}/enroll [post]
func (h *EnrollmentHandler) Enroll(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	courseID, err := strconv.ParseUint(c.Param("course_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	enrollment, err := h.service.Enroll(claims.UserID, uint(courseID))
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyEnrolled) {
			c.JSON(http.StatusConflict, gin.H{"error": "already enrolled in this course"})
			return
		}
		if errors.Is(err, service.ErrCourseNotPublished) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "course is not published"})
			return
		}
		if errors.Is(err, service.ErrCannotEnrollInOwnCourse) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot enroll in your own course"})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enroll"})
		return
	}

	c.JSON(http.StatusCreated, enrollment)
}

// Unenroll handles student unenrollment from a course
// @Summary Unenroll from a course
// @Description Student unenrolls from a course
// @Tags enrollments
// @Produce json
// @Param course_id path int true "Course ID"
// @Security BearerAuth
// @Success 200 {object} map[string]bool
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /courses/{course_id}/unenroll [post]
func (h *EnrollmentHandler) Unenroll(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	courseID, err := strconv.ParseUint(c.Param("course_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	if err := h.service.Unenroll(claims.UserID, uint(courseID)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "enrollment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unenroll"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetMyEnrollments returns all enrollments for the current user
// @Summary Get my enrollments
// @Description Get all courses the current user is enrolled in
// @Tags enrollments
// @Produce json
// @Param status query string false "Filter by status (active, completed, dropped)"
// @Security BearerAuth
// @Success 200 {array} domain.Enrollment
// @Failure 401 {object} ErrorResponse
// @Router /enrollments/my-courses [get]
func (h *EnrollmentHandler) GetMyEnrollments(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	status := c.Query("status")

	enrollments, err := h.service.GetMyEnrollments(claims.UserID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get enrollments"})
		return
	}

	c.JSON(http.StatusOK, enrollments)
}

// GetCourseEnrollments returns all enrollments for a course (teacher view)
// @Summary Get course enrollments
// @Description Get all students enrolled in a course (teacher only)
// @Tags enrollments
// @Produce json
// @Param course_id path int true "Course ID"
// @Security BearerAuth
// @Success 200 {array} domain.Enrollment
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /courses/{course_id}/enrollments [get]
func (h *EnrollmentHandler) GetCourseEnrollments(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	courseID, err := strconv.ParseUint(c.Param("course_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	enrollments, err := h.service.GetCourseEnrollments(uint(courseID), claims.UserID)
	if err != nil {
		if err.Error() == "not authorized to view enrollments for this course" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get enrollments"})
		return
	}

	c.JSON(http.StatusOK, enrollments)
}

// GetEnrollmentStatus checks enrollment status for a specific course
// @Summary Get enrollment status
// @Description Check if current user is enrolled in a course
// @Tags enrollments
// @Produce json
// @Param course_id path int true "Course ID"
// @Security BearerAuth
// @Success 200 {object} domain.Enrollment
// @Failure 404 {object} ErrorResponse
// @Router /courses/{course_id}/enrollment-status [get]
func (h *EnrollmentHandler) GetEnrollmentStatus(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	courseID, err := strconv.ParseUint(c.Param("course_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	enrollment, err := h.service.GetEnrollmentStatus(claims.UserID, uint(courseID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"enrolled": false,
				"message":  "not enrolled in this course",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get enrollment status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enrolled":   true,
		"enrollment": enrollment,
	})
}

// UpdateProgress updates enrollment progress
// @Summary Update enrollment progress
// @Description Update progress percentage for an enrollment
// @Tags enrollments
// @Accept json
// @Produce json
// @Param enrollment_id path int true "Enrollment ID"
// @Param body ProgressUpdateRequest true "Progress update"
// @Security BearerAuth
// @Success 200 {object} map[string]bool
// @Failure 400 {object} ErrorResponse
// @Router /enrollments/{enrollment_id}/progress [put]
func (h *EnrollmentHandler) UpdateProgress(c *gin.Context) {
	enrollmentID, err := strconv.ParseUint(c.Param("enrollment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid enrollment id"})
		return
	}

	var req struct {
		Progress float64 `json:"progress" binding:"required,min=0,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateProgress(uint(enrollmentID), req.Progress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
