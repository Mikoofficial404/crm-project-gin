package v1

import (
	"crm-project/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	activityService *service.ActivityService
}

type CreateActivity struct {
	Type   string `json:"type" binding:"required"`
	Notes  string `json:"notes" binding:"required"`
	LeadID string `json:"leadID" binding:"required"`
}

func NewAcitivyHandler(activityService *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		activityService: activityService,
	}
}

func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req CreateActivity
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("user_id").(string)
	data, err := h.activityService.CreateActivity(req.Type, req.Notes, req.LeadID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "Activity Created", "data": data})
}

func (h *ActivityHandler) GetActivities(c *gin.Context) {
	leadId := c.Param("lead_id")
	if leadId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID deal tidak boleh kosong"})
		return
	}
	role := c.MustGet("role").(string)
	data, err := h.activityService.GetActivitiesByLeadID(leadId, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Get Activites", "data": data})
}
