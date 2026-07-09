package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"crm-project/internal/worker"
	"errors"
	"time"

	"github.com/hibiken/asynq"
)

type CampaignService struct {
	campaignRepo  *postgres.CampaignRepository
	recipientRepo *postgres.CampaignRecipientRepository
	leadRepo      *postgres.LeadRepository
	contactRepo   *postgres.ContactRepository
	asynqClient   *asynq.Client
}

func NewCampaignService(
	campaignRepo *postgres.CampaignRepository,
	recipientRepo *postgres.CampaignRecipientRepository,
	leadRepo *postgres.LeadRepository,
	contactRepo *postgres.ContactRepository,
	asynqClient *asynq.Client,
) *CampaignService {
	return &CampaignService{
		campaignRepo:  campaignRepo,
		recipientRepo: recipientRepo,
		leadRepo:      leadRepo,
		contactRepo:   contactRepo,
		asynqClient:   asynqClient,
	}
}

func (s *CampaignService) CreateCampaign(name, subject, body, channel, createdBy string) (*entity.Campaign, error) {
	if name == "" || subject == "" || body == "" {
		return nil, errors.New("nama, subjek, dan isi pesan wajib diisi")
	}
	if channel != "email" && channel != "whatsapp" {
		channel = "email"
	}
	campaign := &entity.Campaign{
		Name:      name,
		Subject:   subject,
		Body:      body,
		Status:    "DRAFT",
		Channel:   channel,
		CreatedBy: createdBy,
	}
	return s.campaignRepo.CreateCampaign(campaign)
}

func (s *CampaignService) GetAllCampaigns(status string) ([]entity.Campaign, error) {
	return s.campaignRepo.GetAllCampaigns(status)
}

func (s *CampaignService) GetCampaignByID(campaignID string) (*entity.Campaign, error) {
	return s.campaignRepo.GetCampaignByID(campaignID)
}

func (s *CampaignService) UpdateCampaign(campaignID, userID string, name, subject, body *string) error {
	campaign, err := s.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return err
	}
	if campaign.Status != "DRAFT" {
		return errors.New("hanya kampanye DRAFT yang bisa diedit")
	}

	updates := map[string]interface{}{}
	if name != nil {
		updates["name"] = *name
	}
	if subject != nil {
		updates["subject"] = *subject
	}
	if body != nil {
		updates["body"] = *body
	}
	if len(updates) == 0 {
		return nil
	}
	return s.campaignRepo.UpdateCampaign(campaignID, updates)
}

func (s *CampaignService) DeleteCampaign(campaignID string) error {
	campaign, err := s.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return err
	}
	if campaign.Status != "DRAFT" {
		return errors.New("hanya kampanye DRAFT yang bisa dihapus")
	}
	return s.campaignRepo.DeleteCampaign(campaignID)
}

func (s *CampaignService) AddRecipientsByLeadIDs(campaignID string, leadIDs []string) error {
	campaign, err := s.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return err
	}

	if campaign.Status != "DRAFT" {
		return errors.New("tidak bisa menambahkan penerima ke kampanye yang bukan DRAFT")
	}

	var recipients []entity.CampaignRecipient
	for _, leadID := range leadIDs {
		lead, err := s.leadRepo.GetLeadByID(leadID)
		if err != nil {
			continue
		}
		lid := leadID
		if campaign.Channel == "whatsapp" {
			if lead.Phone == "" {
				continue
			}
			recipients = append(recipients, entity.CampaignRecipient{
				CampaignID: campaignID,
				LeadID:     &lid,
				Phone:      lead.Phone,
				Status:     "PENDING",
			})
		} else {
			if lead.Email == "" {
				continue
			}
			recipients = append(recipients, entity.CampaignRecipient{
				CampaignID: campaignID,
				LeadID:     &lid,
				Email:      lead.Email,
				Status:     "PENDING",
			})
		}
	}

	if len(recipients) == 0 {
		return errors.New("tidak ada penerima yang valid ditemukan")
	}
	return s.recipientRepo.BulkInsertRecipients(recipients)
}

func (s *CampaignService) AddRecipientsByContactIDs(campaignID string, contactIDs []string) error {
	campaign, err := s.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return err
	}
	if campaign.Status != "DRAFT" {
		return errors.New("tidak bisa menambahkan penerima ke kampanye yang bukan DRAFT")
	}

	var recipients []entity.CampaignRecipient
	for _, contactID := range contactIDs {
		contact, err := s.contactRepo.FindByID(contactID)
		if err != nil {
			continue
		}
		cid := contactID
		if campaign.Channel == "whatsapp" {
			if contact.Phone == "" {
				continue
			}
			recipients = append(recipients, entity.CampaignRecipient{
				CampaignID: campaignID,
				ContactID:  &cid,
				Phone:      contact.Phone,
				Status:     "PENDING",
			})
		} else {
			if contact.Email == nil || *contact.Email == "" {
				continue
			}
			recipients = append(recipients, entity.CampaignRecipient{
				CampaignID: campaignID,
				ContactID:  &cid,
				Email:      *contact.Email,
				Status:     "PENDING",
			})
		}
	}

	if len(recipients) == 0 {
		return errors.New("tidak ada penerima yang valid ditemukan")
	}
	return s.recipientRepo.BulkInsertRecipients(recipients)
}

func (s *CampaignService) SendCampaign(campaignID string) error {
	campaign, err := s.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return err
	}
	if campaign.Status != "DRAFT" && campaign.Status != "SCHEDULED" {
		return errors.New("kampanye harus berstatus DRAFT atau SCHEDULED untuk dikirim")
	}

	pending, err := s.recipientRepo.GetPendingRecipients(campaignID)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return errors.New("tidak ada penerima tertunda yang ditemukan")
	}

	if err := s.campaignRepo.UpdateCampaign(campaignID, map[string]interface{}{
		"status": "SENDING",
	}); err != nil {
		return err
	}

	for _, r := range pending {
		var task *asynq.Task
		var err error
		if campaign.Channel == "whatsapp" {
			task, err = worker.NewCampaignWhatsAppTask(campaignID, r.ID, r.Phone, campaign.Body)
		} else {
			task, err = worker.NewCampaignEmailTask(campaignID, r.ID, r.Email, campaign.Subject, campaign.Body)
		}
		if err != nil {
			continue
		}
		s.asynqClient.Enqueue(task)
	}

	return nil
}

func (s *CampaignService) ScheduleCampaign(campaignID string, scheduledAt time.Time) error {
	campaign, err := s.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return err
	}
	if campaign.Status != "DRAFT" {
		return errors.New("hanya kampanye DRAFT yang bisa dijadwalkan")
	}
	// Berikan toleransi waktu mundur 2 menit untuk proses klik dan loading
	if scheduledAt.Before(time.Now().Add(-2 * time.Minute)) {
		return errors.New("waktu jadwal harus di masa depan")
	}

	pending, err := s.recipientRepo.GetPendingRecipients(campaignID)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return errors.New("tambahkan penerima terlebih dahulu sebelum menjadwalkan")
	}

	return s.campaignRepo.UpdateCampaign(campaignID, map[string]interface{}{
		"status":       "SCHEDULED",
		"scheduled_at": scheduledAt,
	})
}

func (s *CampaignService) GetCampaignStats(campaignID string) (*postgres.CampaignStats, error) {
	_, err := s.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return nil, err
	}
	return s.recipientRepo.GetCampaignStats(campaignID)
}

func (s *CampaignService) GetRecipients(campaignID string, page, limit int) ([]entity.CampaignRecipient, int64, error) {
	_, err := s.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return nil, 0, err
	}
	return s.recipientRepo.GetRecipientsByCampaign(campaignID, page, limit)
}

func (s *CampaignService) ProcessScheduledCampaigns() {
	campaigns, err := s.campaignRepo.GetScheduledCampaigns()
	if err != nil {
		return
	}
	for _, c := range campaigns {
		s.SendCampaign(c.ID)
	}
}

func (s *CampaignService) MarkCampaignSentIfDone(campaignID string) error {
	pending, err := s.recipientRepo.CountPendingRecipients(campaignID)
	if err != nil {
		return err
	}
	if pending == 0 {
		now := time.Now()
		return s.campaignRepo.UpdateCampaign(campaignID, map[string]interface{}{
			"status":  "SENT",
			"sent_at": now,
		})
	}
	return nil
}
