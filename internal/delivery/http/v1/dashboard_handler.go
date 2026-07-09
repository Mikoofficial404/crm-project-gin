package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// @Summary      Get statistik dashboard
// @Tags         Dashboard
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /dashboard [get]
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("role")

	stats, err := h.dashboardService.GetStats(role, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal mengambil data dashboard"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil data dashboard", stats))
}

// @Summary      Get analytics forecasting
// @Tags         Dashboard
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /analytics/forecasting [get]
func (h *DashboardHandler) GetAnalytics(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("role")

	data, err := h.dashboardService.GetForecastingAnalytics(role, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal mengambil data analytics"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil data analytics", data))
}
