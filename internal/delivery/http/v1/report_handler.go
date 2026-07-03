package v1

import (
	"crm-project/internal/service"
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
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	result, err := h.reportService.GetSalesSummary(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil sales summary", "data": result})
}

func (h *ReportHandler) GetPipelineReport(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	pipelineID := c.Query("pipeline_id")

	result, err := h.reportService.GetPipelineReport(startDate, endDate, pipelineID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil pipeline report", "data": result})
}

func (h *ReportHandler) GetSalesPerformance(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	result, err := h.reportService.GetSalesPerformance(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil sales performance", "data": result})
}

func (h *ReportHandler) GetLeadSourceReport(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	result, err := h.reportService.GetLeadSourceReport(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil lead source report", "data": result})
}

func (h *ReportHandler) GetActivityReport(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	result, err := h.reportService.GetActivityReport(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil activity report", "data": result})
}
