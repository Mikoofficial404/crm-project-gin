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

func (r *DashboardRepository) GetStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	if err := r.dbGorm.Model(&entity.Lead{}).Count(&stats.TotalLeads).Error; err != nil {
		return nil, err
	}

	if err := r.dbGorm.Model(&entity.Deal{}).Count(&stats.TotalDeals).Error; err != nil {
		return nil, err
	}

	if err := r.dbGorm.Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ?", true).
		Count(&stats.TotalWonDeals).Error; err != nil {
		return nil, err
	}

	if err := r.dbGorm.Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ?", true).
		Select("COALESCE(SUM(value), 0)").
		Scan(&stats.TotalRevenueWon).Error; err != nil {
		return nil, err
	}

	if err := r.dbGorm.Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ? AND pipeline_stages.is_closed_lost = ?", false, false).
		Select("COALESCE(SUM(value), 0)").
		Scan(&stats.PotentialRevenue).Error; err != nil {
		return nil, err
	}

	if err := r.dbGorm.Model(&entity.Contact{}).Count(&stats.TotalContacts).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *DashboardRepository) GetTotalRevenue() (float64, error) {
	var total float64
	err := r.dbGorm.Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ?", true).
		Select("COALESCE(SUM(value), 0)").
		Scan(&total).Error
	return total, err
}

func (r *DashboardRepository) GetProjectedRevenue() (float64, error) {
	var total float64
	err := r.dbGorm.Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ? AND pipeline_stages.is_closed_lost = ?", false, false).
		Select("COALESCE(SUM(value), 0)").
		Scan(&total).Error
	return total, err
}

func (r *DashboardRepository) GetTotalWonDeals() (int64, error) {
	var total int64
	err := r.dbGorm.Model(&entity.Deal{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("pipeline_stages.is_closed_won = ?", true).
		Count(&total).Error
	return total, err
}

func (r *DashboardRepository) GetTotalLeads() (int64, error) {
	var total int64
	err := r.dbGorm.Model(&entity.Lead{}).Count(&total).Error
	return total, err
}

func (r *DashboardRepository) GetTotalContacts() (int64, error) {
	var total int64
	err := r.dbGorm.Model(&entity.Contact{}).Count(&total).Error
	return total, err
}
