package service

import (
	"context"
	"crm-project/internal/delivery/websocket"
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"crm-project/internal/worker"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type LeadService struct {
	lead        *postgres.LeadRepository
	user        *postgres.UserRepository
	contact     *postgres.ContactRepository
	AuditLog    *postgres.AuditRepository
	AsynqClient *asynq.Client
	redisClient *redis.Client
}

func NewLeadService(leadRepo *postgres.LeadRepository, userRepo *postgres.UserRepository, asyncClient *asynq.Client, redisClient *redis.Client, AuditLog *postgres.AuditRepository, contactRepo *postgres.ContactRepository) *LeadService {
	return &LeadService{
		lead:        leadRepo,
		user:        userRepo,
		contact:     contactRepo,
		AsynqClient: asyncClient,
		redisClient: redisClient,
		AuditLog:    AuditLog,
	}
}

func (s *LeadService) CreateLead(name string, email string, phone string, userID string, customFields map[string]interface{}) (*entity.Lead, error) {
	var contactID *string
	if phone != "" {
		existing, err := s.contact.FindByPhone(phone)
		if err == nil && existing != nil {
			contactID = &existing.ID
		}
	}
	return s.CreateLeadWithContact(name, email, phone, userID, customFields, contactID)
}

func (s *LeadService) CreateLeadWithContact(name string, email string, phone string, userID string, customFields map[string]interface{}, contactID *string) (*entity.Lead, error) {
	lead := entity.Lead{
		Name:         name,
		Email:        email,
		Phone:        phone,
		Status:       "NEW",
		AssignedTo:   userID,
		ContactID:    contactID,
		CustomFields: customFields,
	}

	newLead, err := s.lead.CreateLead(&lead)
	if err != nil {
		return nil, err
	}

	s.redisClient.Del(context.Background(), "crm_dashboard_stats")

	NewDataJson, err := json.Marshal(newLead)
	if err != nil {
		return nil, err
	}

	logData := entity.AuditLog{
		UserIDAudit: userID,
		Action:      "CREATE-LEAD",
		TargetID:    newLead.ID,
		OldData:     "-",
		NewData:     string(NewDataJson),
	}
	s.AuditLog.CreateAuditLog(&logData)
	return newLead, err
}

func (s *LeadService) CheckStaleLeads() {
	leads, _, err := s.lead.GetAllLeads(1, 1000, "", "")
	if err != nil {
		return
	}
	for _, data := range leads {
		if data.Status == "NEW" && time.Since(data.CreatedAt).Hours() > 72 {
			findUser, err := s.user.FindByID(data.AssignedTo)
			if err != nil {
				continue
			}
			messageToLead := fmt.Sprintf("Halo bos %s, Lead bernama %s belum...", findUser.Name, data.Name)
			pesan := fmt.Sprintf("Peringatan! Lead prospek bernama %s sudah lebih dari 3 hari belum Anda follow-up!", data.Name)
			taskEmail, errTask := worker.NewEmailDeliveryTask(findUser.Email, "Peringatan Follow-up", messageToLead)
			if errTask == nil {
				s.AsynqClient.Enqueue(taskEmail)
			}
			admins, err := s.user.GetAdmins()
			for _, admin := range admins {
				message := fmt.Sprintf("Peringatan Klien  Bernama %s telah ditelantarkan oleh Sale %s selama 3 hari!", admin.Name, findUser.Name)
				websocket.SendMessageToUser(admin.Email, message)
			}
			websocket.SendMessageToUser(data.AssignedTo, pesan)

			err = s.lead.UpdateStatus(data.ID, "ESCALATED")
			if err != nil {
				continue
			}
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

func (s *LeadService) GetLeadByID(leadID string) (*entity.Lead, error) {
	return s.lead.GetLeadByID(leadID)
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

	oldDataJSON, _ := json.Marshal(load)
	err = s.lead.UpdateStatus(leadID, status)
	if err != nil {
		return err
	}

	newLead, err := s.lead.GetLeadByID(leadID)
	if err != nil {
		return nil
	}
	newDataJSON, _ := json.Marshal(newLead)
	logData := entity.AuditLog{
		UserIDAudit: userID,
		Action:      "UPDATE_LEAD_STATUS",
		TargetID:    leadID,
		OldData:     string(oldDataJSON),
		NewData:     string(newDataJSON),
	}
	s.AuditLog.CreateAuditLog(&logData)
	return nil
}

func (s *LeadService) DeleteLead(leadID string, userID string, role string) error {
	if leadID == "" {
		return errors.New("ID lead wajib diisi")
	}

	oldLead, err := s.lead.GetLeadByID(leadID)
	if err != nil {
		return fmt.Errorf("lead tidak ditemukan: %w", err)
	}

	if role == "sales" {
		if oldLead.AssignedTo != userID {
			return errors.New("unauthorized: ini bukan lead Anda")
		}
	}

	err = s.lead.SoftDeleteLead(leadID)
	if err == nil {
		s.redisClient.Del(context.Background(), "crm_dashboard_stats")
	}
	oldDataJSON, err := json.Marshal(oldLead)
	if err != nil {
		return nil
	}

	logData := entity.AuditLog{
		UserIDAudit: userID,
		Action:      "DELETE-LEAD",
		TargetID:    leadID,
		OldData:     string(oldDataJSON),
		NewData:     "-",
	}
	s.AuditLog.CreateAuditLog(&logData)
	return err
}

func (s *LeadService) UpdateLead(leadID string, name string, email string, phone string, userID string, role string) error {
	if leadID == "" {
		return errors.New("ID lead wajib diisi")
	}

	oldLead, err := s.lead.GetLeadByID(leadID)
	if err != nil {
		return fmt.Errorf("lead tidak ditemukan: %w", err)
	}

	if role == "sales" {
		if oldLead.AssignedTo != userID {
			return errors.New("unauthorized: ini bukan lead Anda")
		}
	}

	err = s.lead.UpdateLead(leadID, name, email, phone)
	if err != nil {
		return err
	}

	s.redisClient.Del(context.Background(), "crm_dashboard_stats")
	newLead, err := s.lead.GetLeadByID(leadID)
	if err != nil {
		return nil
	}
	oldDataJSON, err := json.Marshal(oldLead)
	if err != nil {
		return nil
	}
	newDataJSON, err := json.Marshal(newLead)
	if err != nil {
		return nil
	}
	logData := entity.AuditLog{
		UserIDAudit: userID,
		Action:      "UPDATE_LEAD",
		TargetID:    leadID,
		OldData:     string(oldDataJSON),
		NewData:     string(newDataJSON),
	}

	s.AuditLog.CreateAuditLog(&logData)

	return nil
}

func (s *LeadService) ImportBulkLeads(leads []entity.Lead) error {
	_, err := s.lead.CreateBulkLeads(&leads)
	if err == nil {
		s.redisClient.Del(context.Background(), "crm_dashboard_stats")
	}
	return err
}

func (s *LeadService) GetTrashedLeads(userID string, role string) (*[]entity.Lead, error) {
	leads, err := s.lead.GetTrashedLeads(userID, role)
	if err != nil {
		return nil, err
	}
	return leads, nil
}

func (s *LeadService) RestoreLead(leadID string, userID string) error {
	if leadID == "" {
		return errors.New("ID lead wajib diisi")
	}

	err := s.lead.RestoreLead(leadID)
	if err != nil {
		return err
	}
	s.redisClient.Del(context.Background(), "crm_dashboard_stats")
	restoredLead, err := s.lead.GetLeadByID(leadID)
	if err != nil {
		return nil
	}
	newDataJSON, err := json.Marshal(restoredLead)
	if err != nil {
		return nil
	}
	logData := entity.AuditLog{
		UserIDAudit: userID,
		Action:      "RESTORE_LEAD",
		TargetID:    leadID,
		OldData:     "-",
		NewData:     string(newDataJSON),
	}

	s.AuditLog.CreateAuditLog(&logData)
	return nil
}
