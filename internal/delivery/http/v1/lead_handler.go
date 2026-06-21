package v1

import (
	"crm-project/internal/service"
	"net/http"

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
	data, err := h.leadService.GetLeads(userId, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Get Leads", "data": data})
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
