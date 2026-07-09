package entity

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                 string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name               string         `json:"name" gorm:"column:name"`
	Email              string         `json:"email" gorm:"column:email"`
	Password           string         `json:"password" gorm:"column:password"`
	Role               string         `json:"role" gorm:"column:role"`
	IsActive           bool           `json:"is_active" gorm:"column:is_active;default:true"`
	IsOnline           bool           `json:"is_online" gorm:"column:is_online;default:true"`
	TwoFactorSecret    string         `json:"two_factor_secret" gorm:"column:two_factor_secret"`
	IsTwoFactorEnabled bool           `json:"is_two_factor_enabled" gorm:"column:is_two_factor_enabled"`
	RefreshToken       string         `json:"refresh_token" gorm:"column:refresh_token"`
	RefreshTokenExpiry time.Time      `json:"refresh_token_expiry" gorm:"column:refresh_token_expiry"`
	TeamID             *string        `gorm:"type:uuid;index" json:"team_id"`
	Team               *Team          `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Lead               []Lead         `gorm:"foreignKey:AssignedTo"`
	Deals              []Deal         `gorm:"foreignKey:AssignedTo"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
	LastLogin          *time.Time     `json:"last_login"`
}
