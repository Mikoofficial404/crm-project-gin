package v1

import (
	"crm-project/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notifService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notifService,
	}
}

func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	notifications, err := h.notificationService.GetUnreadByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil notifikasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   notifications,
	})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	notifID := c.Param("id")

	err := h.notificationService.MarkAsRead(userID, notifID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Notifikasi berhasil ditandai sudah dibaca",
	})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := h.notificationService.MarkAllAsRead(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Semua notifikasi berhasil dibersihkan",
	})
}
