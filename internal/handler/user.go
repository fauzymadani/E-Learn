package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"elearning/internal/domain"
	"elearning/internal/middleware"
	"elearning/internal/repository"
	"elearning/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService    service.UserService
	enrollmentRepo repository.EnrollmentRepository
	courseRepo     repository.CourseRepository
}

func NewUserHandler(
	userService service.UserService,
	enrollmentRepo repository.EnrollmentRepository,
	courseRepo repository.CourseRepository,
) *UserHandler {
	return &UserHandler{
		userService:    userService,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.GetProfile(claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	response := gin.H{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.Role,
		"avatar":     user.Avatar,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	if claims.Role == string(domain.RoleStudent) {
		enrollments, _ := h.enrollmentRepo.FindByUser(claims.UserID)
		response["enrolled_courses_count"] = len(enrollments)
	}

	if claims.Role == string(domain.RoleTeacher) {
		courses, total, _ := h.courseRepo.GetByInstructorID(c.Request.Context(), claims.UserID, 1, 1)
		_ = courses
		response["taught_courses_count"] = total
	}

	c.JSON(http.StatusOK, response)
}

type UpdateProfileRequest struct {
	Name   string  `json:"name"`
	Avatar *string `json:"avatar"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	name := c.PostForm("name")
	var avatarURL *string

	file, err := c.FormFile("avatar")
	if err == nil {
		allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
		ext := strings.ToLower(filepath.Ext(file.Filename))

		if !allowedExts[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only image files allowed (jpg, jpeg, png, gif)"})
			return
		}

		if file.Size > 5*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File size must be less than 5MB"})
			return
		}

		uploadDir := "./uploads/avatars"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		filename := fmt.Sprintf("%d_%d%s", claims.UserID, time.Now().Unix(), ext)
		join := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, join); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		url := fmt.Sprintf("/uploads/avatars/%s", filename)
		avatarURL = &url
	}

	if err := h.userService.UpdateProfile(claims.UserID, name, avatarURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.userService.ChangePassword(claims.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOldPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (h *UserHandler) GetEnrolledCourses(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != string(domain.RoleStudent) && claims.Role != string(domain.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only students can view enrolled courses"})
		return
	}

	enrollments, err := h.enrollmentRepo.FindByUser(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch enrolled courses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enrollments": enrollments,
		"total":       len(enrollments),
	})
}

func (h *UserHandler) GetTaughtCourses(c *gin.Context) {
	claims, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if claims.Role != string(domain.RoleTeacher) && claims.Role != string(domain.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only teachers can view taught courses"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	courses, total, err := h.courseRepo.GetByInstructorID(c.Request.Context(), claims.UserID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch taught courses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"courses": courses,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}
