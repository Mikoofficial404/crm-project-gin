package v1

import (
	"crm-project/internal/service"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.CreateProduct(req.Name, req.Description, req.Unit, req.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Produk berhasil dibuat", "data": product})
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	onlyActive := true
	if activeStr := c.Query("active"); activeStr != "" {
		parsed, err := strconv.ParseBool(activeStr)
		if err == nil {
			onlyActive = parsed
		}
	}

	products, err := h.productService.GetAllProducts(onlyActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil daftar produk", "data": products})
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	productID := c.Param("id")

	product, err := h.productService.GetProductByID(productID)
	if err != nil {
		if err.Error() == "produk tidak ditemukan" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil produk", "data": product})
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.productService.UpdateProduct(productID, req.Name, req.Description, req.Unit, req.Price, req.IsActive)
	if err != nil {
		if err.Error() == "produk tidak ditemukan" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil diupdate"})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")

	err := h.productService.DeleteProduct(productID)
	if err != nil {
		if err.Error() == "produk tidak ditemukan" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil dihapus"})
}
