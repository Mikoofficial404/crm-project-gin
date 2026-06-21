package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"errors"
	"fmt"
)

type LeadService struct {
	lead *postgres.LeadRepository
}

func NewLeadService(leadRepo *postgres.LeadRepository) *LeadService {
	return &LeadService{
		lead: leadRepo,
	}
}

func (s *LeadService) CreateLead(name string, email string, phone string, userID string) (*entity.Lead, error) {
	lead := entity.Lead{
		Name:       name,
		Email:      email,
		Phone:      phone,
		Status:     "NEW",
		AssignedTo: userID,
	}

	isResult, err := s.lead.CreateLead(&lead)
	if err != nil {
		return nil, err
	}
	return isResult, err
}

func (s *LeadService) GetLeads(userID string, role string) ([]entity.Lead, error) {
	if role == "admin" {
		return s.lead.GetAllLeads()
	} else {
		return s.lead.GetLeadsByUserId(userID)
	}
}

func (s *LeadService) UpdateLeadStatus(leadID string, status string, userID string, role string) error {
	if leadID == "" || status == "" || userID == "" || role == "" {
		return errors.New("semua field wajib diisi")
	}
	load, err := s.lead.GetLeadByID(leadID)
	if err != nil {
		return fmt.Errorf("prospek tidak ditemukan: %w", err)
	}
	if role == "sales" {
		if load.AssignedTo != userID {
			return errors.New("unauthorized: ini bukan prospek Anda")
		}
	}
	return s.lead.UpdateStatus(leadID, status)
}
