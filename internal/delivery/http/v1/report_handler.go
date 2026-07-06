package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// @Summary      Sales summary report
// @Tags         Reports
// @Produce      json
// @Security     BearerAuth
// @Param        start_date query string false "Start date (YYYY-MM-DD)"
// @Param        end_date query string false "End date (YYYY-MM-DD)"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /reports/sales-summary [get]
func (h *ReportHandler) GetSalesSummary(c *gin.Context) {
	result, err := h.reportService.GetSalesSummary(c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil sales summary", result))
}

// @Summary      Pipeline report
// @Tags         Reports
// @Produce      json
// @Security     BearerAuth
// @Param        start_date query string false "Start date (YYYY-MM-DD)"
// @Param        end_date query string false "End date (YYYY-MM-DD)"
// @Param        pipeline_id query string false "Pipeline ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /reports/pipeline [get]
func (h *ReportHandler) GetPipelineReport(c *gin.Context) {
	result, err := h.reportService.GetPipelineReport(c.Query("start_date"), c.Query("end_date"), c.Query("pipeline_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil pipeline report", result))
}

// @Summary      Sales performance report
// @Tags         Reports
// @Produce      json
// @Security     BearerAuth
// @Param        start_date query string false "Start date (YYYY-MM-DD)"
// @Param        end_date query string false "End date (YYYY-MM-DD)"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /reports/sales-performance [get]
func (h *ReportHandler) GetSalesPerformance(c *gin.Context) {
	result, err := h.reportService.GetSalesPerformance(c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil sales performance", result))
}

// @Summary      Lead source report
// @Tags         Reports
// @Produce      json
// @Security     BearerAuth
// @Param        start_date query string false "Start date (YYYY-MM-DD)"
// @Param        end_date query string false "End date (YYYY-MM-DD)"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /reports/lead-source [get]
func (h *ReportHandler) GetLeadSourceReport(c *gin.Context) {
	result, err := h.reportService.GetLeadSourceReport(c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil lead source report", result))
}

// @Summary      Activity report
// @Tags         Reports
// @Produce      json
// @Security     BearerAuth
// @Param        start_date query string false "Start date (YYYY-MM-DD)"
// @Param        end_date query string false "End date (YYYY-MM-DD)"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /reports/activity [get]
func (h *ReportHandler) GetActivityReport(c *gin.Context) {
	result, err := h.reportService.GetActivityReport(c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil activity report", result))
}
