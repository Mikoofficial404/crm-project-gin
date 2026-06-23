package postgres

import (
	"context"
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type LeadRepository struct {
	dbGorm *gorm.DB
}

func NewLeadRepository(db *gorm.DB) *LeadRepository {
	return &LeadRepository{dbGorm: db}
}

func (r *LeadRepository) CreateLead(lead *entity.Lead) (*entity.Lead, error) {
	isCreate := r.dbGorm.Create(lead)
	err := isCreate.Error
	if err != nil {
		return nil, err
	}
	return lead, nil
}

func (r *LeadRepository) GetAllLeads(page int, limit int, search string, status string) ([]entity.Lead, int64, error) {
	var leads []entity.Lead
	var total int64

	query := r.dbGorm.Model(&entity.Lead{})
	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&leads).Error
	if err != nil {
		return nil, 0, err
	}
	return leads, total, nil
}

func (r *LeadRepository) GetLeadsByUserId(userID string, page int, limit int, search string, status string) ([]entity.Lead, int64, error) {
	ctx := context.Background()
	var leads []entity.Lead
	var total int64

	query := r.dbGorm.WithContext(ctx).Model(&entity.Lead{}).Where("assigned_to = ?", userID)
	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&leads).Error
	if err != nil {
		return nil, 0, err
	}
	return leads, total, nil
}

func (r *LeadRepository) UpdateStatus(leadID string, status string) error {
	ctx := context.Background()
	var leads entity.Lead
	err := r.dbGorm.WithContext(ctx).Model(&leads).Where("id = ?", leadID).Update("status", status).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *LeadRepository) GetLeadByID(leadID string) (*entity.Lead, error) {
	ctx := context.Background()
	var lead entity.Lead
	err := r.dbGorm.WithContext(ctx).Where("id = ?", leadID).First(&lead).Error
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

func (r *LeadRepository) SoftDeleteLead(leadID string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Where("id = ?", leadID).Delete(&entity.Lead{}).Error
	return err
}

func (r *LeadRepository) CreateBulkLeads(leadEntity *[]entity.Lead) ([]entity.Lead, error) {
	csvInsert := r.dbGorm.Create(&leadEntity)
	err := csvInsert.Error
	if err != nil {
		return nil, err
	}
	return *leadEntity, nil
}
