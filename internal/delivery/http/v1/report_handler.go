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

func (h *ReportHandler) GetSalesSummary(c *gin.Context) {
	result, err := h.reportService.GetSalesSummary(c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil sales summary", result))
}

func (h *ReportHandler) GetPipelineReport(c *gin.Context) {
	result, err := h.reportService.GetPipelineReport(c.Query("start_date"), c.Query("end_date"), c.Query("pipeline_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil pipeline report", result))
}

func (h *ReportHandler) GetSalesPerformance(c *gin.Context) {
	result, err := h.reportService.GetSalesPerformance(c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil sales performance", result))
}

func (h *ReportHandler) GetLeadSourceReport(c *gin.Context) {
	result, err := h.reportService.GetLeadSourceReport(c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil lead source report", result))
}

func (h *ReportHandler) GetActivityReport(c *gin.Context) {
	result, err := h.reportService.GetActivityReport(c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil activity report", result))
}
