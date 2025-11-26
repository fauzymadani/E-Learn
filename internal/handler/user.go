package handler

import (
	"bytes"
	"elearning/internal/domain"
	"elearning/internal/middleware"
	"elearning/internal/repository"
	"elearning/internal/service"
	"elearning/pkg/storage"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService    service.UserService
	enrollmentRepo repository.EnrollmentRepository
	courseRepo     repository.CourseRepository
	gcsUploader    *storage.GCSUploader
	gcsEnabled     bool
}

func NewUserHandler(
	userService service.UserService,
	enrollmentRepo repository.EnrollmentRepository,
	courseRepo repository.CourseRepository,
	gcsUploader *storage.GCSUploader,
	gcsEnabled bool,
) *UserHandler {
	return &UserHandler{
		userService:    userService,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
		gcsUploader:    gcsUploader,
		gcsEnabled:     gcsEnabled,
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

	c.JSON(http.StatusOK, user)
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
		allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
		ext := strings.ToLower(filepath.Ext(file.Filename))

		if !allowedExts[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPG, JPEG, PNG allowed"})
			return
		}

		if file.Size > 5*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File size must be less than 5MB"})
			return
		}

		user, err := h.userService.GetProfile(claims.UserID)
		if err == nil && user.Avatar != nil && *user.Avatar != "" {
			if h.gcsEnabled {
				filename := filepath.Base(*user.Avatar)
				err := h.gcsUploader.Delete(c.Request.Context(), filename)
				if err != nil {
					return
				}
			} else {
				oldPath := "." + *user.Avatar
				if _, err := os.Stat(oldPath); err == nil {
					err := os.Remove(oldPath)
					if err != nil {
						return
					}
				}
			}
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		defer func(src multipart.File) {
			err := src.Close()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close file"})
			}
		}(src)

		img, _, err := image.Decode(src)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image file"})
			return
		}

		img = imaging.Fit(img, 500, 500, imaging.Lanczos)

		filename := fmt.Sprintf("avatars/%d_%d.jpg", claims.UserID, time.Now().Unix())

		if h.gcsEnabled {
			buf := new(bytes.Buffer)
			if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compress image"})
				return
			}

			url, err := h.gcsUploader.Upload(c.Request.Context(), buf, filename)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to cloud"})
				return
			}
			avatarURL = &url
		} else {
			uploadDir := "./uploads/avatars"
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
				return
			}

			savePath := filepath.Join(uploadDir, filepath.Base(filename))
			out, err := os.Create(savePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
				return
			}
			defer func(out *os.File) {
				err := out.Close()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close file"})
				}
			}(out)

			if err := jpeg.Encode(out, img, &jpeg.Options{Quality: 85}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compress image"})
				return
			}

			url := fmt.Sprintf("/uploads/avatars/%s", filepath.Base(filename))
			avatarURL = &url
		}
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
