package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,min=0"`
	Unit        string  `json:"unit"`
}

type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Unit        *string  `json:"unit,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	product, err := h.productService.CreateProduct(req.Name, req.Description, req.Unit, req.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Produk berhasil dibuat", product))
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	onlyActive := true
	if activeStr := c.Query("active"); activeStr != "" {
		if parsed, err := strconv.ParseBool(activeStr); err == nil {
			onlyActive = parsed
		}
	}
	products, err := h.productService.GetAllProducts(onlyActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil daftar produk", products))
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	product, err := h.productService.GetProductByID(c.Param("id"))
	if err != nil {
		if err.Error() == "produk tidak ditemukan" {
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil produk", product))
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	err := h.productService.UpdateProduct(c.Param("id"), req.Name, req.Description, req.Unit, req.Price, req.IsActive)
	if err != nil {
		if err.Error() == "produk tidak ditemukan" {
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Produk berhasil diupdate", nil))
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	err := h.productService.DeleteProduct(c.Param("id"))
	if err != nil {
		if err.Error() == "produk tidak ditemukan" {
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Produk berhasil dihapus", nil))
}
