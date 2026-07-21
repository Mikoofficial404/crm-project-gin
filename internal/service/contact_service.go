package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"crm-project/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	gormErrors "gorm.io/gorm"
)

type ContactService struct {
	contactRepo *postgres.ContactRepository
	AuditLog    *postgres.AuditRepository
}

func NewContactService(contactRepo *postgres.ContactRepository, auditLog *postgres.AuditRepository) *ContactService {
	return &ContactService{contactRepo: contactRepo, AuditLog: auditLog}
}

func (s *ContactService) CreateContact(name string, phone string, email *string, company *string, position *string, source string, assignedTo string) (*entity.Contact, error) {

	phone = utils.NormalizePhone(phone)

	existing, err := s.contactRepo.FindByPhone(phone)
	if err != nil && !errors.Is(err, gormErrors.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("nomor telepon sudah terdaftar")
	}

	if email != nil && *email != "" {
		if !utils.IsValidEmail(*email) {
			return nil, errors.New("format email tidak valid")
		}
		existingEmail, err := s.contactRepo.FindByEmail(*email)
		if err != nil && !errors.Is(err, gormErrors.ErrRecordNotFound) {
			return nil, err
		}
		if existingEmail != nil {
			return nil, errors.New("email sudah terdaftar")
		}
	}

	contact := &entity.Contact{
		Name:       name,
		Phone:      phone,
		Email:      email,
		Company:    company,
		Position:   position,
		Source:     source,
		AssignedTo: assignedTo,
	}

	newContact, err := s.contactRepo.CreateContact(contact)
	if err != nil {
		return nil, err
	}
	newDatajson, err := json.Marshal(newContact)
	if err != nil {
		return nil, err
	}

	logData := entity.AuditLog{
		UserIDAudit: newContact.AssignedTo,
		Action:      "create_contact",
		TargetID:    newContact.ID,
		OldData:     "",
		NewData:     string(newDatajson),
	}

	s.AuditLog.CreateAuditLog(&logData)
	return newContact, nil
}

func (s *ContactService) GetAllContacts(page int, limit int, search string, source string) ([]entity.Contact, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	return s.contactRepo.FindAll(page, limit, search, source)
}

func (s *ContactService) GetContactByID(contactID string, userID string, role string) (*entity.Contact, error) {
	if contactID == "" {
		return nil, errors.New("ID contact tidak boleh kosong")
	}
	contact, err := s.contactRepo.FindByID(contactID)
	if errors.Is(err, gormErrors.ErrRecordNotFound) {
		return nil, errors.New("contact tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}

	if role == "sales" && contact.AssignedTo != userID {
		return nil, errors.New("contact tidak ditemukan")
	}

	return contact, nil
}

func (s *ContactService) UpdateContact(contactID string, updates map[string]interface{}) error {
	if contactID == "" {
		return errors.New("ID contact tidak boleh kosong")
	}

	if phone, ok := updates["phone"].(string); ok && phone != "" {
		normalized := utils.NormalizePhone(phone)
		updates["phone"] = normalized

		existing, err := s.contactRepo.FindByPhone(normalized)
		if err != nil && !errors.Is(err, gormErrors.ErrRecordNotFound) {
			return err
		}
		if existing != nil && existing.ID != contactID {
			return errors.New("nomor telepon sudah digunakan contact lain")
		}
	}

	if email, ok := updates["email"].(string); ok && email != "" {
		if !utils.IsValidEmail(email) {
			return errors.New("format email tidak valid")
		}
		existingEmail, err := s.contactRepo.FindByEmail(email)
		if err != nil && !errors.Is(err, gormErrors.ErrRecordNotFound) {
			return err
		}
		if existingEmail != nil && existingEmail.ID != contactID {
			return errors.New("email sudah digunakan contact lain")
		}
	}

	err := s.contactRepo.UpdateContact(contactID, updates)
	if err != nil {
		return err
	}

	updatedContact, err := s.contactRepo.FindByID(contactID)
	if err != nil {
		return err
	}

	newDatajson, err := json.Marshal(updatedContact)
	if err != nil {
		return err
	}

	logData := entity.AuditLog{
		UserIDAudit: updatedContact.AssignedTo,
		Action:      "update_contact",
		TargetID:    updatedContact.ID,
		OldData:     "",
		NewData:     string(newDatajson),
	}
	s.AuditLog.CreateAuditLog(&logData)

	return nil
}

func (s *ContactService) ExportContacts(userID, role, source string) ([]entity.Contact, error) {
	if role == "sales" {
		return s.contactRepo.GetAllContactsForExport(userID, source)
	}
	return s.contactRepo.GetAllContactsForExport("", source)
}

func (s *ContactService) DeleteContact(contactID string) error {
	if contactID == "" {
		return errors.New("ID contact tidak boleh kosong")
	}

	oldContact, err := s.contactRepo.FindByID(contactID)
	if errors.Is(err, gormErrors.ErrRecordNotFound) {
		return errors.New("contact tidak ditemukan")
	}
	if err != nil {
		return err
	}

	oldDataJson, _ := json.Marshal(oldContact)

	logData := entity.AuditLog{
		UserIDAudit: oldContact.AssignedTo,
		Action:      "delete_contact",
		TargetID:    contactID,
		OldData:     string(oldDataJson),
		NewData:     "",
	}
	s.AuditLog.CreateAuditLog(&logData)

	return s.contactRepo.DeleteContact(contactID)
}

func (s *ContactService) GetTrashedContacts(userID, role string) ([]entity.Contact, error) {
	return s.contactRepo.GetTrashedContacts(userID, role)
}

func (s *ContactService) RestoreContact(contactID string) error {
	return s.contactRepo.RestoreContact(contactID)
}

func (s *ContactService) ImportContactsFromCSV(records [][]string, userID string) (imported int, skipped int, errMessages []string) {
	for i, record := range records {
		if len(record) < 2 {
			skipped++
			errMessages = append(errMessages, fmt.Sprintf("baris %d: kolom tidak lengkap", i+1))
			continue
		}

		name := strings.TrimSpace(record[0])
		phone := strings.TrimSpace(record[1])
		email := ""
		if len(record) >= 3 {
			email = strings.TrimSpace(record[2])
		}

		phone = strings.ReplaceAll(phone, " ", "")
		if strings.HasPrefix(phone, "08") {
			phone = "+628" + phone[2:]
		}

		if email != "" && !strings.Contains(email, "@") {
			skipped++
			errMessages = append(errMessages, fmt.Sprintf("baris %d: email tidak valid", i+1))
			continue
		}

		existing, _ := s.contactRepo.FindByPhone(phone)
		if existing != nil {
			skipped++
			errMessages = append(errMessages, fmt.Sprintf("baris %d: nomor %s sudah terdaftar", i+1, phone))
			continue
		}

		var emailPtr *string
		if email != "" {
			emailPtr = &email
		}

		_, err := s.contactRepo.CreateContact(&entity.Contact{
			Name:       name,
			Phone:      phone,
			Email:      emailPtr,
			AssignedTo: userID,
		})
		if err != nil {
			skipped++
			errMessages = append(errMessages, fmt.Sprintf("baris %d: %s", i+1, err.Error()))
			continue
		}

		imported++
	}
	return imported, skipped, errMessages
}

func (s *ContactService) FindOrCreateContact(name string, phone string, assignedTo string) (*entity.Contact, error) {

	phone = utils.NormalizePhone(phone)

	existing, err := s.contactRepo.FindByPhone(phone)
	if err != nil && !errors.Is(err, gormErrors.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	contact := &entity.Contact{
		Name:       name,
		Phone:      phone,
		Source:     "whatsapp",
		AssignedTo: assignedTo,
	}
	return s.contactRepo.CreateContact(contact)
}
