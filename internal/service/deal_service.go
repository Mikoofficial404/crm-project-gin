package service

import (
	"bytes"
	"crm-project/internal/delivery/websocket"
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"crm-project/internal/worker"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/jung-kurt/gofpdf"
)

type DealService struct {
	deal        *postgres.DealRepository
	AuditLog    *postgres.AuditRepository
	AsynqClient *asynq.Client
}

func NewDealService(dealRepo *postgres.DealRepository, auditRepo *postgres.AuditRepository, asyncClient *asynq.Client) *DealService {
	return &DealService{
		deal:        dealRepo,
		AuditLog:    auditRepo,
		AsynqClient: asyncClient,
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
		task, errTask := worker.NewEmailDeliveryTask(
			"klien_anda@gmail.com",
			"SELAMAT! Deal Anda Berhasil!",
			"<h1>Terima Kasih!</h1><p>Kami sangat senang bekerja sama dengan Anda.</p>",
		)
		if errTask == nil {
			s.AsynqClient.Enqueue(task)
		}
		websocket.SendMessageToUser(userID, "SELAMAT! Anda baru saja memenangkan Deal!!")
	}

	errUpdate := s.deal.UpdateStage(dealID, status)
	if errUpdate != nil {
		return errUpdate
	}
	messages := fmt.Sprintf("Merubah status Deal menjadi %s", status)
	logData := entity.AuditLog{
		UserIDAudit: userID,
		Action:      messages,
		TargetID:    dealID,
	}

	s.AuditLog.CreateAuditLog(&logData)
	return nil
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

func (s *DealService) ExportDealsToPDF() ([]byte, error) {
	data, _, err := s.deal.GetAllDeals(1, 10, "", "")
	if err != nil {
		return nil, err
	}
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(60, 10, "Nama Deal")
	pdf.Cell(60, 10, "Stage")
	pdf.Cell(50, 10, "Nilai Rp")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	for _, value := range data {
		pdf.Cell(60, 10, value.Name)
		pdf.Cell(60, 10, value.Stage)
		pdf.Cell(50, 10, formatRupiah(value.Value))
		pdf.Ln(8)
	}
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func formatRupiah(amount float64) string {
	return fmt.Sprintf("Rp %.2f", amount)
}
