package postgres

import (
	"context"
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type ActivityRepository struct {
	dbGorm *gorm.DB
}

func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{dbGorm: db}
}

func (r *ActivityRepository) CreateActivity(activity *entity.Activity) (*entity.Activity, error) {
	isCreate := r.dbGorm.Create(activity)
	err := isCreate.Error
	if err != nil {
		return nil, err
	}
	return activity, nil
}

func (r *ActivityRepository) GetActivitiesByLeadID(leadId string) ([]entity.Activity, error) {
	ctx := context.Background()
	var activity []entity.Activity
	err := r.dbGorm.WithContext(ctx).
		Where("lead_id = ?", leadId).
		Find(&activity).Error
	if err != nil {
		return nil, err
	}
	return activity, nil
}
