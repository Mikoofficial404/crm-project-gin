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

	type StageData struct {
		StageID    string  `json:"stage_id"`
		StageName  string  `json:"stage_name"`
		Name       string  `json:"name"`
		DealsCount int64   `json:"deals_count"`
		TotalValue float64 `json:"total_value"`
	}

	var pipelineID, pipelineName string
	var stages []StageData
	var totalDeals int64
	var totalValue float64

	for _, st := range result {
		pipelineID = st.PipelineID
		pipelineName = st.PipelineName
		stages = append(stages, StageData{
			StageID:    st.StageID,
			StageName:  st.StageName,
			Name:       st.StageName,
			DealsCount: st.DealCount,
			TotalValue: st.TotalValue,
		})
		totalDeals += st.DealCount
		totalValue += st.TotalValue
	}

	finalResult := gin.H{
		"pipeline_id":   pipelineID,
		"pipeline_name": pipelineName,
		"stages":        stages,
		"total_deals":   totalDeals,
		"total_value":   totalValue,
	}

	c.JSON(http.StatusOK, response.Success("Berhasil mengambil pipeline report", finalResult))
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

	type SourceData struct {
		Name       string  `json:"name"`
		Count      int64   `json:"count"`
		Percentage float64 `json:"percentage"`
	}

	var total int64
	var sources []SourceData

	for _, rs := range result {
		total += rs.TotalLeads
	}

	for _, rs := range result {
		perc := float64(0)
		if total > 0 {
			perc = (float64(rs.TotalLeads) / float64(total)) * 100
		}
		sources = append(sources, SourceData{
			Name:       rs.Source,
			Count:      rs.TotalLeads,
			Percentage: perc,
		})
	}

	finalResult := gin.H{
		"sources": sources,
		"total":   total,
	}

	c.JSON(http.StatusOK, response.Success("Berhasil mengambil lead source report", finalResult))
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

	type ActivityData struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	var total int64
	byType := make(map[string]int64)
	var activities []ActivityData

	for _, rs := range result {
		total += rs.TotalCount
		byType[rs.ActivityType] += rs.TotalCount
	}
	for k, v := range byType {
		activities = append(activities, ActivityData{
			Type:  k,
			Count: v,
		})
	}

	finalResult := gin.H{
		"activities": activities,
		"total":      total,
		"by_type":    byType,
	}

	c.JSON(http.StatusOK, response.Success("Berhasil mengambil activity report", finalResult))
}
