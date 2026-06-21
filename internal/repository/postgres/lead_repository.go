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

func (r *LeadRepository) GetAllLeads() ([]entity.Lead, error) {
	var leads []entity.Lead
	err := r.dbGorm.Find(&leads).Error
	if err != nil {
		return nil, err
	}
	return leads, nil
}

func (r *LeadRepository) GetLeadsByUserId(userID string) ([]entity.Lead, error) {
	ctx := context.Background()
	var leads []entity.Lead
	err := r.dbGorm.WithContext(ctx).
		Where("assigned_to = ?", userID).
		Find(&leads).Error
	if err != nil {
		return nil, err
	}
	return leads, nil
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
