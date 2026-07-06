package entity

import "time"

type LoginHistory struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	IPAddress string    `gorm:"type:varchar(45);not null"`
	UserAgent string    `gorm:"type:text"`
	LoginAt   time.Time `gorm:"autoCreateTime"`
	Success   bool      `gorm:"not null;default:false"`

	User User `gorm:"foreignKey:UserID"`
}
