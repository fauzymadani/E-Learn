package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"elearning/internal/domain"
	"elearning/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LessonHandler struct {
	service service.LessonServiceInterface // INTERFACE
}

func NewLessonHandler(service service.LessonServiceInterface) *LessonHandler {
	return &LessonHandler{service}
}

func (h *LessonHandler) Create(c *gin.Context) {
	courseID, err := strconv.ParseInt(c.Param("course_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	// Parse multipart form with size limit (32 MB)
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to parse multipart form",
			"details": err.Error(),
		})
		return
	}

	// Debug: Log received form fields and files
	log.Printf("Received form fields: %v", c.Request.MultipartForm.Value)
	if c.Request.MultipartForm.File != nil {
		for fieldName, files := range c.Request.MultipartForm.File {
			log.Printf("Received file field '%s' with %d file(s)", fieldName, len(files))
			for i, file := range files {
				log.Printf("  File %d: %s (size: %d bytes)", i, file.Filename, file.Size)
			}
		}
	}

	title := c.PostForm("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	content := c.PostForm("content")

	var videoURL string
	videoFile, err := c.FormFile("video")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(videoFile.Filename))
		if ext != ".mp4" && ext != ".mov" && ext != ".avi" && ext != ".webm" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "video must be mp4, mov, avi, or webm"})
			return
		}

		// Ensure videos directory exists
		if err := os.MkdirAll("uploads/videos", 0755); err != nil {
			log.Printf("Failed to create videos directory: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
			return
		}

		filename := fmt.Sprintf("videos/%d_%d%s", courseID, time.Now().Unix(), ext)
		fullPath := "uploads/" + filename
		if err := c.SaveUploadedFile(videoFile, fullPath); err != nil {
			log.Printf("Failed to save video file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to save video",
				"details": err.Error(),
			})
			return
		}
		videoURL = "/uploads/" + filename
		log.Printf("Video uploaded successfully: %s", fullPath)
	} else if !errors.Is(err, http.ErrMissingFile) {
		log.Printf("Error getting video file: %v", err)
	}

	var fileURL string
	pdfFile, err := c.FormFile("file")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(pdfFile.Filename))
		if ext != ".pdf" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file must be a PDF"})
			return
		}

		// Ensure files directory exists
		if err := os.MkdirAll("uploads/files", 0755); err != nil {
			log.Printf("Failed to create files directory: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
			return
		}

		filename := fmt.Sprintf("files/%d_%d%s", courseID, time.Now().Unix(), ext)
		fullPath := "uploads/" + filename
		if err := c.SaveUploadedFile(pdfFile, fullPath); err != nil {
			log.Printf("Failed to save PDF file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to save file",
				"details": err.Error(),
			})
			return
		}
		fileURL = "/uploads/" + filename
		log.Printf("PDF uploaded successfully: %s", fullPath)
	} else if !errors.Is(err, http.ErrMissingFile) {
		log.Printf("Error getting PDF file: %v", err)
	}

	lesson := domain.Lesson{
		CourseID: uint(courseID),
		Title:    title,
		Content:  content,
		VideoURL: videoURL,
		FileURL:  fileURL,
	}

	if err := h.service.CreateLesson(&lesson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create lesson"})
		return
	}

	c.JSON(http.StatusCreated, lesson)
}

func (h *LessonHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("lesson_id"), 10, 64)

	lesson, err := h.service.GetLesson(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load lesson"})
		return
	}

	c.JSON(http.StatusOK, lesson)
}

func (h *LessonHandler) ListByCourse(c *gin.Context) {
	courseID, _ := strconv.ParseInt(c.Param("course_id"), 10, 64)

	lessons, err := h.service.GetLessonsByCourse(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load lessons"})
		return
	}

	c.JSON(http.StatusOK, lessons)
}

func (h *LessonHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("lesson_id"), 10, 64)

	lesson, err := h.service.GetLesson(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load lesson"})
		return
	}

	// Check content type to determine if this is multipart or JSON
	contentType := c.ContentType()

	if strings.Contains(contentType, "multipart/form-data") {
		// Handle multipart form data (with file uploads)
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "failed to parse multipart form",
				"details": err.Error(),
			})
			return
		}

		// Update text fields if provided
		if title := c.PostForm("title"); title != "" {
			lesson.Title = title
		}
		if content := c.PostForm("content"); content != "" {
			lesson.Content = content
		}

		// Handle video upload
		videoFile, err := c.FormFile("video")
		if err == nil {
			ext := strings.ToLower(filepath.Ext(videoFile.Filename))
			if ext != ".mp4" && ext != ".mov" && ext != ".avi" && ext != ".webm" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "video must be mp4, mov, avi, or webm"})
				return
			}

			if err := os.MkdirAll("uploads/videos", 0755); err != nil {
				log.Printf("Failed to create videos directory: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
				return
			}

			// Delete old video file if exists
			if lesson.VideoURL != "" {
				oldVideoPath := strings.TrimPrefix(lesson.VideoURL, "/")
				if err := os.Remove(oldVideoPath); err != nil && !os.IsNotExist(err) {
					log.Printf("Failed to delete old video file %s: %v", oldVideoPath, err)
				}
			}

			filename := fmt.Sprintf("videos/%d_%d%s", lesson.CourseID, time.Now().Unix(), ext)
			fullPath := "uploads/" + filename
			if err := c.SaveUploadedFile(videoFile, fullPath); err != nil {
				log.Printf("Failed to save video file: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "failed to save video",
					"details": err.Error(),
				})
				return
			}
			lesson.VideoURL = "/uploads/" + filename
			log.Printf("Video updated successfully: %s", fullPath)
		}

		// Handle PDF upload
		pdfFile, err := c.FormFile("file")
		if err == nil {
			ext := strings.ToLower(filepath.Ext(pdfFile.Filename))
			if ext != ".pdf" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "file must be a PDF"})
				return
			}

			if err := os.MkdirAll("uploads/files", 0755); err != nil {
				log.Printf("Failed to create files directory: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
				return
			}

			// Delete old PDF file if exists
			if lesson.FileURL != "" {
				oldFilePath := strings.TrimPrefix(lesson.FileURL, "/")
				if err := os.Remove(oldFilePath); err != nil && !os.IsNotExist(err) {
					log.Printf("Failed to delete old PDF file %s: %v", oldFilePath, err)
				}
			}

			filename := fmt.Sprintf("files/%d_%d%s", lesson.CourseID, time.Now().Unix(), ext)
			fullPath := "uploads/" + filename
			if err := c.SaveUploadedFile(pdfFile, fullPath); err != nil {
				log.Printf("Failed to save PDF file: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "failed to save file",
					"details": err.Error(),
				})
				return
			}
			lesson.FileURL = "/uploads/" + filename
			log.Printf("PDF updated successfully: %s", fullPath)
		}
	} else {
		// Handle JSON data (backward compatibility)
		var body struct {
			Title    *string `json:"title"`
			Content  *string `json:"content"`
			VideoURL *string `json:"video_url"`
			FileURL  *string `json:"file_url"`
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if body.Title != nil {
			lesson.Title = *body.Title
		}
		if body.Content != nil {
			lesson.Content = *body.Content
		}
		if body.VideoURL != nil {
			lesson.VideoURL = *body.VideoURL
		}
		if body.FileURL != nil {
			lesson.FileURL = *body.FileURL
		}
	}

	if err := h.service.UpdateLesson(lesson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update lesson"})
		return
	}

	c.JSON(http.StatusOK, lesson)
}

func (h *LessonHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("lesson_id"), 10, 64)

	err := h.service.DeleteLesson(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete lesson"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *LessonHandler) Reorder(c *gin.Context) {
	courseID, _ := strconv.ParseInt(c.Param("course_id"), 10, 64)

	var body map[int64]int
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.Reorder(courseID, body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
