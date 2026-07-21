package v1

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"crm-project/internal/service"
	"crm-project/pkg/response"
	"encoding/csv"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const ImportCSVMaxSize = 10 << 20 // 10 MB

type LeadHandler struct {
	leadService    *service.LeadService
	leadRepository *postgres.LeadRepository
	waService      *service.WhatsAppService
	activtyService *service.ActivityService
}

func NewLeadHandler(leadService *service.LeadService, leadRepo *postgres.LeadRepository, waServ *service.WhatsAppService, activityServ *service.ActivityService) *LeadHandler {
	return &LeadHandler{
		leadService:    leadService,
		leadRepository: leadRepo,
		waService:      waServ,
		activtyService: activityServ,
	}
}

type UpdateStatusReques struct {
	Status string `json:"status" binding:"required"`
}

type SendReplyRequest struct {
	Message string `json:"message" binding:"required"`
}

type CreateLeadRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Email        string                 `json:"email" binding:"required,email"`
	Phone        string                 `json:"phone" binding:"required"`
	CustomFields map[string]interface{} `json:"custom_fields" binding:"required"`
}

type UpdateLeadRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Phone string `json:"phone" binding:"required"`
}

// CreateLeader godoc
// @Summary     Buat lead baru
// @Description Buat lead baru ke sistem
// @Tags        leads
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body CreateLeadRequest true "Data lead"
// @Success     201 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /leads [post]
func (h *LeadHandler) CreateLeader(c *gin.Context) {
	var req CreateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	userID := c.MustGet("user_id").(string)
	data, err := h.leadService.CreateLead(req.Name, req.Email, req.Phone, userID, req.CustomFields)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Lead berhasil dibuat", data))
}

// GetTrashedLeads godoc
// @Summary     Lead yang dihapus
// @Description Ambil daftar lead yang sudah di-soft delete
// @Tags        leads
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /leads/trash [get]
func (h *LeadHandler) GetTrashedLeads(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	data, err := h.leadService.GetTrashedLeads(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil lead yang dihapus", data))
}

// RestoreLead godoc
// @Summary     Restore lead
// @Description Pulihkan lead yang sudah dihapus (admin only)
// @Tags        leads
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Lead ID"
// @Success     200 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /leads/trash/{id}/restore [post]
func (h *LeadHandler) RestoreLead(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("user_id").(string)

	err := h.leadService.RestoreLead(id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Lead berhasil direstore", nil))
}

// GetLeads godoc
// @Summary     List leads
// @Description Ambil daftar leads dengan filter dan pagination
// @Tags        leads
// @Security    BearerAuth
// @Produce     json
// @Param       page      query int    false "Halaman (default 1)"
// @Param       limit     query int    false "Jumlah per halaman (default 10)"
// @Param       search    query string false "Kata kunci pencarian"
// @Param       status    query string false "Filter status lead"
// @Param       startDate query string false "Filter dari tanggal (YYYY-MM-DD)"
// @Param       endDate   query string false "Filter sampai tanggal (YYYY-MM-DD)"
// @Success     200 {object} response.PaginatedResponse
// @Failure     400 {object} response.Response
// @Router      /leads [get]
func (h *LeadHandler) GetLeads(c *gin.Context) {
	userId := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	status := c.Query("status")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	data, total, err := h.leadService.GetLeads(userId, role, page, limit, search, status, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Paginated(data, total, page, limit))
}

// GetLeadByID godoc
// @Summary     Detail lead
// @Description Ambil detail lead berdasarkan ID
// @Tags        leads
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Lead ID"
// @Success     200 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /leads/{id} [get]
func (h *LeadHandler) GetLeadByID(c *gin.Context) {
	leadID := c.Param("id")
	if leadID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID lead tidak boleh kosong"))
		return
	}

	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	data, err := h.leadService.GetLeadByID(leadID, userID, role)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Lead tidak ditemukan"))
		return
	}

	c.JSON(http.StatusOK, response.Success("Berhasil mengambil lead", data))
}

// UpdateStatusLeads godoc
// @Summary     Update status lead
// @Description Update status lead (NEW, CONTACTED, dll)
// @Tags        leads
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string             true "Lead ID"
// @Param       body body UpdateStatusReques true "Status baru"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /leads/{id}/status [patch]
func (h *LeadHandler) UpdateStatusLeads(c *gin.Context) {
	catchId := c.Param("id")
	var request struct {
		Status string `json:"status"`
	}
	if catchId == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID prospek tidak boleh kosong"))
		return
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	userId := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	err := h.leadService.UpdateLeadStatus(catchId, request.Status, userId, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusAccepted, response.Success("Status lead berhasil diupdate", gin.H{"id": catchId, "status": request.Status}))
}

// DeleteLead godoc
// @Summary     Hapus lead
// @Description Soft delete lead berdasarkan ID (admin only)
// @Tags        leads
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Lead ID"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /leads/{id} [delete]
func (h *LeadHandler) DeleteLead(c *gin.Context) {
	leadID := c.Param("id")
	if leadID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID lead tidak boleh kosong"))
		return
	}

	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	err := h.leadService.DeleteLead(leadID, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Lead berhasil dihapus", gin.H{"id": leadID}))
}

// ReplyWhatsApp godoc
// @Summary     Balas pesan WhatsApp
// @Description Kirim balasan WhatsApp ke lead
// @Tags        leads
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string          true "Lead ID"
// @Param       body body SendReplyRequest true "Pesan balasan"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /leads/{id}/reply [post]
func (h *LeadHandler) ReplyWhatsApp(c *gin.Context) {
	idMessage := c.Param("id")
	if idMessage == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID lead tidak boleh kosong"))
		return
	}
	var req SendReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	idSales, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error("unauthorized"))
		return
	}
	idSalesStr := idSales.(string)

	lead, err := h.leadRepository.GetLeadByID(idMessage)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Lead tidak ditemukan"))
		return
	}
	clientHandphone := lead.Phone
	go func() {
		if err := h.waService.SendWA(clientHandphone, req.Message); err != nil {
			logrus.WithError(err).Warn("Gagal kirim pesan WhatsApp")
		}
	}()
	activity, _ := h.activtyService.CreateActivity("WhatsApp Reply", req.Message, lead.ID, idSalesStr, "")
	c.JSON(http.StatusOK, response.Success("Pesan WhatsApp terkirim", activity))
}

// ImportCSV godoc
// @Summary     Import leads dari CSV
// @Description Upload file CSV untuk import data leads
// @Tags        leads
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       file formData file true "File CSV"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /leads/import [post]
func (h *LeadHandler) ImportCSV(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, ImportCSVMaxSize)
	if err := c.Request.ParseMultipartForm(ImportCSVMaxSize); err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, response.Error("File terlalu besar (maks 10MB)"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".csv" {
		c.JSON(http.StatusBadRequest, response.Error("Hanya file CSV yang diizinkan"))
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	defer src.Close()
	reader := csv.NewReader(src)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	var daftarBaru []entity.Lead
	for i, row := range records {
		if i == 0 {
			continue
		}
		lead := entity.Lead{
			Name:       row[0],
			Email:      row[1],
			Phone:      row[2],
			Status:     "NEW",
			AssignedTo: userID,
		}
		daftarBaru = append(daftarBaru, lead)
	}
	err = h.leadService.ImportBulkLeads(daftarBaru)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal menyimpan data lead"))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Sukses import lead dari CSV", gin.H{"total": len(daftarBaru)}))
}

// GetAgingLeads godoc
// @Summary     Lead yang sudah lama tidak diupdate
// @Description Ambil leads yang tidak disentuh lebih dari X hari
// @Tags        leads
// @Security    BearerAuth
// @Produce     json
// @Param       days query int false "Jumlah hari (default 7)"
// @Success     200 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /leads/aging [get]
func (h *LeadHandler) GetAgingLeads(c *gin.Context) {
	days := 7
	if c.Query("days") != "" {
		days, _ = strconv.Atoi(c.Query("days"))
	}
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	leads, err := h.leadService.GetAgingLeads(userID, role, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Daftar lead yang menderita kemarin", leads))
}

// GetLeadTimeline godoc
// @Summary     Timeline aktivitas lead
// @Description Ambil semua aktivitas, deals, dan tasks dari sebuah lead
// @Tags        leads
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Lead ID"
// @Success     200 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /leads/{id}/timeline [get]
func (h *LeadHandler) GetLeadTimeline(c *gin.Context) {
	leadID := c.Param("id")
	if leadID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID lead tidak boleh kosong"))
		return
	}
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	timelines, err := h.leadService.GetLeadTimeline(leadID, userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Timeline lead", timelines))
}

// UpdateLead godoc
// @Summary     Update data lead
// @Description Update nama, email, dan nomor HP lead
// @Tags        leads
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string          true "Lead ID"
// @Param       body body UpdateLeadRequest true "Data lead yang diupdate"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /leads/{id} [put]
func (h *LeadHandler) UpdateLead(c *gin.Context) {
	leadID := c.Param("id")
	if leadID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID lead tidak boleh kosong"))
		return
	}

	var req UpdateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	err := h.leadService.UpdateLead(leadID, req.Name, req.Email, req.Phone, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success("Update lead berhasil", gin.H{"id": leadID, "name": req.Name, "email": req.Email, "phone": req.Phone}))
}
