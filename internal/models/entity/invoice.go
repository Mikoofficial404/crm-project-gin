package entity

import (
	"time"

	"gorm.io/gorm"
)

type Invoice struct {
	ID         string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	InvoiceNo  string  `json:"invoiceNo" gorm:"type:varchar(50);uniqueIndex;not null"`
	DealID     string  `json:"dealId" gorm:"type:uuid;not null;index"`
	SubTotal   float64 `json:"subTotal" gorm:"type:numeric(15,2);not null"`
	Tax        float64 `json:"tax" gorm:"type:numeric(15,2);not null"`
	GrandTotal float64 `json:"grandTotal" gorm:"type:numeric(15,2);not null"`
	Status     string  `json:"status" gorm:"type:varchar(20);not null;default:'UNPAID'"`

	// Relasi
	Deal  Deal          `json:"deal,omitempty" gorm:"foreignKey:DealID"`
	Items []InvoiceItem `json:"items,omitempty" gorm:"foreignKey:InvoiceID"`

	// Timestamps
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

type InvoiceItem struct {
	ID        string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	InvoiceID string  `json:"invoiceId" gorm:"type:uuid;not null;index"`
	ProductID string  `json:"productId" gorm:"type:uuid;not null;index"`
	Quantity  int     `json:"quantity" gorm:"not null;default:1"`
	UnitPrice float64 `json:"unitPrice" gorm:"type:numeric(15,2);not null"`
	SubTotal  float64 `json:"subTotal" gorm:"type:numeric(15,2);not null"`

	// Relasi
	Invoice Invoice `json:"invoice,omitempty" gorm:"foreignKey:InvoiceID"`
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`

	// Timestamps
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}
