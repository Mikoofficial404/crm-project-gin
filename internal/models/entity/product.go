package entity

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string  `json:"name" gorm:"type:varchar(100);not null"`
	Description string  `json:"description" gorm:"type:text"`
	Price       float64 `json:"price" gorm:"type:numeric(15,2);not null;default:0"`
	Unit        string  `json:"unit" gorm:"type:varchar(30);not null;default:'pcs'"` // pcs, jam, bulan, dll
	IsActive    bool    `json:"is_active" gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
