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

func (r *ActivityRepository) GetActivitiesByLeadID(leadId string, page int, limit int) ([]entity.Activity, int, error) {
	ctx := context.Background()
	var activities []entity.Activity
	var total int64

	baseQuery := r.dbGorm.WithContext(ctx).Model(&entity.Activity{}).Where("lead_id = ?", leadId)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := baseQuery.Offset(offset).Limit(limit).Order("created_at DESC").Find(&activities).Error
	if err != nil {
		return nil, 0, err
	}
	return activities, int(total), nil
}

func (r *ActivityRepository) GetActivityByID(id string) (*entity.Activity, error) {
	ctx := context.Background()
	var activity entity.Activity
	err := r.dbGorm.WithContext(ctx).Where("id = ?", id).First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ActivityRepository) UpdateActivity(id string, notes string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Model(&entity.Activity{}).Where("id = ?", id).Update("notes", notes).Error
	return err
}

func (r *ActivityRepository) DeleteActivity(id string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Where("id = ?", id).Delete(&entity.Activity{}).Error
	return err
}
