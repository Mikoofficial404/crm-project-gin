package v1

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/service"
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LeadHandler struct {
	leadService *service.LeadService
}

func NewLeadHandler(leadService *service.LeadService) *LeadHandler {
	return &LeadHandler{
		leadService: leadService,
	}
}

type UpdateStatusReques struct {
	Status string `json:"status" binding:"required"`
}

type CreateLeadRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Phone string `json:"phone" binding:"required"`
}

func (h *LeadHandler) CreateLeader(c *gin.Context) {
	var req CreateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("user_id").(string)
	data, err := h.leadService.CreateLead(req.Name, req.Email, req.Phone, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "Lead Created", "data": data})
}

func (h *LeadHandler) GetLeads(c *gin.Context) {
	userId := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	status := c.Query("status")
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	data, total, err := h.leadService.GetLeads(userId, role, page, limit, search, status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "Get Leads",
		"data":   data,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *LeadHandler) UpdateStatusLeads(c *gin.Context) {
	catchId := c.Param("id")
	var request struct {
		Status string `json:"status"`
	}
	if catchId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID prospek tidak boleh kosong",
		})
		return
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	err := h.leadService.UpdateLeadStatus(catchId, request.Status, userId, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "update status Leads", "data": gin.H{"id": catchId, "status": request.Status}})
}

func (h *LeadHandler) DeleteLead(c *gin.Context) {
	leadID := c.Param("id")
	if leadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID lead tidak boleh kosong"})
		return
	}

	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	err := h.leadService.DeleteLead(leadID, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Lead berhasil dihapus", "data": gin.H{"id": leadID}})
}

func (h *LeadHandler) ImportCSV(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()
	reader := csv.NewReader(src)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan ratusan data"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "Sukses Import Ratusan Lead!"})
}
