package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	MaxUploadSize = 1 << 20 // 1 MB
)

type ActivityHandler struct {
	activityService *service.ActivityService
}

type CreateActivity struct {
	Type       string `json:"type" binding:"required"`
	Notes      string `json:"notes" binding:"required"`
	LeadID     string `json:"leadID" binding:"required"`
	Attachment string `json:"attachment"`
}

type UpdateActivityRequest struct {
	Notes string `json:"notes" binding:"required"`
}

func NewAcitivyHandler(activityService *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		activityService: activityService,
	}
}

// @Summary      Buat aktivitas baru
// @Tags         Activities
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateActivity true "Data aktivitas"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /activities [post]
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req CreateActivity
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	userID := c.MustGet("user_id").(string)
	data, err := h.activityService.CreateActivity(req.Type, req.Notes, req.LeadID, userID, req.Attachment)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Activity berhasil dibuat", data))
}

// @Summary      List aktivitas by lead
// @Tags         Activities
// @Produce      json
// @Security     BearerAuth
// @Param        lead_id path string true "Lead ID"
// @Param        page query int false "Halaman" default(1)
// @Param        limit query int false "Limit per halaman" default(10)
// @Success      200 {object} response.PaginatedResponse
// @Failure      400 {object} response.Response
// @Router       /activities/{lead_id} [get]
func (h *ActivityHandler) GetActivities(c *gin.Context) {
	leadId := c.Param("lead_id")
	if leadId == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID lead tidak boleh kosong"))
		return
	}
	role := c.MustGet("role").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, total, err := h.activityService.GetActivitiesByLeadID(leadId, role, page, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Paginated(data, int64(total), page, limit))
}

// @Summary      Upload file attachment
// @Tags         Activities
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "File (maks 1MB)"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /upload [post]
func (h *ActivityHandler) UploadFile(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxUploadSize)
	if err := c.Request.ParseMultipartForm(MaxUploadSize); err != nil {
		if _, ok := err.(*http.MaxBytesError); ok {
			c.JSON(http.StatusRequestEntityTooLarge, response.Error(fmt.Sprintf("file terlalu besar (maks: %d bytes)", MaxUploadSize)))
			return
		}
		c.JSON(http.StatusBadGateway, response.Error(err.Error()))
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("file form diperlukan"))
		return
	}

	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%d%s", time.Now().Unix(), ext)
	dst := filepath.Join("./uploads", newFileName)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal menyimpan file"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Upload berhasil", gin.H{"url": "/uploads/" + newFileName}))
}

// @Summary      Update aktivitas
// @Tags         Activities
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Activity ID"
// @Param        body body UpdateActivityRequest true "Data update aktivitas"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /activities/{id} [put]
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID aktivitas tidak boleh kosong"))
		return
	}

	var req UpdateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	err := h.activityService.UpdateActivity(activityID, req.Notes, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success("Aktivitas berhasil diupdate", gin.H{"id": activityID, "notes": req.Notes}))
}

// @Summary      Hapus aktivitas
// @Tags         Activities
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Activity ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /activities/{id} [delete]
func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	activityID := c.Param("id")
	if activityID == "" {
		c.JSON(http.StatusBadRequest, response.Error("ID aktivitas tidak boleh kosong"))
		return
	}

	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	err := h.activityService.DeleteActivity(activityID, userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success("Aktivitas berhasil dihapus", gin.H{"id": activityID}))
}
