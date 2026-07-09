package postgres

import (
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type DashboardRepository struct {
	dbGorm *gorm.DB
}

type DashboardStats struct {
	TotalLeads       int64   `json:"total_leads"`
	TotalDeals       int64   `json:"total_deals"`
	TotalWonDeals    int64   `json:"total_won_deals"`
	TotalRevenueWon  float64 `json:"total_revenue_won"`
	PotentialRevenue float64 `json:"potential_revenue"`
	TotalContacts    int64   `json:"total_contacts"`
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{
		dbGorm: db,
	}
}

func (r *DashboardRepository) getDbWithRole(role, userID string) *gorm.DB {
	if role == "sales" && userID != "" {
		return r.dbGorm.Where("assigned_to = ?", userID)
	}
	return r.dbGorm
}

func (r *DashboardRepository) GetStats(role, userID string) (*DashboardStats, error) {
	stats := &DashboardStats{}

	if err := r.getDbWithRole(role, userID).Model(&entity.Lead{}).Count(&stats.TotalLeads).Error; err != nil {
		return nil, err
	}

	if err := r.getDbWithRole(role, userID).Model(&entity.Deal{}).Count(&stats.TotalDeals).Error; err != nil {
		return nil, err
	}

	if err := r.getDbWithRole(role, userID).Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ?", true).
		Count(&stats.TotalWonDeals).Error; err != nil {
		return nil, err
	}

	if err := r.getDbWithRole(role, userID).Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ?", true).
		Select("COALESCE(SUM(value), 0)").
		Scan(&stats.TotalRevenueWon).Error; err != nil {
		return nil, err
	}

	if err := r.getDbWithRole(role, userID).Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ? AND pipeline_stages.is_closed_lost = ?", false, false).
		Select("COALESCE(SUM(value), 0)").
		Scan(&stats.PotentialRevenue).Error; err != nil {
		return nil, err
	}

	if err := r.getDbWithRole(role, userID).Model(&entity.Contact{}).Count(&stats.TotalContacts).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *DashboardRepository) GetTotalRevenue(role, userID string) (float64, error) {
	var total float64
	err := r.getDbWithRole(role, userID).Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ?", true).
		Select("COALESCE(SUM(value), 0)").
		Scan(&total).Error
	return total, err
}

func (r *DashboardRepository) GetProjectedRevenue(role, userID string) (float64, error) {
	var total float64
	err := r.getDbWithRole(role, userID).Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ? AND pipeline_stages.is_closed_lost = ?", false, false).
		Select("COALESCE(SUM(value), 0)").
		Scan(&total).Error
	return total, err
}

func (r *DashboardRepository) GetTotalWonDeals(role, userID string) (int64, error) {
	var total int64
	err := r.getDbWithRole(role, userID).Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ?", true).
		Count(&total).Error
	return total, err
}

func (r *DashboardRepository) GetTotalLeads(role, userID string) (int64, error) {
	var total int64
	err := r.getDbWithRole(role, userID).Model(&entity.Lead{}).Count(&total).Error
	return total, err
}

func (r *DashboardRepository) GetTotalContacts(role, userID string) (int64, error) {
	var total int64
	err := r.getDbWithRole(role, userID).Model(&entity.Contact{}).Count(&total).Error
	return total, err
}
