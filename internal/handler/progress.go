package handler

import (
	"net/http"
	"strconv"

	"elearning/internal/middleware"
	"elearning/internal/service"

	"github.com/gin-gonic/gin"
)

type ProgressHandler struct {
	service *service.ProgressService
}

func NewProgressHandler(service *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{service: service}
}

// MarkCompleted marks a lesson as completed
func (h *ProgressHandler) MarkCompleted(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	lessonID, err := strconv.ParseUint(c.Param("lesson_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	if err := h.service.MarkLessonCompleted(claims.UserID, uint(lessonID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "lesson marked as completed",
		"lesson_id": lessonID,
	})
}

// UnmarkCompleted unmarks a lesson as completed
func (h *ProgressHandler) UnmarkCompleted(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	lessonID, err := strconv.ParseUint(c.Param("lesson_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	if err := h.service.UnmarkLessonCompleted(claims.UserID, uint(lessonID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "lesson unmarked",
		"lesson_id": lessonID,
	})
}

// GetCourseProgress gets overall progress for a course
func (h *ProgressHandler) GetCourseProgress(c *gin.Context) {
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

	progress, err := h.service.GetCourseProgress(claims.UserID, uint(courseID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetLessonProgress gets progress for a specific lesson
func (h *ProgressHandler) GetLessonProgress(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	lessonID, err := strconv.ParseUint(c.Param("lesson_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson id"})
		return
	}

	progress, err := h.service.GetLessonProgress(claims.UserID, uint(lessonID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "progress not found"})
		return
	}

	c.JSON(http.StatusOK, progress)
}
