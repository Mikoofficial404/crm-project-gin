package entity

import (
	"time"

	"gorm.io/gorm"
)

type Activity struct {
	ID         string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Type       string         `json:"type" gorm:"type:varchar(20);not null"`
	Notes      string         `json:"notes" gorm:"type:text"`
	LeadID     string         `json:"lead_id" gorm:"not null"`
	AssignedTo string         `json:"assigned_to" gorm:"not null"`
	Attachment string         `json:"attachment" gorm:"default:null"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}
