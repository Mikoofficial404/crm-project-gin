package entity

import "time"

type AuditLog struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserIDAudit string    `gorm:"not null;index" json:"user_id_audit"`
	Action      string    `gorm:"type:varchar(255);not null" json:"action"`
	TargetID    string    `gorm:"not null" json:"target_id"`
	OldData     string    `gorm:"not null" json:"old_data"`
	NewData     string    `gorm:"not null" json:"new_data"`
	Waktu       time.Time `gorm:"autoCreateTime"           json:"waktu"`
}
