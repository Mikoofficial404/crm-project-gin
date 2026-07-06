package postgres

import (
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type LoginHistoryRepository struct {
	gormDB *gorm.DB
}

func NewLoginHistoryRepository(db *gorm.DB) *LoginHistoryRepository {
	return &LoginHistoryRepository{gormDB: db}
}

func (r *LoginHistoryRepository) CreateLoginHistory(history *entity.LoginHistory) error {
	if err := r.gormDB.Create(history).Error; err != nil {
		return err
	}
	return nil
}

func (r *LoginHistoryRepository) GetLoginHistoriesByUserID(userID string, limit int) ([]entity.LoginHistory, error) {
	var histories []entity.LoginHistory
	if err := r.gormDB.Where("user_id = ?", userID).Order("login_at DESC").Limit(limit).Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}
