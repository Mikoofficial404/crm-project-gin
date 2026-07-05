package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"errors"
	"fmt"
)

type ActivityService struct {
	activity *postgres.ActivityRepository
}

func NewActivityService(activityRepo *postgres.ActivityRepository) *ActivityService {
	return &ActivityService{
		activity: activityRepo,
	}
}

func (s *ActivityService) CreateActivity(types string, notes string, leadID string, AssignedTo string, attachment string) (*entity.Activity, error) {
	activity := entity.Activity{
		Type:       types,
		Notes:      notes,
		LeadID:     leadID,
		AssignedTo: AssignedTo,
		Attachment: attachment,
	}

	result, err := s.activity.CreateActivity(&activity)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ActivityService) GetActivitiesByLeadID(leadID string, role string, page int, limit int) ([]entity.Activity, int, error) {
	if leadID == "" || role == "" {
		return nil, 0, errors.New("semua field wajib diisi")
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	activities, total, err := s.activity.GetActivitiesByLeadID(leadID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("aktivitas tidak ditemukan: %w", err)
	}

	return activities, total, nil
}

func (s *ActivityService) UpdateActivity(activityID string, notes string, userID string, role string) error {
	if activityID == "" {
		return errors.New("ID aktivitas wajib diisi")
	}

	activity, err := s.activity.GetActivityByID(activityID)
	if err != nil {
		return fmt.Errorf("aktivitas tidak ditemukan: %w", err)
	}

	if role == "sales" {
		if activity.AssignedTo != userID {
			return errors.New("unauthorized: ini bukan aktivitas Anda")
		}
	}

	return s.activity.UpdateActivity(activityID, notes)
}

func (s *ActivityService) DeleteActivity(activityID string, userID string, role string) error {
	if activityID == "" {
		return errors.New("ID aktivitas wajib diisi")
	}

	activity, err := s.activity.GetActivityByID(activityID)
	if err != nil {
		return fmt.Errorf("aktivitas tidak ditemukan: %w", err)
	}

	if role == "sales" {
		if activity.AssignedTo != userID {
			return errors.New("unauthorized: ini bukan aktivitas Anda")
		}
	}

	return s.activity.DeleteActivity(activityID)
}
