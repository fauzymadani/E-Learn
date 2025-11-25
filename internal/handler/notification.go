package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"elearning/pkg/grpcclient"
	"elearning/pkg/token"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notifClient *grpcclient.NotificationClient
}

func NewNotificationHandler(notifClient *grpcclient.NotificationClient) *NotificationHandler {
	return &NotificationHandler{notifClient: notifClient}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	payloadRaw, exists := c.Get("auth_payload")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims := payloadRaw.(*token.Claims)
	userID := claims.UserID

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	unreadOnly := c.Query("unread_only") == "true"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.notifClient.GetNotifications(ctx, int64(userID), int32(page), int32(limit), unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notifications"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetUint("user_id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := h.notifClient.GetUnreadCount(ctx, int64(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	// Get user ID from auth claims
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

	notifID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.notifClient.MarkAsRead(ctx, notifID, int64(claims.UserID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	// Get user ID from auth claims
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := h.notifClient.MarkAllAsRead(ctx, int64(claims.UserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark all as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "all notifications marked as read",
		"count":   count,
	})
}
