package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"errors"
	"fmt"
)

type DealService struct {
	deal *postgres.DealRepository
}

func NewDealService(dealRepo *postgres.DealRepository) *DealService {
	return &DealService{
		deal: dealRepo,
	}
}

func (s *DealService) CreateDeal(name string, value float64, leadID string, userID string) (*entity.Deal, error) {
	deal := entity.Deal{
		Name:       name,
		Value:      value,
		Stage:      "PROSPECTING",
		LeadID:     leadID,
		AssignedTo: userID,
	}

	result, err := s.deal.CreateDeal(&deal)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *DealService) GetDeals(userID string, role string) ([]entity.Deal, error) {
	if role == "sales" {
		return s.deal.GetDealByUserId(userID)
	} else {
		return s.deal.GetAllDeals()
	}
}

func (s *DealService) UpdateStage(dealID string, status string, userID string, role string) error {
	if dealID == "" || status == "" || userID == "" || role == "" {
		return errors.New("semua field wajib diisi")
	}

	deal, err := s.deal.GetDealByID(dealID)
	if err != nil {
		return fmt.Errorf("deal tidak ditemukan: %w", err)
	}

	if role == "sales" {
		if deal.AssignedTo != userID {
			return errors.New("unauthorized: ini bukan deal Anda")
		}
	}

	return s.deal.UpdateStage(dealID, status)
}
