package postgres

import (
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type AuditRepository struct {
	dbGorm *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{dbGorm: db}
}

func (r *AuditRepository) CreateAuditLog(auditlog *entity.AuditLog) (*entity.AuditLog, error) {
	isCreate := r.dbGorm.Create(auditlog)
	err := isCreate.Error
	if err != nil {
		return nil, err
	}
	return auditlog, nil
}

func (r *AuditRepository) GetAuditLogsByTargetID(targetID string) (*[]entity.AuditLog, error) {
	var logs []entity.AuditLog
	err := r.dbGorm.Where("target_id = ?", targetID).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return &logs, nil
}
