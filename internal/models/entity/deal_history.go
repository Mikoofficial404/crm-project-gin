package entity

import (
	"time"
)

type DealHistory struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DealID       string    `gorm:"type:uuid;not null;index"`
	UserID       string    `gorm:"type:uuid;not null"`
	FieldChanged string    `gorm:"type:varchar(50);not null"`
	OldValue     string    `gorm:"type:text"`
	NewValue     string    `gorm:"type:text"`
	ChangedAt    time.Time `gorm:"autoCreateTime"`

	Deal Deal `gorm:"foreignKey:DealID"`
	User User `gorm:"foreignKey:UserID"`
}
