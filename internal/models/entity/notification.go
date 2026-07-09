package entity

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User      User      `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Title     string    `json:"title" gorm:"type:varchar(255);not null"`
	Message   string    `json:"message" gorm:"type:text;not null"`
	Link      string    `json:"link" gorm:"type:varchar(255)"`
	IsRead    bool      `json:"is_read" gorm:"default:false;not null"`
	CreatedAt time.Time `json:"CreatedAt" gorm:"autoCreateTime"`
}
