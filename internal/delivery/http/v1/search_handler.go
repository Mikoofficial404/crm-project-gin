package v1

import (
	"crm-project/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchServices service.SearchService
}

func NewServiceHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchServices: *searchService,
	}
}

func (h *SearchHandler) GlobalSearch(c *gin.Context) {
	keyword := c.Query("q")
	data, err := h.searchServices.GlobalSearch(keyword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}
