package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
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
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	userID := c.MustGet("user_id").(string)
	campaign, err := h.campaignService.CreateCampaign(req.Name, req.Subject, req.Body, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Campaign berhasil dibuat", campaign))
}

func (h *CampaignHandler) GetAllCampaigns(c *gin.Context) {
	campaigns, err := h.campaignService.GetAllCampaigns(c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil daftar campaign", campaigns))
}

func (h *CampaignHandler) GetCampaignByID(c *gin.Context) {
	campaign, err := h.campaignService.GetCampaignByID(c.Param("id"))
	if err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil campaign", campaign))
}

func (h *CampaignHandler) UpdateCampaign(c *gin.Context) {
	id := c.Param("id")
	var req UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	userID := c.MustGet("user_id").(string)
	if err := h.campaignService.UpdateCampaign(id, userID, req.Name, req.Subject, req.Body); err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Campaign berhasil diupdate", nil))
}

func (h *CampaignHandler) DeleteCampaign(c *gin.Context) {
	if err := h.campaignService.DeleteCampaign(c.Param("id")); err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Campaign berhasil dihapus", nil))
}

func (h *CampaignHandler) AddRecipients(c *gin.Context) {
	id := c.Param("id")
	var req AddRecipientsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
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
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Penerima berhasil ditambahkan", nil))
}

func (h *CampaignHandler) SendCampaign(c *gin.Context) {
	if err := h.campaignService.SendCampaign(c.Param("id")); err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Campaign sedang dikirim", nil))
}

func (h *CampaignHandler) ScheduleCampaign(c *gin.Context) {
	id := c.Param("id")
	var req ScheduleCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Format scheduled_at tidak valid, gunakan RFC3339 (contoh: 2006-01-02T15:04:05Z)"))
		return
	}
	if err := h.campaignService.ScheduleCampaign(id, scheduledAt); err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Campaign berhasil dijadwalkan", nil))
}

func (h *CampaignHandler) GetCampaignStats(c *gin.Context) {
	stats, err := h.campaignService.GetCampaignStats(c.Param("id"))
	if err != nil {
		if err.Error() == "campaign not found" {
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil statistik campaign", stats))
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
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Paginated(recipients, total, page, limit))
}
