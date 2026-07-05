package entity

import (
	"time"

	"gorm.io/gorm"
)

type DealComment struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DealID           string         `gorm:"type:uuid;not null;index"`
	UserID           string         `gorm:"type:uuid;not null"`
	Content          string         `gorm:"type:text;not null"`
	MentionedUserIDs string         `gorm:"type:jsonb;default:'[]';serializer:json"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	Deal Deal `gorm:"foreignKey:DealID"`
	User User `gorm:"foreignKey:UserID"`
}
