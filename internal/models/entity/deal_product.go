package entity

import "time"

type DealProduct struct {
	ID        string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DealID    string  `json:"deal_id" gorm:"type:uuid;not null;index"`
	ProductID string  `json:"product_id" gorm:"type:uuid;not null;index"`
	Quantity  int     `json:"quantity" gorm:"not null;default:1"`
	UnitPrice float64 `json:"unit_price" gorm:"type:numeric(15,2);not null"`
	SubTotal  float64 `json:"sub_total" gorm:"type:numeric(15,2);not null"`

	// Relasi
	Deal    Deal    `json:"deal,omitempty" gorm:"foreignKey:DealID"`
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
