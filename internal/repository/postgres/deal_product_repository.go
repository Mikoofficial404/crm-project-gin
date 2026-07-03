package postgres

import (
	"crm-project/internal/models/entity"
	"errors"

	"gorm.io/gorm"
)

type DealProductRepository struct {
	dbgorm *gorm.DB
}

func NewDealProductRepository(db *gorm.DB) *DealProductRepository {
	return &DealProductRepository{dbgorm: db}
}

func (r *DealProductRepository) AssignProduct(dealProduct *entity.DealProduct) (*entity.DealProduct, error) {

	var existing entity.DealProduct
	err := r.dbgorm.Where("deal_id = ? AND product_id = ?", dealProduct.DealID, dealProduct.ProductID).First(&existing).Error
	if err == nil {
		return nil, errors.New("produk sudah di-assign ke deal ini")
	}

	if err := r.dbgorm.Create(dealProduct).Error; err != nil {
		return nil, err
	}
	return dealProduct, nil
}

func (r *DealProductRepository) RemoveProduct(dealID, productID string) error {
	result := r.dbgorm.Where("deal_id = ? AND product_id = ?", dealID, productID).Delete(&entity.DealProduct{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("produk tidak ditemukan di deal ini")
	}
	return nil
}

func (r *DealProductRepository) GetProductsByDealID(dealID string) ([]entity.DealProduct, error) {
	var dealProducts []entity.DealProduct
	err := r.dbgorm.
		Preload("Product").
		Where("deal_id = ?", dealID).
		Find(&dealProducts).Error
	return dealProducts, err
}
