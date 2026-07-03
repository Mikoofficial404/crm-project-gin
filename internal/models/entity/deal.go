package entity

import (
	"time"

	"gorm.io/gorm"
)

type Deal struct {
	ID         string  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name       string  `json:"name" gorm:"not null"`
	Value      float64 `json:"value"`
	PipelineID string  `json:"pipelineId" gorm:"type:uuid;not null;index"`
	StageID    string  `json:"stageId" gorm:"type:uuid;not null;index"`
	LeadID     string  `json:"leadId" gorm:"type:uuid;not null;index"`
	AssignedTo string  `json:"assignedTo" gorm:"type:uuid;not null"`
	Position   int     `json:"position"`

	// Relasi
	Pipeline     Pipeline      `json:"pipeline,omitempty" gorm:"foreignKey:PipelineID"`
	Stage        PipelineStage `json:"stage,omitempty" gorm:"foreignKey:StageID"`
	Lead         Lead          `json:"lead,omitempty" gorm:"foreignKey:LeadID"`
	AssignedUser User          `json:"assignedUser,omitempty" gorm:"foreignKey:AssignedTo"`

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
