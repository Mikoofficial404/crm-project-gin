package entity

import "time"

type User struct {
	ID                 string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name               string    `json:"name" gorm:"column:name"`
	Email              string    `json:"email" gorm:"column:email"`
	Password           string    `json:"password" gorm:"column:password"`
	Role               string    `json:"role" gorm:"column:role"`
	TwoFactorSecret    string    `json:"two_factor_secret" gorm:"column:two_factor_secret"`
	IsTwoFactorEnabled bool      `json:"is_two_factor_enabled" gorm:"column:is_two_factor_enabled"`
	RefreshToken       string    `json:"refresh_token" gorm:"column:refresh_token"`
	RefreshTokenExpiry time.Time `json:"refresh_token_expiry" gorm:"column:refresh_token_expiry"`
	TeamID             *string   `gorm:"type:uuid;index" json:"team_id"`
	Team               *Team     `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	Lead               []Lead    `gorm:"foreignKey:AssignedTo"`
	Deals              []Deal    `gorm:"foreignKey:AssignedTo"`
}
