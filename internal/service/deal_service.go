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
	"github.com/xuri/excelize/v2"
)

type DealService struct {
	deal        *postgres.DealRepository
	AuditLog    *postgres.AuditRepository
	AsynqClient *asynq.Client
	invoiceRepo *postgres.InvoiceRepository
}

func NewDealService(dealRepo *postgres.DealRepository, auditRepo *postgres.AuditRepository, asyncClient *asynq.Client, invoiceRepo *postgres.InvoiceRepository) *DealService {
	return &DealService{
		deal:        dealRepo,
		AuditLog:    auditRepo,
		AsynqClient: asyncClient,
		invoiceRepo: invoiceRepo,
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
		subTotal := deal.Value
		pajakPpn := subTotal * 0.11
		total := subTotal + pajakPpn
		var InvoiceNo string
		InvoiceNo = "INV"
		wrapTeks := fmt.Sprintf("%s-%s", InvoiceNo, dealID)
		invoice := entity.Invoice{
			InvoiceNo:  wrapTeks,
			DealID:     dealID,
			SubTotal:   subTotal,
			GrandTotal: total,
			Tax:        pajakPpn,
			Status:     "UNPAID",
		}
		_, err := s.invoiceRepo.CreateInvoice(&invoice)
		if err != nil {
			return err
		}
	}
	dealById, err := s.deal.GetDealByID(dealID)
	if err != nil {
		return err
	}

	oldData := dealById.Stage

	errUpdate := s.deal.UpdateStage(dealID, status)
	if errUpdate != nil {
		return errUpdate
	}
	newData := status
	messages := fmt.Sprintf("Merubah status Deal menjadi %s", status)
	logData := entity.AuditLog{
		UserIDAudit: userID,
		Action:      messages,
		TargetID:    dealID,
		OldData:     oldData,
		NewData:     newData,
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

func (s *DealService) ExportDealsToExcel() (*bytes.Buffer, error) {
	data, _, err := s.deal.GetAllDeals(1, 1000, "", "")
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	f.SetCellValue("Sheet1", "A1", "ID Deal")
	f.SetCellValue("Sheet1", "B1", "Nama Deal")
	f.SetCellValue("Sheet1", "C1", "Status Stage")
	f.SetCellValue("Sheet1", "D1", "Nilai (Rp)")
	f.SetCellStyle("Sheet1", "A1", "D1", headerStyle)
	f.SetColWidth("Sheet1", "A", "D", 20)

	for i, deal := range data {
		row := i + 2
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), deal.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), deal.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), deal.Stage)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), deal.Value)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (s *DealService) GenerateInvoicePDF(dealID string) ([]byte, error) {

	invoice, err := s.invoiceRepo.GetInvoiceByDealID(dealID)
	if err != nil {
		return nil, fmt.Errorf("invoice tidak ditemukan: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "INVOICE TAGIHAN", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(60, 10, "Nomor Invoice")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 10, invoice.InvoiceNo)
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(60, 10, "Harga Awal (SubTotal)")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 10, fmt.Sprintf("Rp %.2f", invoice.SubTotal))
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(60, 10, "Pajak PPN (11%)")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 10, fmt.Sprintf("Rp %.2f", invoice.Tax))
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(60, 10, "Grand Total")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 10, fmt.Sprintf("Rp %.2f", invoice.GrandTotal))
	pdf.Ln(20)

	pdf.SetFont("Arial", "I", 10)
	pdf.CellFormat(0, 10, "Harap transfer segera ke Rekening BCA 123456", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("gagal generate PDF: %w", err)
	}

	return buf.Bytes(), nil
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

func (s *DealService) ReorderDeals(dealIDs []string) error {
	return s.deal.UpdateDealPositions(dealIDs)
}
