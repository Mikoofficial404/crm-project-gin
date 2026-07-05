package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DealCommentHandler struct {
	dealCommentService *service.DealCommentService
}

func NewDealCommentHandler(dealCommentService *service.DealCommentService) *DealCommentHandler {
	return &DealCommentHandler{dealCommentService: dealCommentService}
}

type AddCommentRequest struct {
	Content          string   `json:"content" binding:"required"`
	MentionedUserIDs []string `json:"mentioned_user_ids"`
}

func (h *DealCommentHandler) AddComment(c *gin.Context) {
	dealID := c.Param("id")
	userID := c.MustGet("user_id").(string)

	var req AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	err := h.dealCommentService.AddComment(dealID, userID, req.Content, req.MentionedUserIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Komentar berhasil ditambahkan", nil))
}

func (h *DealCommentHandler) GetComments(c *gin.Context) {
	dealID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	comments, total, err := h.dealCommentService.GetComments(dealID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Paginated(comments, total, page, limit))
}

func (h *DealCommentHandler) DeleteComment(c *gin.Context) {
	commentID := c.Param("commentId")

	if err := h.dealCommentService.DeleteComment(commentID); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Komentar berhasil dihapus", nil))
}
