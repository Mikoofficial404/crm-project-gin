package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	MaxUploadSize = 1 << 20 // 1 MB
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".pdf":  true,
	".csv":  true,
	".xlsx": true,
	".xls":  true,
	".docx": true,
	".doc":  true,
	".txt":  true,
}

var allowedMimeTypes = map[string]bool{
	"image/jpeg":                              true,
	"image/png":                               true,
	"image/gif":                               true,
	"application/pdf":                         true,
	"text/csv":                                true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"application/vnd.ms-excel":                                           true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/msword":                                                     true,
	"text/plain":                                                             true,
}

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

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, response.Error("Tipe file tidak diizinkan"))
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal membaca file"))
		return
	}
	defer src.Close()

	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal membaca file"))
		return
	}
	mimeType := http.DetectContentType(buffer)
	if !allowedMimeTypes[mimeType] {
		c.JSON(http.StatusBadRequest, response.Error("Tipe file tidak diizinkan"))
		return
	}

	newFileName := fmt.Sprintf("%d%s", time.Now().UnixMilli(), ext)
	dst := filepath.Join("./uploads", newFileName)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal menyimpan file"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Upload berhasil", gin.H{"file_id": newFileName}))
}

// @Summary      Download file
// @Tags         Activities
// @Produce      octet-stream
// @Security     BearerAuth
// @Param        file_id path string true "File ID"
// @Success      200 {file} file
// @Failure      404 {object} response.Response
// @Router       /files/{file_id} [get]
func (h *ActivityHandler) DownloadFile(c *gin.Context) {
	fileID := c.Param("file_id")

	ext := filepath.Ext(fileID)
	if ext == "" || !allowedExtensions[strings.ToLower(ext)] {
		c.JSON(http.StatusBadRequest, response.Error("File tidak valid"))
		return
	}

	cleanName := filepath.Base(fileID)
	filePath := filepath.Join("./uploads", cleanName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, response.Error("File tidak ditemukan"))
		return
	}

	c.File(filePath)
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
