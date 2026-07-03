package v1

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PipelineHandler struct {
	pipelineService *service.PipelineService
}

func NewPipelineHandler(pipelineService *service.PipelineService) *PipelineHandler {
	return &PipelineHandler{pipelineService: pipelineService}
}

type CreatePipelineRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

type UpdatePipelineRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
	IsDefault   *bool   `json:"is_default"`
}

type CreateStageRequest struct {
	Name         string `json:"name" binding:"required"`
	Color        string `json:"color"`
	IsClosedWon  bool   `json:"is_closed_won"`
	IsClosedLost bool   `json:"is_closed_lost"`
}

type UpdatePipelineStageRequest struct {
	Name         *string `json:"name"`
	Color        *string `json:"color"`
	IsClosedWon  *bool   `json:"is_closed_won"`
	IsClosedLost *bool   `json:"is_closed_lost"`
}

type ReorderStagesRequest struct {
	Stages []entity.StageOrder `json:"stages" binding:"required"`
}

func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
	var req CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.MustGet("user_id").(string)
	data, err := h.pipelineService.CreatePipeline(req.Name, req.Description, userID, req.IsDefault)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "Pipeline berhasil dibuat", "data": data})
}

func (h *PipelineHandler) GetAllPipelines(c *gin.Context) {
	data, err := h.pipelineService.GetAllPipelines()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *PipelineHandler) GetPipelineByID(c *gin.Context) {
	pipelineID := c.Param("id")
	data, err := h.pipelineService.GetPipelineByID(pipelineID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *PipelineHandler) UpdatePipeline(c *gin.Context) {
	pipelineID := c.Param("id")
	var req UpdatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	setDefault := req.IsDefault != nil && *req.IsDefault

	if err := h.pipelineService.UpdatePipeline(pipelineID, updates, setDefault); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Pipeline berhasil diupdate"})
}

func (h *PipelineHandler) DeletePipeline(c *gin.Context) {
	pipelineID := c.Param("id")
	if err := h.pipelineService.DeletePipeline(pipelineID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Pipeline berhasil dihapus"})
}

func (h *PipelineHandler) AddStage(c *gin.Context) {
	pipelineID := c.Param("id")
	var req CreateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := h.pipelineService.AddStage(pipelineID, req.Name, req.Color, req.IsClosedWon, req.IsClosedLost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "Stage berhasil ditambahkan", "data": data})
}

func (h *PipelineHandler) UpdateStage(c *gin.Context) {
	stageID := c.Param("stageId")
	var req UpdatePipelineStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.IsClosedWon != nil {
		updates["is_closed_won"] = *req.IsClosedWon
	}
	if req.IsClosedLost != nil {
		updates["is_closed_lost"] = *req.IsClosedLost
	}

	if err := h.pipelineService.UpdateStage(stageID, updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Stage berhasil diupdate"})
}

func (h *PipelineHandler) DeleteStage(c *gin.Context) {
	stageID := c.Param("stageId")
	if err := h.pipelineService.DeleteStage(stageID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Stage berhasil dihapus"})
}

func (h *PipelineHandler) ReorderStages(c *gin.Context) {
	pipelineID := c.Param("id")
	var req ReorderStagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.pipelineService.ReorderStages(pipelineID, req.Stages); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Urutan stage berhasil diupdate"})
}
