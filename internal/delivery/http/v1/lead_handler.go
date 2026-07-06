package v1

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"crm-project/internal/service"
	"crm-project/pkg/response"
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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

func (h *LeadHandler) GetLeadByID(c *gin.Context) {
	leadID := c.Param("id")
	if leadID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID lead tidak boleh kosong"))
		return
	}

	data, err := h.leadService.GetLeadByID(leadID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Lead tidak ditemukan"))
		return
	}

	c.JSON(http.StatusOK, response.Success("Berhasil mengambil lead", data))
}

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

func (h *LeadHandler) ImportCSV(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
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
