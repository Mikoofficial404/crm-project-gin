package entity

import (
	"time"

	"gorm.io/gorm"
)

type Campaign struct {
	ID          string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string     `json:"name" gorm:"not null"`
	Subject     string     `json:"subject" gorm:"not null"`
	Body        string     `json:"body" gorm:"type:text;not null"`
	Status      string     `json:"status" gorm:"type:varchar(20);not null;default:'DRAFT'"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	CreatedBy   string     `json:"created_by" gorm:"type:uuid;not null"`

	// Relasi
	CreatedByUser User                `json:"createdByUser,omitempty" gorm:"foreignKey:CreatedBy"`
	Recipients    []CampaignRecipient `json:"recipients,omitempty" gorm:"foreignKey:CampaignID"`

	// Timestamps
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CampaignRecipient struct {
	ID         string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CampaignID string     `json:"campaignId" gorm:"type:uuid;not null;index"`
	LeadID     *string    `json:"leadId,omitempty" gorm:"type:uuid;index"`
	ContactID  *string    `json:"contactId,omitempty" gorm:"type:uuid;index"`
	Email      string     `json:"email" gorm:"type:varchar(255);not null"`
	Status     string     `json:"status" gorm:"type:varchar(20);not null;default:'PENDING'"`
	SentAt     *time.Time `json:"sentAt,omitempty"`
	ErrorMsg   *string    `json:"errorMsg,omitempty" gorm:"type:text"`

	// Relasi
	Campaign Campaign `json:"campaign,omitempty" gorm:"foreignKey:CampaignID"`
	Lead     *Lead    `json:"lead,omitempty" gorm:"foreignKey:LeadID"`
	Contact  *Contact `json:"contact,omitempty" gorm:"foreignKey:ContactID"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
