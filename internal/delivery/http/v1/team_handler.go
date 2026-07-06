package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ManagerID   string `json:"manager_id" binding:"required"`
}

type UpdateTeamRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	ManagerID   *string `json:"manager_id,omitempty"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// CreateTeam godoc
// @Summary     Buat team baru
// @Description Buat team baru (admin only)
// @Tags        teams
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body CreateTeamRequest true "Data team"
// @Success     201 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /teams [post]
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	team, err := h.teamService.CreateTeam(req.Name, req.Description, req.ManagerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Team berhasil dibuat", team))
}

// GetAllTeams godoc
// @Summary     Daftar semua team
// @Description Ambil semua team (admin only)
// @Tags        teams
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /teams [get]
func (h *TeamHandler) GetAllTeams(c *gin.Context) {
	teams, err := h.teamService.GetAllTeams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil daftar team", teams))
}

// GetTeamByID godoc
// @Summary     Detail team
// @Description Ambil detail team berdasarkan ID
// @Tags        teams
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Team ID"
// @Success     200 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /teams/{id} [get]
func (h *TeamHandler) GetTeamByID(c *gin.Context) {
	team, err := h.teamService.GetTeamByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil team", team))
}

// UpdateTeam godoc
// @Summary     Update team
// @Description Update data team berdasarkan ID (admin only)
// @Tags        teams
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string          true "Team ID"
// @Param       body body UpdateTeamRequest true "Data update team"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /teams/{id} [patch]
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	teamID := c.Param("id")
	var req UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	if err := h.teamService.UpdateTeam(teamID, req.Name, req.Description, req.ManagerID); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Team berhasil diupdate", nil))
}

// DeleteTeam godoc
// @Summary     Hapus team
// @Description Hapus team berdasarkan ID (admin only)
// @Tags        teams
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Team ID"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /teams/{id} [delete]
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	if err := h.teamService.DeleteTeam(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Team berhasil dihapus", nil))
}

// AddMember godoc
// @Summary     Tambah member ke team
// @Description Tambah user ke dalam team (admin only)
// @Tags        teams
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string          true "Team ID"
// @Param       body body AddMemberRequest true "Data member"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /teams/{id}/members [post]
func (h *TeamHandler) AddMember(c *gin.Context) {
	teamID := c.Param("id")
	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	if err := h.teamService.AddMember(teamID, req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Member berhasil ditambahkan ke team", nil))
}

// RemoveMember godoc
// @Summary     Hapus member dari team
// @Description Hapus user dari team berdasarkan ID (admin only)
// @Tags        teams
// @Security    BearerAuth
// @Produce     json
// @Param       id     path string true "Team ID"
// @Param       userId path string true "User ID"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /teams/{id}/members/{userId} [delete]
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID := c.Param("id")
	userID := c.Param("userId")
	if err := h.teamService.RemoveMember(teamID, userID); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Member berhasil dihapus dari team", nil))
}
