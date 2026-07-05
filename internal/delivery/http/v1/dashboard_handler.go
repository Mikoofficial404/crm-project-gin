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

func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.dashboardService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal mengambil data dashboard"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil data dashboard", stats))
}

func (h *DashboardHandler) GetAnalytics(c *gin.Context) {
	data, err := h.dashboardService.GetForecastingAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal mengambil data analytics"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil data analytics", data))
}
