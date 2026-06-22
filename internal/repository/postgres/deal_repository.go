package postgres

import (
	"context"
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type DealRepository struct {
	dbGorm *gorm.DB
}

func NewDealRepository(db *gorm.DB) *DealRepository {
	return &DealRepository{dbGorm: db}
}

func (r *DealRepository) CreateDeal(deal *entity.Deal) (*entity.Deal, error) {
	isCreate := r.dbGorm.Create(deal)
	err := isCreate.Error
	if err != nil {
		return nil, err
	}
	return deal, nil
}

func (r *DealRepository) GetDealByID(dealID string) (*entity.Deal, error) {
	ctx := context.Background()
	var deal entity.Deal
	err := r.dbGorm.WithContext(ctx).Where("id = ?", dealID).First(&deal).Error
	if err != nil {
		return nil, err
	}
	return &deal, nil
}

func (r *DealRepository) GetAllDeals(page int, limit int, search string, stage string) ([]entity.Deal, int64, error) {
	var deals []entity.Deal
	var total int64

	query := r.dbGorm.Model(&entity.Deal{})
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	if stage != "" {
		query = query.Where("stage = ?", stage)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&deals).Error
	if err != nil {
		return nil, 0, err
	}
	return deals, total, nil
}

func (r *DealRepository) GetDealByUserId(userId string, page int, limit int, search string, stage string) ([]entity.Deal, int64, error) {
	ctx := context.Background()
	var deals []entity.Deal
	var total int64

	query := r.dbGorm.WithContext(ctx).Model(&entity.Deal{}).Where("assigned_to = ?", userId)
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}
	if stage != "" {
		query = query.Where("stage = ?", stage)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&deals).Error
	if err != nil {
		return nil, 0, err
	}
	return deals, total, nil
}

func (r *DealRepository) UpdateStage(dealID string, stage string) error {
	ctx := context.Background()
	var deals entity.Deal
	err := r.dbGorm.WithContext(ctx).Model(&deals).Where("id = ?", dealID).Update("stage", stage).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *DealRepository) SoftDeleteDeal(dealID string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Where("id = ?", dealID).Delete(&entity.Deal{}).Error
	return err
}
