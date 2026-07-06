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
	LAHours     *int   `json:"la_hours"`
}

type UpdatePipelineRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
	IsDefault   *bool   `json:"is_default"`
	Probability *int    `json:"probability"`
	SLAHours    *int    `json:"sla_hours"`
}

type CreateStageRequest struct {
	Name         string `json:"name" binding:"required"`
	Color        string `json:"color"`
	IsClosedWon  bool   `json:"is_closed_won"`
	IsClosedLost bool   `json:"is_closed_lost"`
	Probability  *int   `json:"probability"`
	SLAHours     int    `json:"sla_hours"`
}

type UpdatePipelineStageRequest struct {
	Name         *string `json:"name"`
	Color        *string `json:"color"`
	IsClosedWon  *bool   `json:"is_closed_won"`
	IsClosedLost *bool   `json:"is_closed_lost"`
	Probability  *int    `json:"probability"`
	SLAHours     *int    `json:"sla_hours"`
}

type ReorderStagesRequest struct {
	Stages []entity.StageOrder `json:"stages" binding:"required"`
}

// @Summary      Buat pipeline baru
// @Tags         Pipelines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreatePipelineRequest true "Data pipeline"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /pipelines [post]
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

// @Summary      List semua pipeline
// @Tags         Pipelines
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /pipelines [get]
func (h *PipelineHandler) GetAllPipelines(c *gin.Context) {
	data, err := h.pipelineService.GetAllPipelines()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil daftar pipeline", data))
}

// @Summary      Detail pipeline
// @Tags         Pipelines
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Pipeline ID"
// @Success      200 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /pipelines/{id} [get]
func (h *PipelineHandler) GetPipelineByID(c *gin.Context) {
	data, err := h.pipelineService.GetPipelineByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Pipeline tidak ditemukan"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil pipeline", data))
}

// @Summary      Update pipeline
// @Tags         Pipelines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Pipeline ID"
// @Param        body body UpdatePipelineRequest true "Data update pipeline"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /pipelines/{id} [patch]
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

// @Summary      Hapus pipeline
// @Tags         Pipelines
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Pipeline ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /pipelines/{id} [delete]
func (h *PipelineHandler) DeletePipeline(c *gin.Context) {
	if err := h.pipelineService.DeletePipeline(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Pipeline berhasil dihapus", nil))
}

// @Summary      Tambah stage ke pipeline
// @Tags         Pipelines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Pipeline ID"
// @Param        body body CreateStageRequest true "Data stage"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /pipelines/{id}/stages [post]
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
	data, err := h.pipelineService.AddStage(pipelineID, req.Name, req.Color, req.IsClosedWon, req.IsClosedLost, probability, req.SLAHours)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Stage berhasil ditambahkan", data))
}

// @Summary      Update stage
// @Tags         Pipelines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Pipeline ID"
// @Param        stageId path string true "Stage ID"
// @Param        body body UpdatePipelineStageRequest true "Data update stage"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /pipelines/{id}/stages/{stageId} [patch]
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
	if req.Probability != nil {
		updates["probability"] = *req.Probability
	}
	if req.SLAHours != nil {
		updates["sla_hours"] = *req.SLAHours
	}

	if err := h.pipelineService.UpdateStage(stageID, updates); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Stage berhasil diupdate", nil))
}

// @Summary      Hapus stage
// @Tags         Pipelines
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Pipeline ID"
// @Param        stageId path string true "Stage ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /pipelines/{id}/stages/{stageId} [delete]
func (h *PipelineHandler) DeleteStage(c *gin.Context) {
	if err := h.pipelineService.DeleteStage(c.Param("stageId")); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Stage berhasil dihapus", nil))
}

// @Summary      Reorder stages
// @Tags         Pipelines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Pipeline ID"
// @Param        body body ReorderStagesRequest true "Urutan stage baru"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /pipelines/{id}/stages/reorder [patch]
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
