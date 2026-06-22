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
	TotalRevenueWon  float64 `json:"total_revenue_won"`
	PotentialRevenue float64 `json:"potential_revenue"`
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
		Where("stage = ?", "WON").
		Select("COALESCE(SUM(value), 0)").
		Scan(&stats.TotalRevenueWon).Error; err != nil {
		return nil, err
	}
	
	if err := r.dbGorm.Model(&entity.Deal{}).
		Where("stage = ?", "PROSPECTING").
		Select("COALESCE(SUM(value), 0)").
		Scan(&stats.PotentialRevenue).Error; err != nil {
		return nil, err
	}
	
	return stats, nil
}
