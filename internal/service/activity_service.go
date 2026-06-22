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

func (s *ActivityService) CreateActivity(types string, notes string, leadID string, AssignedTo string) (*entity.Activity, error) {
	activity := entity.Activity{
		Type:       types,
		Notes:      notes,
		LeadID:     leadID,
		AssignedTo: AssignedTo,
	}

	result, err := s.activity.CreateActivity(&activity)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ActivityService) GetActivitiesByLeadID(leadID string, role string) ([]entity.Activity, error) {
	if leadID == "" || role == "" {
		return nil, errors.New("semua field wajib diisi")
	}

	load, err := s.activity.GetActivitiesByLeadID(leadID)
	if err != nil {
		return nil, fmt.Errorf("prospek tidak ditemukan: %w", err)
	}

	return load, nil
}
