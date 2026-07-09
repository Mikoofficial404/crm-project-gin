package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"errors"

	"github.com/google/uuid"
)

type NotificationService struct {
	notificationRepo *postgres.NotificationRepository
}

func NewNotificationService(notifRepo *postgres.NotificationRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notifRepo,
	}
}

func (s *NotificationService) CreateNotification(userID string, title string, message string, link string) error {
	parseUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user id")
	}

	notif := &entity.Notification{
		UserID:  parseUUID,
		Title:   title,
		Message: message,
		Link:    link,
		IsRead:  false,
	}
	_, err = s.notificationRepo.CreateNotification(notif)
	return err
}

func (s *NotificationService) GetUnreadByUserID(userID string) ([]entity.Notification, error) {
	return s.notificationRepo.GetUnreadNotifications(userID)
}

func (s *NotificationService) MarkAsRead(userID string, notifID string) error {
	notification, err := s.notificationRepo.GetNotificationByID(notifID)
	if err != nil {
		return errors.New("notifikasi tidak ditemukan")
	}
	if notification.UserID.String() != userID {
		return errors.New("akses ditolak: notifikasi bukan milik Anda")
	}
	return s.notificationRepo.MarkAsRead(notifID)
}

func (s *NotificationService) MarkAllAsRead(userID string) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}
