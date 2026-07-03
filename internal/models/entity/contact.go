package entity

import (
	"time"

	"gorm.io/gorm"
)

type Contact struct {
	ID           string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name         string         `json:"name" gorm:"type:varchar(255);not null"`
	Email        *string        `json:"email,omitempty" gorm:"type:varchar(255);uniqueIndex"`
	Phone        string         `json:"phone" gorm:"type:varchar(50);uniqueIndex;not null"`
	Company      *string        `json:"company,omitempty" gorm:"type:varchar(255)"`
	Position     *string        `json:"position,omitempty" gorm:"type:varchar(255)"`
	Source       string         `json:"source" gorm:"type:varchar(50);not null;default:'manual'"`
	AssignedTo   string         `json:"assigned_to" gorm:"type:uuid;not null"`
	AssignedUser User           `json:"assigned_user,omitempty" gorm:"foreignKey:AssignedTo;references:ID"`
	Leads        []Lead         `json:"leads,omitempty" gorm:"foreignKey:ContactID"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
