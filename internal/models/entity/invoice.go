package entity

import (
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	InvoiceNo  string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"invoice_no"`
	DealID     string    `gorm:"type:varchar(50);not null;index" json:"deal_id"`
	SubTotal   float64   `gorm:"type:numeric(15,2);not null" json:"sub_total"`
	Tax        float64   `gorm:"type:numeric(15,2);not null" json:"tax"`
	GrandTotal float64   `gorm:"type:numeric(15,2);not null" json:"grand_total"`
	Status     string    `gorm:"type:varchar(20);not null;default:'UNPAID'" json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
