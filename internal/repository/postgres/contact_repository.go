package postgres

import (
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type ContactRepository struct {
	dbgorm *gorm.DB
}

func NewContactRepository(db *gorm.DB) *ContactRepository {
	return &ContactRepository{dbgorm: db}
}

func (r *ContactRepository) CreateContact(contact *entity.Contact) (*entity.Contact, error) {
	err := r.dbgorm.Create(contact).Error
	if err != nil {
		return nil, err
	}
	return contact, nil
}

func (r *ContactRepository) FindByID(contactID string) (*entity.Contact, error) {
	var contact entity.Contact
	err := r.dbgorm.
		Preload("AssignedUser").
		Preload("Leads").
		Preload("Leads.Deals").
		Where("id = ?", contactID).
		First(&contact).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (r *ContactRepository) FindAll(page int, limit int, search string, source string) ([]entity.Contact, int64, error) {
	var contacts []entity.Contact
	var total int64

	baseQuery := r.dbgorm.Model(&entity.Contact{})
	if search != "" {
		baseQuery = baseQuery.Where("to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(phone,'') || ' ' || coalesce(email,'')) @@ plainto_tsquery('simple', ?)", search)
	}
	if source != "" {
		baseQuery = baseQuery.Where("source = ?", source)
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := baseQuery.
		Preload("AssignedUser").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&contacts).Error
	if err != nil {
		return nil, 0, err
	}
	return contacts, total, nil
}

func (r *ContactRepository) SearchContacts(keyword string) ([]entity.Contact, error) {
	var contacts []entity.Contact
	err := r.dbgorm.
		Preload("AssignedUser").
		Where("to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(phone,'') || ' ' || coalesce(email,'')) @@ plainto_tsquery('simple', ?)", keyword).
		Find(&contacts).Error
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

func (r *ContactRepository) UpdateContact(contactID string, updates map[string]interface{}) error {
	return r.dbgorm.Model(&entity.Contact{}).Where("id = ?", contactID).Updates(updates).Error
}

func (r *ContactRepository) DeleteContact(contactID string) error {
	return r.dbgorm.Where("id = ?", contactID).Delete(&entity.Contact{}).Error
}

func (r *ContactRepository) FindByPhone(phone string) (*entity.Contact, error) {
	var contact entity.Contact
	err := r.dbgorm.Where("phone = ?", phone).First(&contact).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (r *ContactRepository) FindByEmail(email string) (*entity.Contact, error) {
	var contact entity.Contact
	err := r.dbgorm.Where("email = ?", email).First(&contact).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (r *ContactRepository) GetTrashedContacts(userID, role string) ([]entity.Contact, error) {
	var contacts []entity.Contact
	query := r.dbgorm.Unscoped().Where("deleted_at IS NOT NULL")
	if role == "sales" {
		query = query.Where("assigned_to = ?", userID)
	}
	err := query.Find(&contacts).Error
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

func (r *ContactRepository) RestoreContact(contactID string) error {
	return r.dbgorm.Unscoped().Where("id = ?", contactID).Update("deleted_at", nil).Error
}
