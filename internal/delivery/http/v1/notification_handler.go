package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
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

// @Summary      List notifikasi yang belum dibaca
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /notifications [get]
func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	notifications, err := h.notificationService.GetUnreadByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal mengambil notifikasi"))
		return
	}

	c.JSON(http.StatusOK, response.Success("Berhasil mengambil notifikasi", notifications))
}

// @Summary      Tandai notifikasi sudah dibaca
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Notification ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /notifications/{id}/read [patch]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	notifID := c.Param("id")

	err := h.notificationService.MarkAsRead(userID, notifID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success("Notifikasi berhasil ditandai sudah dibaca", nil))
}

// @Summary      Tandai semua notifikasi sudah dibaca
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /notifications/read-all [patch]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	err := h.notificationService.MarkAllAsRead(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success("Semua notifikasi berhasil dibersihkan", nil))
}
