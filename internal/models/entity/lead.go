package entity

import (
	"time"

	"gorm.io/gorm"
)

type Lead struct {
	ID           string                 `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name         string                 `json:"name" gorm:"column:name"`
	Email        string                 `json:"email" gorm:"column:email"`
	Phone        string                 `json:"phone" gorm:"column:phone"`
	Status       string                 `json:"status" gorm:"size:20;not null;default:NEW"`
	AssignedTo   string                 `json:"assignedTo" gorm:"type:uuid;not null"`
	ContactID    *string                `json:"contactId,omitempty" gorm:"type:uuid"`
	Deals        []Deal                 `gorm:"foreignKey:LeadID"`
	CustomFields map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"custom_fields"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Contact *Contact `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
}
