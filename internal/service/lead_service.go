package service

import (
	"crm-project/internal/delivery/websocket"
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"errors"
	"fmt"
	"time"
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

func (s *LeadService) CheckStaleLeads() {
	leads, _, err := s.lead.GetAllLeads(1, 1000, "", "")
	if err != nil {
		return
	}
	for _, data := range leads {
		if data.Status == "NEW" && time.Since(data.CreatedAt).Hours() > 72 {
			pesan := fmt.Sprintf("Peringatan! Lead prospek bernama %s sudah lebih dari 3 hari belum Anda follow-up!", data.Name)
			websocket.SendMessageToUser(data.AssignedTo, pesan)
		}
	}
}

func (s *LeadService) GetLeads(userID string, role string, page int, limit int, search string, status string) ([]entity.Lead, int64, error) {
	if role == "admin" {
		return s.lead.GetAllLeads(page, limit, search, status)
	} else {
		return s.lead.GetLeadsByUserId(userID, page, limit, search, status)
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

func (s *LeadService) DeleteLead(leadID string, userID string, role string) error {
	if leadID == "" {
		return errors.New("ID lead wajib diisi")
	}

	lead, err := s.lead.GetLeadByID(leadID)
	if err != nil {
		return fmt.Errorf("lead tidak ditemukan: %w", err)
	}

	if role == "sales" {
		if lead.AssignedTo != userID {
			return errors.New("unauthorized: ini bukan lead Anda")
		}
	}

	return s.lead.SoftDeleteLead(leadID)
}

func (s *LeadService) ImportBulkLeads(leads []entity.Lead) error {
	_, err := s.lead.CreateBulkLeads(&leads)
	return err
}

func (s *LeadService) GetTrashedLeads(userID string, role string) (*[]entity.Lead, error) {
	leads, err := s.lead.GetTrashedLeads(userID, role)
	if err != nil {
		return nil, err
	}
	return leads, nil
}

func (s *LeadService) RestoreLead(leadID string) error {
	err := s.lead.RestoreLead(leadID)
	if err != nil {
		return err
	}
	return nil
}
