package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
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
	Name       string  `json:"name" binding:"required"`
	Value      float64 `json:"value" binding:"required"`
	LeadID     string  `json:"leadId" binding:"required"`
	PipelineID string  `json:"pipelineId" binding:"required"`
	StageID    string  `json:"stageId" binding:"required"`
}

type UpdateDealRequest struct {
	Name  string  `json:"name" binding:"required"`
	Value float64 `json:"value" binding:"required"`
}

type ReorderRequest struct {
	DealIDs []string `json:"deal_ids" binding:"required"`
}

type UpdateStageRequest struct {
	StageID string `json:"stage_id" binding:"required"`
}

type UpdateInvoiceStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type AssignProductRequest struct {
	ProductID string  `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	UnitPrice float64 `json:"unit_price" binding:"required,min=0"`
}

func NewDealHandler(dealService *service.DealService) *DealHandler {
	return &DealHandler{dealService: dealService}
}

func (h *DealHandler) Reorder(c *gin.Context) {
	var req ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	if err := h.dealService.ReorderDeals(req.DealIDs); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal menyimpan urutan: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Reorder berhasil", nil))
}

func (h *DealHandler) CreateDeal(c *gin.Context) {
	var req CreateDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	userID := c.MustGet("user_id").(string)
	data, err := h.dealService.CreateDeal(req.Name, req.Value, req.LeadID, userID, req.PipelineID, req.StageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Deal berhasil dibuat", data))
}

func (h *DealHandler) GetDeals(c *gin.Context) {
	userId := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	stageID := c.Query("stage_id")
	pipelineID := c.Query("pipeline_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	data, total, err := h.dealService.GetDeals(userId, role, page, limit, search, stageID, pipelineID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Paginated(data, total, page, limit))
}

func (h *DealHandler) UpdateStage(c *gin.Context) {
	dealID := c.Param("id")
	if dealID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID deal tidak boleh kosong"))
		return
	}

	var req UpdateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	userId := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	if err := h.dealService.UpdateStage(dealID, req.StageID, userId, role); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusAccepted, response.Success("Stage deal berhasil diupdate", gin.H{"id": dealID, "stage_id": req.StageID}))
}

func (h *DealHandler) DeleteDeal(c *gin.Context) {
	dealID := c.Param("id")
	if dealID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID deal tidak boleh kosong"))
		return
	}

	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	if err := h.dealService.DeleteDeal(dealID, userID, role); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Deal berhasil dihapus", gin.H{"id": dealID}))
}

func (h *DealHandler) ExportCSC(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=laporan_deals.csv")
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	data, _, err := h.dealService.GetDeals(userID, role, 1, 1000, "", "", "", "", "")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	writer := csv.NewWriter(c.Writer)
	writer.Write([]string{"ID Deal", "Nama Deal", "Nilai (Rp)", "Status Stage"})
	for _, deal := range data {
		nilaiString := fmt.Sprintf("%.2f", deal.Value)
		barisData := []string{deal.ID, deal.Name, nilaiString, deal.Stage.Name}
		writer.Write(barisData)
	}
	writer.Flush()
}

func (h *DealHandler) ExportPDF(c *gin.Context) {
	dataByte, err := h.dealService.ExportDealsToPDF()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal generate PDF: "+err.Error()))
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
		c.JSON(http.StatusInternalServerError, response.Error("Gagal generate invoice PDF: "+err.Error()))
		return
	}
	c.Header("Content-Disposition", "attachment; filename=invoice.pdf")
	c.Header("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", dataByte)
}

func (h *DealHandler) ExportExcel(c *gin.Context) {
	fileBuffer, err := h.dealService.ExportDealsToExcel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal generate Excel: "+err.Error()))
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=laporan_deals.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileBuffer.Bytes())
}

func (h *DealHandler) UpdateDeal(c *gin.Context) {
	dealID := c.Param("id")
	if dealID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID deal tidak boleh kosong"))
		return
	}

	var req UpdateDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	userId := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	if err := h.dealService.UpdateDeal(dealID, req.Name, req.Value, userId, role); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Update deal berhasil", gin.H{"id": dealID, "name": req.Name, "value": req.Value}))
}

func (h *DealHandler) UpdateInvoiceStatus(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID invoice tidak boleh kosong"))
		return
	}

	var req UpdateInvoiceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	if err := h.dealService.UpdateInvoiceStatus(invoiceID, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Status invoice berhasil diupdate", gin.H{"id": invoiceID, "status": req.Status}))
}

func (h *DealHandler) AssignProduct(c *gin.Context) {
	dealID := c.Param("id")
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	var req AssignProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	result, err := h.dealService.AssignProductToDeal(dealID, req.ProductID, req.Quantity, req.UnitPrice, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Produk berhasil ditambahkan ke deal", result))
}

func (h *DealHandler) RemoveProduct(c *gin.Context) {
	dealID := c.Param("id")
	productID := c.Param("productId")
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	if err := h.dealService.RemoveProductFromDeal(dealID, productID, userID, role); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Produk berhasil dihapus dari deal", nil))
}

func (h *DealHandler) GetTrashedDeals(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	deals, err := h.dealService.GetTrashedDeals(userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil deals yang dihapus", deals))
}

func (h *DealHandler) RestoreDeal(c *gin.Context) {
	dealID := c.Param("id")

	if err := h.dealService.RestoreDeal(dealID); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengembalikan deal", nil))
}

func (h *DealHandler) GetDealProducts(c *gin.Context) {
	dealID := c.Param("id")

	products, err := h.dealService.GetDealProducts(dealID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil produk deal", products))
}

func (h *DealHandler) GetDealHistory(c *gin.Context) {
	dealID := c.Param("id")

	history, err := h.dealService.GetHistoriesByDealID(dealID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil history deal", history))
}
