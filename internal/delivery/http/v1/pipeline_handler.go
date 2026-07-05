package v1

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/service"
	"crm-project/pkg/response"
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
	Probability *int    `json:"probability"`
}

type CreateStageRequest struct {
	Name         string `json:"name" binding:"required"`
	Color        string `json:"color"`
	IsClosedWon  bool   `json:"is_closed_won"`
	IsClosedLost bool   `json:"is_closed_lost"`
	Probability  *int   `json:"probability"`
}

type UpdatePipelineStageRequest struct {
	Name         *string `json:"name"`
	Color        *string `json:"color"`
	IsClosedWon  *bool   `json:"is_closed_won"`
	IsClosedLost *bool   `json:"is_closed_lost"`
	Probability  *int    `json:"probability"`
}

type ReorderStagesRequest struct {
	Stages []entity.StageOrder `json:"stages" binding:"required"`
}

func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
	var req CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	userID := c.MustGet("user_id").(string)
	data, err := h.pipelineService.CreatePipeline(req.Name, req.Description, userID, req.IsDefault)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Pipeline berhasil dibuat", data))
}

func (h *PipelineHandler) GetAllPipelines(c *gin.Context) {
	data, err := h.pipelineService.GetAllPipelines()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil daftar pipeline", data))
}

func (h *PipelineHandler) GetPipelineByID(c *gin.Context) {
	data, err := h.pipelineService.GetPipelineByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Pipeline tidak ditemukan"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil pipeline", data))
}

func (h *PipelineHandler) UpdatePipeline(c *gin.Context) {
	pipelineID := c.Param("id")
	var req UpdatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
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
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Pipeline berhasil diupdate", nil))
}

func (h *PipelineHandler) DeletePipeline(c *gin.Context) {
	if err := h.pipelineService.DeletePipeline(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Pipeline berhasil dihapus", nil))
}

func (h *PipelineHandler) AddStage(c *gin.Context) {
	pipelineID := c.Param("id")
	var req CreateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	if req.Probability != nil && (*req.Probability < 0 || *req.Probability > 100) {
		c.JSON(http.StatusBadRequest, response.Error("Probability harus antara 0 dan 100"))
		return
	}
	probability := 0
	if req.Probability != nil {
		probability = *req.Probability
	}
	data, err := h.pipelineService.AddStage(pipelineID, req.Name, req.Color, req.IsClosedWon, req.IsClosedLost, probability)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Stage berhasil ditambahkan", data))
}

func (h *PipelineHandler) UpdateStage(c *gin.Context) {
	stageID := c.Param("stageId")
	var req UpdatePipelineStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	if req.Probability != nil && (*req.Probability < 0 || *req.Probability > 100) {
		c.JSON(http.StatusBadRequest, response.Error("probability harus antara 0-100"))
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
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Stage berhasil diupdate", nil))
}

func (h *PipelineHandler) DeleteStage(c *gin.Context) {
	if err := h.pipelineService.DeleteStage(c.Param("stageId")); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Stage berhasil dihapus", nil))
}

func (h *PipelineHandler) ReorderStages(c *gin.Context) {
	pipelineID := c.Param("id")
	var req ReorderStagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	if err := h.pipelineService.ReorderStages(pipelineID, req.Stages); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Urutan stage berhasil diupdate", nil))
}
