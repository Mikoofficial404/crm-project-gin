package entity

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID           string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title        string         `json:"title" gorm:"column:title;not null"`
	Description  *string        `json:"description,omitempty" gorm:"column:description"`
	DueDate      *time.Time     `json:"due_date,omitempty" gorm:"column:due_date"`
	Status       string         `json:"status" gorm:"column:status;type:varchar(20);not null;default:'PENDING';check:status IN ('PENDING','DONE')"`
	Priority     string         `json:"priority" gorm:"column:priority;type:varchar(20);not null;default:'MEDIUM';check:priority IN ('LOW','MEDIUM','HIGH','URGENT')"`
	Category     string         `json:"category" gorm:"column:category;type:varchar(20);not null;default:'OTHER';check:category IN ('FOLLOW_UP','MEETING','CALL','EMAIL','OTHER')"`
	AssignedTo   string         `json:"assigned_to" gorm:"column:assigned_to;type:uuid;not null"`
	LeadID       *string        `json:"lead_id,omitempty" gorm:"column:lead_id;type:uuid"`
	ReminderSent bool           `json:"reminder_sent" gorm:"column:reminder_sent;not null;default:false"`
	CreatedAt    time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index"`
}
