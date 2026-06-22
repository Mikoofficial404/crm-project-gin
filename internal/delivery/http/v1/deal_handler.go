package v1

import (
	"crm-project/internal/service"
	"net/http"

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

type UpdateStageRequest struct {
	Stage string `json:"stage" binding:"required"`
}

func NewDealHandler(dealService *service.DealService) *DealHandler {
	return &DealHandler{
		dealService: dealService,
	}
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
	data, err := h.dealService.GetDeals(userId, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Get Deals", "data": data})
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
