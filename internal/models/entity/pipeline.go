package entity

import (
	"time"

	"gorm.io/gorm"
)

type Pipeline struct {
	ID          string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default" gorm:"default:false"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	CreatedBy   string `json:"created_by" gorm:"type:uuid;not null"`

	// Relasi
	CreatedByUser User            `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
	Stages        []PipelineStage `json:"stages,omitempty" gorm:"foreignKey:PipelineID"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type PipelineStage struct {
	ID           string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PipelineID   string `json:"pipeline_id" gorm:"type:uuid;not null;index"`
	Name         string `json:"name" gorm:"not null"`
	StageOrder   int    `json:"stage_order" gorm:"column:stage_order;not null;default:0"`
	Color        string `json:"color" gorm:"type:varchar(20)"`
	IsClosedWon  bool   `json:"is_closed_won" gorm:"default:false"`
	IsClosedLost bool   `json:"is_closed_lost" gorm:"default:false"`
	Probability  int    `json:"probability" gorm:"not null;default:0;check:probability >= 0 AND probability <= 100"`
	SLAHours     int    `json:"sla_hours" gorm:"not null;default:0"`

	// Relasi
	Pipeline Pipeline `json:"pipeline,omitempty" gorm:"foreignKey:PipelineID"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type StageOrder struct {
	StageID string `json:"stage_id"`
	Order   int    `json:"order"`
}
