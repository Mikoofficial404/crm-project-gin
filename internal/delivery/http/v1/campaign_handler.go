package v1

import (
	"crm-project/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CampaignHandler struct {
	campaignService *service.CampaignService
}

func NewCampaignHandler(campaignService *service.CampaignService) *CampaignHandler {
	return &CampaignHandler{campaignService: campaignService}
}

type CreateCampaignRequest struct {
	Name    string `json:"name" binding:"required"`
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body" binding:"required"`
}

type UpdateCampaignRequest struct {
	Name    *string `json:"name"`
	Subject *string `json:"subject"`
	Body    *string `json:"body"`
}

type AddRecipientsRequest struct {
	Type string   `json:"type" binding:"required,oneof=leads contacts"`
	IDs  []string `json:"ids" binding:"required,min=1"`
}

type ScheduleCampaignRequest struct {
	ScheduledAt string `json:"scheduled_at" binding:"required"`
}

func (h *CampaignHandler) CreateCampaign(c *gin.Context) {
	var req CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	campaign, err := h.campaignService.CreateCampaign(req.Name, req.Subject, req.Body, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Campaign created", "data": campaign})
}

func (h *CampaignHandler) GetAllCampaigns(c *gin.Context) {
	status := c.Query("status")
	campaigns, err := h.campaignService.GetAllCampaigns(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Campaigns retrieved", "data": campaigns})
}

func (h *CampaignHandler) GetCampaignByID(c *gin.Context) {
	id := c.Param("id")
	campaign, err := h.campaignService.GetCampaignByID(id)
	if err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Campaign retrieved", "data": campaign})
}

func (h *CampaignHandler) UpdateCampaign(c *gin.Context) {
	id := c.Param("id")
	var req UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	if err := h.campaignService.UpdateCampaign(id, userID, req.Name, req.Subject, req.Body); err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Campaign updated"})
}

func (h *CampaignHandler) DeleteCampaign(c *gin.Context) {
	id := c.Param("id")
	if err := h.campaignService.DeleteCampaign(id); err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Campaign deleted"})
}

func (h *CampaignHandler) AddRecipients(c *gin.Context) {
	id := c.Param("id")
	var req AddRecipientsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	if req.Type == "leads" {
		err = h.campaignService.AddRecipientsByLeadIDs(id, req.IDs)
	} else {
		err = h.campaignService.AddRecipientsByContactIDs(id, req.IDs)
	}

	if err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recipients added"})
}

func (h *CampaignHandler) SendCampaign(c *gin.Context) {
	id := c.Param("id")
	if err := h.campaignService.SendCampaign(id); err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Campaign is now sending"})
}

func (h *CampaignHandler) ScheduleCampaign(c *gin.Context) {
	id := c.Param("id")
	var req ScheduleCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scheduled_at format, use RFC3339 (e.g. 2006-01-02T15:04:05Z)"})
		return
	}

	if err := h.campaignService.ScheduleCampaign(id, scheduledAt); err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Campaign scheduled"})
}

func (h *CampaignHandler) GetCampaignStats(c *gin.Context) {
	id := c.Param("id")
	stats, err := h.campaignService.GetCampaignStats(id)
	if err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Campaign stats retrieved", "data": stats})
}

func (h *CampaignHandler) GetRecipients(c *gin.Context) {
	id := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	recipients, total, err := h.campaignService.GetRecipients(id, page, limit)
	if err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipients retrieved",
		"data":    recipients,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}
