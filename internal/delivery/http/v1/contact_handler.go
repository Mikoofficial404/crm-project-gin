package v1

import (
	"crm-project/internal/service"
	"net/http"
	"strconv"

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

func (h *ContactHandler) CreateContact(c *gin.Context) {
	var req CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "Contact berhasil dibuat", "data": data})
}

func (h *ContactHandler) GetAllContacts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	source := c.Query("source")

	data, total, err := h.contactService.GetAllContacts(page, limit, search, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  data,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ContactHandler) GetContactByID(c *gin.Context) {
	id := c.Param("id")

	data, err := h.contactService.GetContactByID(id)
	if err != nil {
		if err.Error() == "contact tidak ditemukan" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *ContactHandler) UpdateContact(c *gin.Context) {
	id := c.Param("id")

	var req UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "tidak ada field yang diupdate"})
		return
	}

	if err := h.contactService.UpdateContact(id, updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Contact berhasil diupdate"})
}

func (h *ContactHandler) DeleteContact(c *gin.Context) {
	id := c.Param("id")

	if err := h.contactService.DeleteContact(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Contact berhasil dihapus"})
}
