package postgres

import (
	"context"
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type DealRepository struct {
	dbGorm *gorm.DB
}

func NewDealRepository(db *gorm.DB) *DealRepository {
	return &DealRepository{dbGorm: db}
}

func (r *DealRepository) CreateDeal(deal *entity.Deal) (*entity.Deal, error) {
	isCreate := r.dbGorm.Create(deal)
	err := isCreate.Error
	if err != nil {
		return nil, err
	}
	return deal, nil
}

func (r *DealRepository) GetDealByID(dealID string) (*entity.Deal, error) {
	ctx := context.Background()
	var deal entity.Deal
	err := r.dbGorm.WithContext(ctx).
		Preload("Lead").
		Preload("Stage").
		Preload("Pipeline").
		Preload("AssignedUser").
		Where("id = ?", dealID).
		First(&deal).Error
	if err != nil {
		return nil, err
	}
	return &deal, nil
}

func (r *DealRepository) GetAllDeals(page int, limit int, search string, stageID string, pipelineID string, startDate, endDate string) ([]entity.Deal, int64, error) {
	var deals []entity.Deal
	var total int64

	query := r.dbGorm.Model(&entity.Deal{}).Preload("Lead").Preload("Stage").Preload("Pipeline").Preload("AssignedUser")
	if search != "" {
		query = query.Where("to_tsvector('simple', coalesce(name,'')) @@ plainto_tsquery('simple', ?)", search)
	}
	if stageID != "" {
		query = query.Where("stage_id = ?", stageID)
	}
	if pipelineID != "" {
		query = query.Where("pipeline_id = ?", pipelineID)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	countQuery := r.dbGorm.Model(&entity.Deal{})
	if search != "" {
		countQuery = countQuery.Where("to_tsvector('simple', coalesce(name,'')) @@ plainto_tsquery('simple', ?)", search)
	}
	if stageID != "" {
		countQuery = countQuery.Where("stage_id = ?", stageID)
	}
	if pipelineID != "" {
		countQuery = countQuery.Where("pipeline_id = ?", pipelineID)
	}
	if startDate != "" {
		countQuery = countQuery.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		countQuery = countQuery.Where("created_at <= ?", endDate+" 23:59:59")
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("position ASC").Find(&deals).Error
	if err != nil {
		return nil, 0, err
	}
	return deals, total, nil
}

func (r *DealRepository) GetDealByUserId(userId string, page int, limit int, search string, stageID string, pipelineID string, startDate, endDate string) ([]entity.Deal, int64, error) {
	var deals []entity.Deal
	var total int64

	allowedUserIDsQuery := r.dbGorm.Model(&entity.User{}).Select("id").Where("id = ? OR team_id IN (SELECT id FROM teams WHERE manager_id = ?)", userId, userId)
	query := r.dbGorm.Model(&entity.Deal{}).Preload("Lead").Preload("Stage").Preload("Pipeline").Preload("AssignedUser").Where("assigned_to IN (?)", allowedUserIDsQuery)
	if search != "" {
		query = query.Where("to_tsvector('simple', coalesce(name,'')) @@ plainto_tsquery('simple', ?)", search)
	}
	if stageID != "" {
		query = query.Where("stage_id = ?", stageID)
	}
	if pipelineID != "" {
		query = query.Where("pipeline_id = ?", pipelineID)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	countQuery := r.dbGorm.Model(&entity.Deal{}).Where("assigned_to IN (?)", allowedUserIDsQuery)
	if search != "" {
		countQuery = countQuery.Where("to_tsvector('simple', coalesce(name,'')) @@ plainto_tsquery('simple', ?)", search)
	}
	if stageID != "" {
		countQuery = countQuery.Where("stage_id = ?", stageID)
	}
	if pipelineID != "" {
		countQuery = countQuery.Where("pipeline_id = ?", pipelineID)
	}
	if startDate != "" {
		countQuery = countQuery.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		countQuery = countQuery.Where("created_at <= ?", endDate+" 23:59:59")
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("position ASC").Find(&deals).Error
	if err != nil {
		return nil, 0, err
	}
	return deals, total, nil
}

func (r *DealRepository) UpdateAssignedTo(dealID string, assignedTo string) error {
	return r.dbGorm.Model(&entity.Deal{}).Where("id = ?", dealID).Update("assigned_to", assignedTo).Error
}

func (r *DealRepository) UpdateStage(dealID string, stageID string) error {
	err := r.dbGorm.Model(&entity.Deal{}).Where("id = ?", dealID).Update("stage_id", stageID).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *DealRepository) SoftDeleteDeal(dealID string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Where("id = ?", dealID).Delete(&entity.Deal{}).Error
	return err
}

func (r *DealRepository) GetDealsWithActiveSLA() ([]entity.Deal, error) {
	var deals []entity.Deal
	err := r.dbGorm.Preload("Stage").
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id").
		Where("pipeline_stages.sla_hours > 0 AND deals.sla_breached = false").
		Find(&deals).Error
	return deals, err
}

func (r *DealRepository) MarkSLABreached(dealID string) error {
	return r.dbGorm.Model(&entity.Deal{}).
		Where("id = ?", dealID).
		Update("sla_breached", true).Error
}

func (r *DealRepository) SearchDeals(keyword string) ([]entity.Deal, error) {
	var deals []entity.Deal
	err := r.dbGorm.
		Where("to_tsvector('simple', coalesce(name,'')) @@ plainto_tsquery('simple', ?)", keyword).
		Find(&deals).Error
	if err != nil {
		return nil, err
	}
	return deals, nil
}
func (r *DealRepository) UpdateDealPositions(dealIDs []string) error {
	tx := r.dbGorm.Begin()
	for index, id := range dealIDs {
		err := tx.Model(&entity.Deal{}).Where("id = ?", id).Update("position", index).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *DealRepository) GetTrashedDeals(userID, role string) ([]entity.Deal, error) {
	var deals []entity.Deal
	query := r.dbGorm.Unscoped().Where("deleted_at IS NOT NULL")
	if role == "sales" {
		query = query.Where("assigned_to = ?", userID)
	}
	err := query.Find(&deals).Error
	if err != nil {
		return nil, err
	}
	return deals, nil
}

func (r *DealRepository) RestoreDeal(dealID string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Unscoped().Where("id = ?", dealID).Update("deleted_at", nil).Error
	return err
}

func (r *DealRepository) UpdateDeal(dealID string, name string, value float64) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Model(&entity.Deal{}).Where("id = ?", dealID).Updates(map[string]interface{}{
		"name":  name,
		"value": value,
	}).Error
	return err
}
