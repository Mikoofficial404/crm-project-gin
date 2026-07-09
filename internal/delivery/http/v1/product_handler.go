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

// @Summary      Buat produk baru
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateProductRequest true "Data produk"
// @Success      201 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /products [post]
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

// @Summary      List semua produk
// @Tags         Products
// @Produce      json
// @Security     BearerAuth
// @Param        active query bool false "Filter produk aktif saja" default(true)
// @Success      200 {object} response.Response
// @Failure      500 {object} response.Response
// @Router       /products [get]
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	var isActive *bool
	if activeStr := c.Query("active"); activeStr != "" {
		if parsed, err := strconv.ParseBool(activeStr); err == nil {
			isActive = &parsed
		}
	}
	products, err := h.productService.GetAllProducts(isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil daftar produk", products))
}

// @Summary      Detail produk
// @Tags         Products
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Product ID"
// @Success      200 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /products/{id} [get]
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

// @Summary      Update produk
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Product ID"
// @Param        body body UpdateProductRequest true "Data update produk"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /products/{id} [patch]
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

// @Summary      Hapus produk
// @Tags         Products
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Product ID"
// @Success      200 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /products/{id} [delete]
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
