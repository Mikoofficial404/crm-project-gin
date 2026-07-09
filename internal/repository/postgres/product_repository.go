package postgres

import (
	"crm-project/internal/models/entity"
	"errors"

	"gorm.io/gorm"
)

type ProductRepository struct {
	dbgorm *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{dbgorm: db}
}

func (r *ProductRepository) CreateProduct(product *entity.Product) (*entity.Product, error) {
	err := r.dbgorm.Create(product).Error
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) GetProductByID(productID string) (*entity.Product, error) {
	var product entity.Product
	err := r.dbgorm.Where("id = ?", productID).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("produk tidak ditemukan")
	}
	return &product, err
}

func (r *ProductRepository) GetAllProducts(isActive *bool) ([]entity.Product, error) {
	var products []entity.Product
	query := r.dbgorm.Where("deleted_at IS NULL")
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	err := query.Order("name ASC").Find(&products).Error
	return products, err
}

func (r *ProductRepository) UpdateProduct(productID string, updates map[string]interface{}) error {
	return r.dbgorm.Model(&entity.Product{}).Where("id = ?", productID).Updates(updates).Error
}

func (r *ProductRepository) DeleteProduct(productID string) error {
	return r.dbgorm.Where("id = ?", productID).Delete(&entity.Product{}).Error
}
