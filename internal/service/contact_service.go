package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"errors"

	gormErrors "gorm.io/gorm"
)

type ContactService struct {
	contactRepo *postgres.ContactRepository
}

func NewContactService(contactRepo *postgres.ContactRepository) *ContactService {
	return &ContactService{contactRepo: contactRepo}
}

func (s *ContactService) CreateContact(name string, phone string, email *string, company *string, position *string, source string, assignedTo string) (*entity.Contact, error) {

	existing, err := s.contactRepo.FindByPhone(phone)
	if err != nil && !errors.Is(err, gormErrors.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("nomor telepon sudah terdaftar")
	}

	if email != nil && *email != "" {
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

	return s.contactRepo.CreateContact(contact)
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

func (s *ContactService) GetContactByID(contactID string) (*entity.Contact, error) {
	if contactID == "" {
		return nil, errors.New("ID contact tidak boleh kosong")
	}
	contact, err := s.contactRepo.FindByID(contactID)
	if errors.Is(err, gormErrors.ErrRecordNotFound) {
		return nil, errors.New("contact tidak ditemukan")
	}
	return contact, err
}

func (s *ContactService) UpdateContact(contactID string, updates map[string]interface{}) error {
	if contactID == "" {
		return errors.New("ID contact tidak boleh kosong")
	}

	if phone, ok := updates["phone"].(string); ok && phone != "" {
		existing, err := s.contactRepo.FindByPhone(phone)
		if err != nil && !errors.Is(err, gormErrors.ErrRecordNotFound) {
			return err
		}
		if existing != nil && existing.ID != contactID {
			return errors.New("nomor telepon sudah digunakan contact lain")
		}
	}

	if email, ok := updates["email"].(string); ok && email != "" {
		existingEmail, err := s.contactRepo.FindByEmail(email)
		if err != nil && !errors.Is(err, gormErrors.ErrRecordNotFound) {
			return err
		}
		if existingEmail != nil && existingEmail.ID != contactID {
			return errors.New("email sudah digunakan contact lain")
		}
	}

	return s.contactRepo.UpdateContact(contactID, updates)
}

func (s *ContactService) DeleteContact(contactID string) error {
	if contactID == "" {
		return errors.New("ID contact tidak boleh kosong")
	}

	_, err := s.contactRepo.FindByID(contactID)
	if errors.Is(err, gormErrors.ErrRecordNotFound) {
		return errors.New("contact tidak ditemukan")
	}
	if err != nil {
		return err
	}

	return s.contactRepo.DeleteContact(contactID)
}

func (s *ContactService) FindOrCreateContact(name string, phone string, assignedTo string) (*entity.Contact, error) {
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
