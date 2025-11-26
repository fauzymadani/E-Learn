package handler

import (
	"net/http"

	"elearning/internal/service"
	"elearning/pkg/token"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService service.DashboardService
}

func NewDashboardHandler(dashboardService service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetStudentDashboard godoc
// @Summary Get student dashboard
// @Description Get student dashboard with enrolled courses, learning progress, and recent notifications
// @Tags dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.StudentDashboard
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Failure 403 {object} map[string]interface{} "access denied - students only"
// @Failure 500 {object} map[string]interface{} "internal server error"
// @Router /dashboard/student [get]
func (h *DashboardHandler) GetStudentDashboard(c *gin.Context) {
	payloadRaw, exists := c.Get("auth_payload")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims, ok := payloadRaw.(*token.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth payload"})
		return
	}

	dashboard, err := h.dashboardService.GetStudentDashboard(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetTeacherDashboard godoc
// @Summary Get teacher dashboard
// @Description Get teacher dashboard with courses, total students, and recent enrollments
// @Tags dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.TeacherDashboard
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Failure 403 {object} map[string]interface{} "access denied - teachers only"
// @Failure 500 {object} map[string]interface{} "internal server error"
// @Router /dashboard/teacher [get]
func (h *DashboardHandler) GetTeacherDashboard(c *gin.Context) {
	payloadRaw, exists := c.Get("auth_payload")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims, ok := payloadRaw.(*token.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth payload"})
		return
	}

	dashboard, err := h.dashboardService.GetTeacherDashboard(c.Request.Context(), int64(claims.UserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetAdminDashboard godoc
// @Summary Get admin dashboard
// @Description Get admin dashboard with total users, courses, enrollments, and statistics
// @Tags dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.AdminDashboard
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Failure 403 {object} map[string]interface{} "access denied - admins only"
// @Failure 500 {object} map[string]interface{} "internal server error"
// @Router /dashboard/admin [get]
func (h *DashboardHandler) GetAdminDashboard(c *gin.Context) {

	dashboard, err := h.dashboardService.GetAdminDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
