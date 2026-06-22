package service

import (
	"crm-project/internal/delivery/websocket"
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

func (s *DealService) GetDeals(userID string, role string, page int, limit int, search string, stage string) ([]entity.Deal, int64, error) {
	if role == "sales" {
		return s.deal.GetDealByUserId(userID, page, limit, search, stage)
	} else {
		return s.deal.GetAllDeals(page, limit, search, stage)
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
	if status == "WON" {
		websocket.SendMessageToUser(userID, "SELAMAT! Anda baru saja memenangkan Deal!!")
	}
	return s.deal.UpdateStage(dealID, status)
}

func (s *DealService) DeleteDeal(dealID string, userID string, role string) error {
	if dealID == "" {
		return errors.New("ID deal wajib diisi")
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

	return s.deal.SoftDeleteDeal(dealID)
}
