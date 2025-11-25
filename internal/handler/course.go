package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"elearning/internal/domain"
	"elearning/internal/middleware"
	"elearning/internal/service"
)

type CourseHandler struct {
	service service.CourseService
}

func NewCourseHandler(s service.CourseService) *CourseHandler {
	return &CourseHandler{s}
}

func (h *CourseHandler) Create(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Thumbnail   string `json:"thumbnail"`
		CategoryID  *int64 `json:"category_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, _ := middleware.GetCurrentUser(c)

	course := domain.Course{
		Title:       req.Title,
		Description: req.Description,
		Thumbnail:   req.Thumbnail,
		CategoryID:  req.CategoryID,
		TeacherID:   int64(claims.UserID),
		IsPublished: false,
	}

	if err := h.service.Create(&course); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create course"})
		return
	}

	c.JSON(http.StatusCreated, course)
}

func (h *CourseHandler) GetList(c *gin.Context) {
	f := map[string]interface{}{}

	if q := c.Query("title"); q != "" {
		f["title"] = q
	}
	if cat := c.Query("category_id"); cat != "" {
		id, _ := strconv.ParseInt(cat, 10, 64)
		f["category_id"] = id
	}

	courses, err := h.service.GetList(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch courses"})
		return
	}

	c.JSON(http.StatusOK, courses)
}

func (h *CourseHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	course, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}
	c.JSON(http.StatusOK, course)
}

func (h *CourseHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	course, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Thumbnail   string `json:"thumbnail"`
		CategoryID  *int64 `json:"category_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	course.Title = req.Title
	course.Description = req.Description
	course.Thumbnail = req.Thumbnail
	course.CategoryID = req.CategoryID

	if err := h.service.Update(course); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update course"})
		return
	}

	c.JSON(http.StatusOK, course)
}

func (h *CourseHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete course"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CourseHandler) Publish(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var body struct {
		Publish bool `json:"publish"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Publish(id, body.Publish); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update publish state"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"published": body.Publish})
}
