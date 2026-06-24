package postgres

import (
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type InvoiceRepository struct {
	dbGorm *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{dbGorm: db}
}

func (r *InvoiceRepository) CreateInvoice(invoice *entity.Invoice) (*entity.Invoice, error) {
	isCreate := r.dbGorm.Create(invoice)
	err := isCreate.Error
	if err != nil {
		return nil, err
	}
	return invoice, nil
}

func (r *InvoiceRepository) GetInvoiceByDealID(dealID string) (*entity.Invoice, error) {
	var invoice *entity.Invoice
	result := r.dbGorm.Where("deal_id = ?", dealID).First(&invoice)
	err := result.Error
	if err != nil {
		return nil, err
	}
	return invoice, err
}
