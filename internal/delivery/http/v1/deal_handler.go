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

type AssignDealRequest struct {
	AssignedTo string `json:"assigned_to" binding:"required"`
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

// Reorder godoc
// @Summary     Reorder deals
// @Description Ubah urutan deals di kanban board
// @Tags        deals
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body ReorderRequest true "Urutan deal ID"
// @Success     200 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /deals/reorder [patch]
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

// CreateDeal godoc
// @Summary     Buat deal baru
// @Description Buat deal baru ke dalam pipeline
// @Tags        deals
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body CreateDealRequest true "Data deal"
// @Success     201 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /deals [post]
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

// GetDeals godoc
// @Summary     List deals
// @Description Ambil daftar deals dengan filter dan pagination
// @Tags        deals
// @Security    BearerAuth
// @Produce     json
// @Param       page        query int    false "Halaman (default 1)"
// @Param       limit       query int    false "Jumlah per halaman (default 10)"
// @Param       search      query string false "Kata kunci pencarian"
// @Param       stage_id    query string false "Filter stage ID"
// @Param       pipeline_id query string false "Filter pipeline ID"
// @Param       start_date  query string false "Filter dari tanggal (YYYY-MM-DD)"
// @Param       end_date    query string false "Filter sampai tanggal (YYYY-MM-DD)"
// @Success     200 {object} response.PaginatedResponse
// @Failure     400 {object} response.Response
// @Router      /deals [get]
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

// GetDealByID godoc
// @Summary     Get deal detail
// @Description Ambil detail deal berdasarkan ID
// @Tags        deals
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Deal ID"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /deals/{id} [get]
func (h *DealHandler) GetDealByID(c *gin.Context) {
	dealID := c.Param("id")
	if dealID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID deal tidak boleh kosong"))
		return
	}

	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	deal, err := h.dealService.GetDealByID(dealID, userID, role)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil detail deal", deal))
}

// UpdateStage godoc
// @Summary     Update stage deal
// @Description Pindahkan deal ke stage yang berbeda
// @Tags        deals
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string           true "Deal ID"
// @Param       body body UpdateStageRequest true "Stage ID baru"
// @Success     202 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /deals/{id}/stage [patch]
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

// AssignDeal godoc
// @Summary     Re-assign deal
// @Description Memindahkan hak milik deal ke sales lain (Hanya Admin)
// @Tags        deals
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string           true "Deal ID"
// @Param       body body AssignDealRequest true "Data user yang baru"
// @Success     202 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /deals/{id}/assign [patch]
func (h *DealHandler) AssignDeal(c *gin.Context) {
	dealID := c.Param("id")
	if dealID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID deal tidak boleh kosong"))
		return
	}

	var req AssignDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	userId := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	if err := h.dealService.AssignDeal(dealID, req.AssignedTo, userId, role); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusAccepted, response.Success("Pemilik deal berhasil diubah", gin.H{"id": dealID, "assigned_to": req.AssignedTo}))
}

// DeleteDeal godoc
// @Summary     Hapus deal
// @Description Soft delete deal berdasarkan ID
// @Tags        deals
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Deal ID"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /deals/{id} [delete]
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

// ExportCSC godoc
// @Summary     Export deals ke CSV
// @Description Download laporan deals dalam format CSV
// @Tags        deals
// @Security    BearerAuth
// @Produce     text/csv
// @Success     200
// @Failure     400 {object} response.Response
// @Router      /deals/export [get]
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

// ExportPDF godoc
// @Summary     Export deals ke PDF
// @Description Download laporan deals dalam format PDF
// @Tags        deals
// @Security    BearerAuth
// @Produce     application/pdf
// @Success     200
// @Failure     500 {object} response.Response
// @Router      /deals/export/pdf [get]
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

// DownloadInvoice godoc
// @Summary     Download invoice PDF
// @Description Download invoice dalam format PDF berdasarkan Deal ID
// @Tags        deals
// @Security    BearerAuth
// @Produce     application/pdf
// @Param       id path string true "Deal ID"
// @Success     200
// @Failure     500 {object} response.Response
// @Router      /deals/{id}/invoice [get]
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

// ExportExcel godoc
// @Summary     Export deals ke Excel
// @Description Download laporan deals dalam format Excel
// @Tags        deals
// @Security    BearerAuth
// @Produce     application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success     200
// @Failure     500 {object} response.Response
// @Router      /deals/export/excel [get]
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

// UpdateDeal godoc
// @Summary     Update deal
// @Description Update nama dan nilai deal
// @Tags        deals
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string          true "Deal ID"
// @Param       body body UpdateDealRequest true "Data deal yang diupdate"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /deals/{id} [put]
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

// UpdateInvoiceStatus godoc
// @Summary     Update status invoice
// @Description Update status invoice (PAID/OVERDUE)
// @Tags        deals
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string                   true "Invoice ID"
// @Param       body body UpdateInvoiceStatusRequest true "Status baru"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /invoices/{id}/status [patch]
func (h *DealHandler) GetInvoices(c *gin.Context) {
	invoices, err := h.dealService.GetAllInvoices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal mengambil data tagihan"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil data tagihan", invoices))
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

// AssignProduct godoc
// @Summary     Assign produk ke deal
// @Description Tambahkan produk ke dalam deal
// @Tags        deals
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string             true "Deal ID"
// @Param       body body AssignProductRequest true "Data produk"
// @Success     201 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /deals/{id}/products [post]
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

// RemoveProduct godoc
// @Summary     Hapus produk dari deal
// @Description Hapus produk yang di-assign ke deal
// @Tags        deals
// @Security    BearerAuth
// @Produce     json
// @Param       id        path string true "Deal ID"
// @Param       productId path string true "Product ID"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /deals/{id}/products/{productId} [delete]
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

// GetTrashedDeals godoc
// @Summary     Deals yang dihapus
// @Description Ambil daftar deal yang sudah di-soft delete
// @Tags        deals
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /deals/trash [get]
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

// RestoreDeal godoc
// @Summary     Restore deal
// @Description Restore deal yang sudah di-soft delete
// @Tags        deals
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Deal ID"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /deals/{id}/restore [post]
func (h *DealHandler) RestoreDeal(c *gin.Context) {
	dealID := c.Param("id")

	if err := h.dealService.RestoreDeal(dealID); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengembalikan deal", nil))
}

// GetDealProducts godoc
// @Summary     List produk deal
// @Description Ambil semua produk yang di-assign ke deal
// @Tags        deals
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Deal ID"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /deals/{id}/products [get]
func (h *DealHandler) GetDealProducts(c *gin.Context) {
	dealID := c.Param("id")

	products, err := h.dealService.GetDealProducts(dealID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil produk deal", products))
}

// GetDealHistory godoc
// @Summary     History perubahan deal
// @Description Ambil log semua perubahan yang terjadi pada deal
// @Tags        deals
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Deal ID"
// @Success     200 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /deals/{id}/history [get]
func (h *DealHandler) GetDealHistory(c *gin.Context) {
	dealID := c.Param("id")

	history, err := h.dealService.GetHistoriesByDealID(dealID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil history deal", history))
}
