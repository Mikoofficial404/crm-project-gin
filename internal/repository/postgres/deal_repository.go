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

func (r *DealRepository) GetAllDeals() ([]entity.Deal, error) {
	var deals []entity.Deal
	err := r.dbGorm.Find(&deals).Error
	if err != nil {
		return nil, err
	}
	return deals, nil
}

func (r *DealRepository) GetDealByUserId(userId string) ([]entity.Deal, error) {
	ctx := context.Background()
	var deals []entity.Deal
	err := r.dbGorm.WithContext(ctx).
		Where("assigned_to = ?", userId).
		Find(&deals).Error
	if err != nil {
		return nil, err
	}
	return deals, nil
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
