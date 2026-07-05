package entity

import (
	"time"

	"gorm.io/gorm"
)

type Team struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null"                     json:"name"`
	Description string         `gorm:"type:text"                                      json:"description"`
	ManagerID   string         `gorm:"type:uuid;not null"                             json:"manager_id"`
	Manager     *User          `gorm:"foreignKey:ManagerID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"manager,omitempty"`
	Members     []User         `gorm:"foreignKey:TeamID"                              json:"members,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                          json:"-"`
}
