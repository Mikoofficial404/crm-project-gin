package postgres

import (
	"context"
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	dbGorm *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{dbGorm: db}
}

func (r *NotificationRepository) CreateNotification(notification *entity.Notification) (*entity.Notification, error) {
	err := r.dbGorm.Create(notification).Error
	if err != nil {
		return nil, err
	}
	return notification, nil
}

func (r *NotificationRepository) GetNotificationByID(notifID string) (entity.Notification, error) {
	ctx := context.Background()
	var notification entity.Notification
	err := r.dbGorm.WithContext(ctx).
		Where("id = ?", notifID).
		First(&notification).Error

	if err != nil {
		return entity.Notification{}, err
	}
	return notification, nil
}

func (r *NotificationRepository) GetUnreadNotifications(userID string) ([]entity.Notification, error) {
	ctx := context.Background()
	var notifications []entity.Notification
	err := r.dbGorm.WithContext(ctx).
		Where("user_id = ? AND is_read = ?", userID, false).
		Order("created_at DESC").
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepository) MarkAsRead(notifID string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("id = ?", notifID).
		Update("is_read", true).Error
	return err
}
