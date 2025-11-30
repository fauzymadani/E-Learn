package handler

import (
	"elearning/internal/domain"
	"elearning/internal/middleware"
	"elearning/internal/service"
	"elearning/pkg/grpcclient"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userService service.UserService
	notifClient *grpcclient.NotificationClient
}

func NewAdminHandler(userService service.UserService, notifClient *grpcclient.NotificationClient) *AdminHandler {
	return &AdminHandler{
		userService: userService,
		notifClient: notifClient,
	}
}

// GetAllUsers List all users
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// CreateUser create user and notify admin + the created user
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var body struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role" binding:"required,oneof=student teacher admin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(body.Name, body.Email, body.Password, body.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send notifications (best-effort)
	if h.notifClient != nil {
		ctx := c.Request.Context()

		// Notify the created user (optional)
		if err := h.notifClient.SendNotification(ctx, int64(user.ID), string(domain.NotificationTypeCompleted), "Account Created", "Your account has been created by an admin."); err != nil {
			c.Header("X-Notif-User-Error", err.Error())
		}

		// Notify the acting admin (from JWT claims)
		claims, err := middleware.GetCurrentUser(c)
		if err == nil {
			if err := h.notifClient.SendNotification(ctx, int64(claims.UserID), string(domain.NotificationTypeCompleted), "User Created", fmt.Sprintf("You created user ID %d.", user.ID)); err != nil {
				c.Header("X-Notif-Admin-Error", err.Error())
			}
		} else {
			// expose reason we couldn't notify admin
			c.Header("X-Notif-Admin-Error", "missing auth claims: "+err.Error())
		}
	}

	c.JSON(http.StatusCreated, user)
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var body struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role" binding:"required,oneof=student teacher admin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateUser(uint(userID), body.Name, body.Email, body.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.notifClient != nil {
		ctx := c.Request.Context()
		if err := h.notifClient.SendNotification(ctx, int64(userID), string(domain.NotificationTypeCompleted), "Account Updated", "Your account has been updated by an admin."); err != nil {
			c.Header("X-Notif-User-Error", err.Error())
		}
		if claims, err := middleware.GetCurrentUser(c); err == nil {
			if err := h.notifClient.SendNotification(ctx, int64(claims.UserID), string(domain.NotificationTypeCompleted), "User Updated", fmt.Sprintf("You updated user ID %d.", userID)); err != nil {
				c.Header("X-Notif-Admin-Error", err.Error())
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.userService.DeleteUser(uint(userID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.notifClient != nil {
		ctx := c.Request.Context()
		if claims, err := middleware.GetCurrentUser(c); err == nil {
			if err := h.notifClient.SendNotification(ctx, int64(claims.UserID), string(domain.NotificationTypeCompleted), "User Deleted", fmt.Sprintf("You deleted user ID %d.", userID)); err != nil {
				c.Header("X-Notif-Admin-Error", err.Error())
			}
		} else {
			c.Header("X-Notif-Admin-Error", "missing auth claims: "+err.Error())
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
