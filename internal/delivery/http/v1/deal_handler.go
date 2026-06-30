package v1

import (
	"crm-project/internal/service"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DealHandler struct {
	dealService *service.DealService
}

type CreateDealRequest struct {
	Name   string  `json:"name" binding:"required"`
	Value  float64 `json:"value" binding:"required"`
	LeadID string  `json:"leadId" binding:"required"`
}

type UpdateDealRequest struct {
	Name  string  `json:"name" binding:"required"`
	Value float64 `json:"value" binding:"required"`
}

type ReorderRequest struct {
	DealIDs []string `json:"deal_ids" binding:"required"`
}

type UpdateStageRequest struct {
	Stage string `json:"stage" binding:"required"`
}

func NewDealHandler(dealService *service.DealService) *DealHandler {
	return &DealHandler{
		dealService: dealService,
	}
}

func (h *DealHandler) Reorder(c *gin.Context) {
	var req ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.dealService.ReorderDeals(req.DealIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan urutan: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reorder berhasil"})
}

func (h *DealHandler) CreateDeal(c *gin.Context) {
	var req CreateDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("user_id").(string)
	data, err := h.dealService.CreateDeal(req.Name, req.Value, req.LeadID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "Deal Created", "data": data})
}

func (h *DealHandler) GetDeals(c *gin.Context) {
	userId := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	stage := c.Query("stage")
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	data, total, err := h.dealService.GetDeals(userId, role, page, limit, search, stage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "Get Deals",
		"data":   data,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *DealHandler) UpdateStage(c *gin.Context) {
	dealID := c.Param("id")
	if dealID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID deal tidak boleh kosong"})
		return
	}

	var req UpdateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	err := h.dealService.UpdateStage(dealID, req.Stage, userId, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "Update stage Deal berhasil", "data": gin.H{"id": dealID, "stage": req.Stage}})
}

func (h *DealHandler) DeleteDeal(c *gin.Context) {
	dealID := c.Param("id")
	if dealID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID deal tidak boleh kosong"})
		return
	}

	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	err := h.dealService.DeleteDeal(dealID, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Deal berhasil dihapus", "data": gin.H{"id": dealID}})
}

func (h *DealHandler) ExportCSC(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=laporan_deals.csv")
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	data, _, err := h.dealService.GetDeals(userID, role, 1, 1000, "", "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	writer := csv.NewWriter(c.Writer)
	writer.Write([]string{"ID Deal", "Nama Deal", "Nilai (Rp)", "Status Stage"})
	for _, deal := range data {
		nilaiString := fmt.Sprintf("%.2f", deal.Value)
		barisData := []string{deal.ID, deal.Name, nilaiString, deal.Stage}
		writer.Write(barisData)
	}
	writer.Flush()
}

func (h *DealHandler) ExportPDF(c *gin.Context) {
	dataByte, err := h.dealService.ExportDealsToPDF()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal melukis PDF: " + err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=laporan_deals.pdf")
	c.Header("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", dataByte)
}

func (h *DealHandler) DownloadInvoice(c *gin.Context) {
	dealID := c.Param("id")
	dataByte, err := h.dealService.GenerateInvoicePDF(dealID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal melukis PDF: " + err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=invoice.pdf")
	c.Header("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", dataByte)
}

func (h *DealHandler) ExportExcel(c *gin.Context) {
	fileBuffer, err := h.dealService.ExportDealsToExcel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencetak Excel: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=laporan_deals.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileBuffer.Bytes())
}

func (h *DealHandler) UpdateDeal(c *gin.Context) {
	dealID := c.Param("id")
	if dealID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID deal tidak boleh kosong"})
		return
	}

	var req UpdateDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	err := h.dealService.UpdateDeal(dealID, req.Name, req.Value, userId, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Update Deal berhasil", "data": gin.H{"id": dealID, "name": req.Name, "value": req.Value}})
}
