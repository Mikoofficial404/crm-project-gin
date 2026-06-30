package entity

import (
	"time"

	"gorm.io/gorm"
)

type Deal struct {
	ID         string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name       string  `json:"name" gorm:"column:name"`
	Value      float64 `json:"value" gorm:"column:value"`
	Stage      string  `json:"stage" gorm:"size:20;not null;default:PROSPECTING"`
	LeadID     string  `json:"leadId" gorm:"type:uuid;not null"`
	Lead       Lead    `json:"lead" gorm:"foreignKey:LeadID"`
	AssignedTo string  `json:"assignedTo" gorm:"type:uuid;not null"`
	Position   int     `json:"position" gorm:"column:position"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}
