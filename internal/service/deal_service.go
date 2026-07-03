package service

import (
	"bytes"
	"context"
	"crm-project/internal/delivery/websocket"
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/jung-kurt/gofpdf"
	"github.com/redis/go-redis/v9"
	"github.com/xuri/excelize/v2"
)

type DealService struct {
	deal            *postgres.DealRepository
	AuditLog        *postgres.AuditRepository
	AsynqClient     *asynq.Client
	invoiceRepo     *postgres.InvoiceRepository
	redisClient     *redis.Client
	stageRepo       *postgres.PipelineStageRepository
	dealProductRepo *postgres.DealProductRepository
}

func NewDealService(dealRepo *postgres.DealRepository, auditRepo *postgres.AuditRepository, asyncClient *asynq.Client, invoiceRepo *postgres.InvoiceRepository, redisClient *redis.Client, stageRepo *postgres.PipelineStageRepository, dealProductRepo *postgres.DealProductRepository) *DealService {
	return &DealService{
		deal:            dealRepo,
		AuditLog:        auditRepo,
		AsynqClient:     asyncClient,
		invoiceRepo:     invoiceRepo,
		redisClient:     redisClient,
		stageRepo:       stageRepo,
		dealProductRepo: dealProductRepo,
	}
}

func (s *DealService) CreateDeal(name string, value float64, leadID string, userID string, pipelineID string, stageID string) (*entity.Deal, error) {
	if name == "" || leadID == "" || pipelineID == "" || stageID == "" {
		return nil, errors.New("name, lead_id, pipeline_id, dan stage_id wajib diisi")
	}

	deal := entity.Deal{
		Name:       name,
		Value:      value,
		PipelineID: pipelineID,
		StageID:    stageID,
		LeadID:     leadID,
		AssignedTo: userID,
	}

	NewDeal, err := s.deal.CreateDeal(&deal)
	if err != nil {
		return nil, err
	}

	s.redisClient.Del(context.Background(), "crm_dashboard_stats")

	NewDataJson, err := json.Marshal(NewDeal)
	if err != nil {
		return nil, err
	}
	logData := entity.AuditLog{
		UserIDAudit: userID,
		Action:      "CREATE-DEAL",
		TargetID:    NewDeal.ID,
		OldData:     "-",
		NewData:     string(NewDataJson),
	}
	s.AuditLog.CreateAuditLog(&logData)
	return NewDeal, nil
}

func (s *DealService) GetDeals(userID string, role string, page int, limit int, search string, stageID string, pipelineID string) ([]entity.Deal, int64, error) {
	if role == "sales" {
		return s.deal.GetDealByUserId(userID, page, limit, search, stageID, pipelineID)
	} else {
		return s.deal.GetAllDeals(page, limit, search, stageID, pipelineID)
	}
}

func (s *DealService) UpdateInvoiceStatus(invoideID string, status string) error {
	if invoideID == "" {
		return errors.New("ID invoice wajib diisi")
	}
	if status != "PAID" && status != "OVERDUE" {
		return errors.New("status invoice tidak valid, hanya boleh PAID atau OVERDUE")
	}
	_, err := s.invoiceRepo.UpdateStatus(invoideID, status)
	if err != nil {
		return err
	}
	return err
}

func (s *DealService) UpdateStage(dealID string, stageID string, userID string, role string) error {

	if dealID == "" || stageID == "" || userID == "" || role == "" {
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
	if stageID == deal.StageID {
		return errors.New("stage sudah sama, tidak ada perubahan")
	}
	dealById, err := s.deal.GetDealByID(dealID)
	if err != nil {
		return err
	}

	oldData := dealById.StageID

	errUpdate := s.deal.UpdateStage(dealID, stageID)
	if errUpdate != nil {
		return errUpdate
	}

	stage, err := s.stageRepo.GetStageByID(stageID)
	if err == nil && stage.IsClosedWon {

		websocket.SendMessageToUser(deal.AssignedTo, "SELAMAT! Anda baru saja memenangkan Deal!!")

		existingInvoice, _ := s.invoiceRepo.GetInvoiceByDealID(dealID)
		if existingInvoice == nil {
			wrapTeks := fmt.Sprintf("INV-%s", dealID)

			dealProducts, _ := s.dealProductRepo.GetProductsByDealID(dealID)

			var subTotal float64
			var invoiceItems []entity.InvoiceItem

			if len(dealProducts) > 0 {
				for _, dp := range dealProducts {
					subTotal += dp.SubTotal
					invoiceItems = append(invoiceItems, entity.InvoiceItem{
						ProductID: dp.ProductID,
						Quantity:  dp.Quantity,
						UnitPrice: dp.UnitPrice,
						SubTotal:  dp.SubTotal,
					})
				}
			} else {

				subTotal = deal.Value
			}

			pajakPpn := subTotal * 0.11
			total := subTotal + pajakPpn

			invoice := entity.Invoice{
				InvoiceNo:  wrapTeks,
				DealID:     dealID,
				SubTotal:   subTotal,
				GrandTotal: total,
				Tax:        pajakPpn,
				Status:     "UNPAID",
				Items:      invoiceItems,
			}
			_, err := s.invoiceRepo.CreateInvoice(&invoice)
			if err != nil {
				return err
			}
		}
	}

	messages := fmt.Sprintf("Merubah stage Deal menjadi %s", stageID)
	logData := entity.AuditLog{
		UserIDAudit: userID,
		Action:      messages,
		TargetID:    dealID,
		OldData:     oldData,
		NewData:     stageID,
	}

	s.AuditLog.CreateAuditLog(&logData)

	s.redisClient.Del(context.Background(), "crm_dashboard_stats")

	return nil
}

func (s *DealService) DeleteDeal(dealID string, userID string, role string) error {
	if dealID == "" {
		return errors.New("ID deal wajib diisi")
	}

	oldDeal, err := s.deal.GetDealByID(dealID)
	if err != nil {
		return fmt.Errorf("deal tidak ditemukan: %w", err)
	}

	if role == "sales" {
		if oldDeal.AssignedTo != userID {
			return errors.New("unauthorized: ini bukan deal Anda")
		}
	}

	err = s.deal.SoftDeleteDeal(dealID)
	if err == nil {
		s.redisClient.Del(context.Background(), "crm_dashboard_stats")
	}
	oldDataJSON, err := json.Marshal(oldDeal)
	if err != nil {
		return nil
	}
	logData := entity.AuditLog{
		UserIDAudit: userID,
		Action:      "DELETE-DEAL",
		TargetID:    dealID,
		OldData:     string(oldDataJSON),
		NewData:     "-",
	}
	s.AuditLog.CreateAuditLog(&logData)
	return err
}

func (s *DealService) ExportDealsToExcel() (*bytes.Buffer, error) {
	data, _, err := s.deal.GetAllDeals(1, 1000, "", "", "")
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
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), deal.Stage.Name)
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
	data, _, err := s.deal.GetAllDeals(1, 10, "", "", "")
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
		pdf.Cell(60, 10, value.Stage.Name)
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

func (s *DealService) UpdateDeal(dealID string, name string, value float64, userID string, role string) error {
	if dealID == "" || name == "" {
		return errors.New("id dan nama deal wajib diisi")
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

	errUpdate := s.deal.UpdateDeal(dealID, name, value)
	if errUpdate != nil {
		return errUpdate
	}

	messages := fmt.Sprintf("Mengubah data Deal (Nama: %s, Value: %.2f)", name, value)
	logData := entity.AuditLog{
		UserIDAudit: userID,
		Action:      messages,
		TargetID:    dealID,
		OldData:     deal.Name,
		NewData:     name,
	}
	s.AuditLog.CreateAuditLog(&logData)
	s.redisClient.Del(context.Background(), "crm_dashboard_stats")

	return nil
}

func (s *DealService) AssignProductToDeal(dealID, productID string, quantity int, unitPrice float64, userID, role string) (*entity.DealProduct, error) {
	if dealID == "" || productID == "" {
		return nil, errors.New("deal ID dan product ID wajib diisi")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity harus lebih dari 0")
	}
	if unitPrice < 0 {
		return nil, errors.New("unit price tidak boleh negatif")
	}

	deal, err := s.deal.GetDealByID(dealID)
	if err != nil {
		return nil, fmt.Errorf("deal tidak ditemukan: %w", err)
	}
	if role == "sales" && deal.AssignedTo != userID {
		return nil, errors.New("unauthorized: ini bukan deal Anda")
	}

	subTotal := float64(quantity) * unitPrice
	dealProduct := &entity.DealProduct{
		DealID:    dealID,
		ProductID: productID,
		Quantity:  quantity,
		UnitPrice: unitPrice,
		SubTotal:  subTotal,
	}

	return s.dealProductRepo.AssignProduct(dealProduct)
}

func (s *DealService) RemoveProductFromDeal(dealID, productID, userID, role string) error {
	if dealID == "" || productID == "" {
		return errors.New("deal ID dan product ID wajib diisi")
	}

	deal, err := s.deal.GetDealByID(dealID)
	if err != nil {
		return fmt.Errorf("deal tidak ditemukan: %w", err)
	}
	if role == "sales" && deal.AssignedTo != userID {
		return errors.New("unauthorized: ini bukan deal Anda")
	}

	return s.dealProductRepo.RemoveProduct(dealID, productID)
}

func (s *DealService) GetDealProducts(dealID string) ([]entity.DealProduct, error) {
	if dealID == "" {
		return nil, errors.New("deal ID wajib diisi")
	}
	return s.dealProductRepo.GetProductsByDealID(dealID)
}
