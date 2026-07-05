package postgres

import (
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type DealHistoryRepository struct {
	gormDb *gorm.DB
}

func NewDealHistoryRepository(gormDb *gorm.DB) *DealHistoryRepository {
	return &DealHistoryRepository{gormDb: gormDb}
}

func (r *DealHistoryRepository) CreateHistory(dealID, userID, field, oldValue, newValue string) (*entity.DealHistory, error) {
	history := entity.DealHistory{
		DealID:       dealID,
		UserID:       userID,
		FieldChanged: field,
		OldValue:     oldValue,
		NewValue:     newValue,
	}
	if err := r.gormDb.Create(&history).Error; err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *DealHistoryRepository) GetHistoriesByDealID(dealID string) ([]entity.DealHistory, error) {
	var histories []entity.DealHistory
	err := r.gormDb.
		Preload("User").
		Where("deal_id = ?", dealID).
		Order("changed_at ASC").
		Find(&histories).Error
	return histories, err
}
