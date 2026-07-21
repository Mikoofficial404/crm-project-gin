package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
	"encoding/csv"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ContactHandler struct {
	contactService *service.ContactService
}

func NewContactHandler(contactService *service.ContactService) *ContactHandler {
	return &ContactHandler{contactService: contactService}
}

type CreateContactRequest struct {
	Name     string  `json:"name" binding:"required"`
	Phone    string  `json:"phone" binding:"required"`
	Email    *string `json:"email,omitempty"`
	Company  *string `json:"company,omitempty"`
	Position *string `json:"position,omitempty"`
}

type UpdateContactRequest struct {
	Name       *string `json:"name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Email      *string `json:"email,omitempty"`
	Company    *string `json:"company,omitempty"`
	Position   *string `json:"position,omitempty"`
	AssignedTo *string `json:"assigned_to,omitempty"`
}

// CreateContact godoc
// @Summary     Buat contact baru
// @Description Buat contact baru ke sistem
// @Tags        contacts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body CreateContactRequest true "Data contact"
// @Success     201 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /contacts [post]
func (h *ContactHandler) CreateContact(c *gin.Context) {
	var req CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	userID := c.MustGet("user_id").(string)

	data, err := h.contactService.CreateContact(
		req.Name,
		req.Phone,
		req.Email,
		req.Company,
		req.Position,
		"manual",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success("Contact berhasil dibuat", data))
}

// GetAllContacts godoc
// @Summary     List contacts
// @Description Ambil daftar contacts dengan filter dan pagination
// @Tags        contacts
// @Security    BearerAuth
// @Produce     json
// @Param       page   query int    false "Halaman (default 1)"
// @Param       limit  query int    false "Jumlah per halaman (default 10)"
// @Param       search query string false "Kata kunci pencarian"
// @Param       source query string false "Filter source (whatsapp/manual)"
// @Success     200 {object} response.PaginatedResponse
// @Failure     500 {object} response.Response
// @Router      /contacts [get]
func (h *ContactHandler) GetAllContacts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	source := c.Query("source")

	data, total, err := h.contactService.GetAllContacts(page, limit, search, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Paginated(data, total, page, limit))
}

// GetContactByID godoc
// @Summary     Detail contact
// @Description Ambil detail contact berdasarkan ID
// @Tags        contacts
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Contact ID"
// @Success     200 {object} response.Response
// @Failure     404 {object} response.Response
// @Router      /contacts/{id} [get]
func (h *ContactHandler) GetContactByID(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)

	data, err := h.contactService.GetContactByID(id, userID, role)
	if err != nil {
		if err.Error() == "contact tidak ditemukan" {
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success("Berhasil mengambil contact", data))
}

// UpdateContact godoc
// @Summary     Update contact
// @Description Update data contact secara partial
// @Tags        contacts
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id   path string              true "Contact ID"
// @Param       body body UpdateContactRequest true "Data yang diupdate"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /contacts/{id} [patch]
func (h *ContactHandler) UpdateContact(c *gin.Context) {
	id := c.Param("id")

	var req UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Company != nil {
		updates["company"] = *req.Company
	}
	if req.Position != nil {
		updates["position"] = *req.Position
	}
	if req.AssignedTo != nil {
		updates["assigned_to"] = *req.AssignedTo
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("tidak ada field yang diupdate"))
		return
	}

	if err := h.contactService.UpdateContact(id, updates); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success("Contact berhasil diupdate", nil))
}

// GetTrashedContacts godoc
// @Summary     Contacts yang dihapus
// @Description Ambil daftar contact yang sudah di-soft delete
// @Tags        contacts
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /contacts/trash [get]
func (h *ContactHandler) GetTrashedContacts(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("role")

	contacts, err := h.contactService.GetTrashedContacts(userID, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success("Berhasil mengambil contact yang dihapus", contacts))
}

// RestoreContact godoc
// @Summary     Restore contact
// @Description Restore contact yang sudah di-soft delete
// @Tags        contacts
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Contact ID"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /contacts/{id}/restore [patch]
func (h *ContactHandler) RestoreContact(c *gin.Context) {
	id := c.Param("id")

	if err := h.contactService.RestoreContact(id); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success("Contact berhasil direstore", nil))
}

// ExportContactsCSV godoc
// @Summary     Export contacts ke CSV
// @Description Download daftar contacts dalam format CSV
// @Tags        contacts
// @Security    BearerAuth
// @Produce     text/csv
// @Param       source query string false "Filter source (whatsapp/manual)"
// @Success     200
// @Failure     500 {object} response.Response
// @Router      /contacts/export [get]
func (h *ContactHandler) ExportContactsCSV(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	role := c.MustGet("role").(string)
	source := c.Query("source")

	data, err := h.contactService.ExportContacts(userID, role, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=contacts.csv")

	writer := csv.NewWriter(c.Writer)
	writer.Write([]string{"ID", "Name", "Email", "Phone", "Company", "Position", "Source", "AssignedTo", "CreatedAt"})

	for _, contact := range data {
		email := ""
		if contact.Email != nil {
			email = *contact.Email
		}
		company := ""
		if contact.Company != nil {
			company = *contact.Company
		}
		position := ""
		if contact.Position != nil {
			position = *contact.Position
		}
		writer.Write([]string{
			contact.ID,
			contact.Name,
			email,
			contact.Phone,
			company,
			position,
			contact.Source,
			contact.AssignedTo,
			contact.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	writer.Flush()
}

// ImportContactsCSV godoc
// @Summary     Import contacts dari CSV
// @Description Upload file CSV untuk import contacts secara massal
// @Tags        contacts
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       file formData file true "File CSV"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /contacts/import [post]
func (h *ContactHandler) ImportContactsCSV(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, ImportCSVMaxSize)
	if err := c.Request.ParseMultipartForm(ImportCSVMaxSize); err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, response.Error("File terlalu besar (maks 10MB)"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("file tidak ditemukan"))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".csv" {
		c.JSON(http.StatusBadRequest, response.Error("Hanya file CSV yang diizinkan"))
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("gagal membuka file"))
		return
	}
	defer src.Close()

	reader := csv.NewReader(src)
	reader.Read()
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("gagal membaca CSV"))
		return
	}

	imported, skipped, errors := h.contactService.ImportContactsFromCSV(records, userID)

	c.JSON(http.StatusOK, response.Success("Import selesai", gin.H{
		"imported": imported,
		"skipped":  skipped,
		"errors":   errors,
	}))
}

// DeleteContact godoc
// @Summary     Hapus contact
// @Description Soft delete contact berdasarkan ID
// @Tags        contacts
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "Contact ID"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /contacts/{id} [delete]
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	id := c.Param("id")

	if err := h.contactService.DeleteContact(id); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success("Contact berhasil dihapus", nil))
}
