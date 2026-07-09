package postgres

import (
	"context"
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
	var invoice entity.Invoice
	err := r.dbGorm.
		Preload("Items").
		Preload("Items.Product").
		Preload("Deal").
		Preload("Deal.Lead").
		Where("deal_id = ?", dealID).
		First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *InvoiceRepository) GetInvoiceByID(invoiceID string) (*entity.Invoice, error) {
	var invoice entity.Invoice
	err := r.dbGorm.
		Preload("Items").
		Preload("Items.Product").
		Preload("Deal").
		Preload("Deal.Lead").
		Where("id = ?", invoiceID).
		First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *InvoiceRepository) UpdateStatus(invoiceID string, status string) (*entity.Invoice, error) {
	ctx := context.Background()
	var invoice entity.Invoice
	err := r.dbGorm.WithContext(ctx).Model(&invoice).Where("id = ?", invoiceID).Update("status", status).Error
	if err != nil {
		return nil, err
	}
	return &invoice, err
}

func (r *InvoiceRepository) GetAllInvoices() ([]entity.Invoice, error) {
	var invoices []entity.Invoice
	err := r.dbGorm.
		Preload("Items").
		Preload("Items.Product").
		Preload("Deal").
		Preload("Deal.Lead").
		Order("created_at desc").
		Find(&invoices).Error
	if err != nil {
		return nil, err
	}
	return invoices, nil
}
