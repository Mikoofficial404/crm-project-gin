package postgres

import (
	"time"

	"gorm.io/gorm"
)

type ReportRepository struct {
	dbgorm *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{dbgorm: db}
}

type SalesSummaryResult struct {
	TotalDeals     int64   `json:"total_deals"`
	TotalValue     float64 `json:"total_value"`
	DealsWon       int64   `json:"deals_won"`
	DealsLost      int64   `json:"deals_lost"`
	DealsActive    int64   `json:"deals_active"`
	WinRate        float64 `json:"win_rate"`
	TotalLeads     int64   `json:"total_leads"`
	NewLeads       int64   `json:"new_leads"`
	ConvertedLeads int64   `json:"converted_leads"`
}

func (r *ReportRepository) GetSalesSummary(startDate, endDate time.Time) (*SalesSummaryResult, error) {
	result := &SalesSummaryResult{}

	r.dbgorm.Table("deals").
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&result.TotalDeals)

	r.dbgorm.Table("deals").
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(SUM(value), 0)").
		Scan(&result.TotalValue)

	r.dbgorm.Table("deals").
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("deals.deleted_at IS NULL AND deals.created_at BETWEEN ? AND ? AND pipeline_stages.is_closed_won = true", startDate, endDate).
		Count(&result.DealsWon)

	r.dbgorm.Table("deals").
		Joins("JOIN pipeline_stages ON pipeline_stages.id = deals.stage_id AND pipeline_stages.deleted_at IS NULL").
		Where("deals.deleted_at IS NULL AND deals.created_at BETWEEN ? AND ? AND pipeline_stages.is_closed_lost = true", startDate, endDate).
		Count(&result.DealsLost)

	result.DealsActive = result.TotalDeals - result.DealsWon - result.DealsLost

	if result.TotalDeals > 0 {
		result.WinRate = float64(result.DealsWon) / float64(result.TotalDeals) * 100
	}

	r.dbgorm.Table("leads").
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&result.TotalLeads)

	r.dbgorm.Table("leads").
		Where("deleted_at IS NULL AND status = 'NEW' AND created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&result.NewLeads)

	r.dbgorm.Table("leads").
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ? AND id IN (SELECT DISTINCT lead_id FROM deals WHERE deleted_at IS NULL)", startDate, endDate).
		Count(&result.ConvertedLeads)

	return result, nil
}

type PipelineStageReport struct {
	PipelineID   string  `json:"pipeline_id"`
	PipelineName string  `json:"pipeline_name"`
	StageID      string  `json:"stage_id"`
	StageName    string  `json:"stage_name"`
	StageOrder   int     `json:"stage_order"`
	Color        string  `json:"color"`
	IsClosedWon  bool    `json:"is_closed_won"`
	IsClosedLost bool    `json:"is_closed_lost"`
	DealCount    int64   `json:"deal_count"`
	TotalValue   float64 `json:"total_value"`
}

func (r *ReportRepository) GetPipelineReport(startDate, endDate time.Time, pipelineID string) ([]PipelineStageReport, error) {
	var results []PipelineStageReport

	query := r.dbgorm.Table("pipeline_stages").
		Select(`
			pipelines.id AS pipeline_id,
			pipelines.name AS pipeline_name,
			pipeline_stages.id AS stage_id,
			pipeline_stages.name AS stage_name,
			pipeline_stages.stage_order,
			pipeline_stages.color,
			pipeline_stages.is_closed_won,
			pipeline_stages.is_closed_lost,
			COUNT(deals.id) AS deal_count,
			COALESCE(SUM(deals.value), 0) AS total_value
		`).
		Joins("JOIN pipelines ON pipelines.id = pipeline_stages.pipeline_id AND pipelines.deleted_at IS NULL").
		Joins("LEFT JOIN deals ON deals.stage_id = pipeline_stages.id AND deals.deleted_at IS NULL AND deals.created_at BETWEEN ? AND ?", startDate, endDate).
		Where("pipeline_stages.deleted_at IS NULL").
		Group("pipelines.id, pipelines.name, pipeline_stages.id, pipeline_stages.name, pipeline_stages.stage_order, pipeline_stages.color, pipeline_stages.is_closed_won, pipeline_stages.is_closed_lost").
		Order("pipelines.name, pipeline_stages.stage_order ASC")

	if pipelineID != "" {
		query = query.Where("pipelines.id = ?", pipelineID)
	}

	err := query.Scan(&results).Error
	return results, err
}

type SalesPerformanceResult struct {
	UserID          string  `json:"user_id"`
	UserName        string  `json:"user_name"`
	UserEmail       string  `json:"user_email"`
	TotalDeals      int64   `json:"total_deals"`
	DealsWon        int64   `json:"deals_won"`
	DealsLost       int64   `json:"deals_lost"`
	TotalValue      float64 `json:"total_value"`
	WonValue        float64 `json:"won_value"`
	WinRate         float64 `json:"win_rate"`
	TotalActivities int64   `json:"total_activities"`
	TotalLeads      int64   `json:"total_leads"`
}

func (r *ReportRepository) GetSalesPerformance(startDate, endDate time.Time) ([]SalesPerformanceResult, error) {
	var results []SalesPerformanceResult

	err := r.dbgorm.Table("users").
		Select(`
			users.id AS user_id,
			users.name AS user_name,
			users.email AS user_email,
			COUNT(DISTINCT deals.id) AS total_deals,
			COUNT(DISTINCT CASE WHEN ps_won.is_closed_won = true THEN deals.id END) AS deals_won,
			COUNT(DISTINCT CASE WHEN ps_lost.is_closed_lost = true THEN deals.id END) AS deals_lost,
			COALESCE(SUM(DISTINCT deals.value), 0) AS total_value,
			COALESCE(SUM(DISTINCT CASE WHEN ps_won.is_closed_won = true THEN deals.value ELSE 0 END), 0) AS won_value,
			COUNT(DISTINCT activities.id) AS total_activities,
			COUNT(DISTINCT leads.id) AS total_leads
		`).
		Joins("LEFT JOIN deals ON deals.assigned_to = users.id AND deals.deleted_at IS NULL AND deals.created_at BETWEEN ? AND ?", startDate, endDate).
		Joins("LEFT JOIN pipeline_stages ps_won ON ps_won.id = deals.stage_id AND ps_won.is_closed_won = true AND ps_won.deleted_at IS NULL").
		Joins("LEFT JOIN pipeline_stages ps_lost ON ps_lost.id = deals.stage_id AND ps_lost.is_closed_lost = true AND ps_lost.deleted_at IS NULL").
		Joins("LEFT JOIN activities ON activities.assigned_to = users.id AND activities.deleted_at IS NULL AND activities.created_at BETWEEN ? AND ?", startDate, endDate).
		Joins("LEFT JOIN leads ON leads.assigned_to = users.id AND leads.deleted_at IS NULL AND leads.created_at BETWEEN ? AND ?", startDate, endDate).
		Where("users.role = 'sales'").
		Group("users.id, users.name, users.email").
		Order("won_value DESC").
		Scan(&results).Error

	for i := range results {
		if results[i].TotalDeals > 0 {
			results[i].WinRate = float64(results[i].DealsWon) / float64(results[i].TotalDeals) * 100
		}
	}

	return results, err
}

type LeadSourceResult struct {
	Source      string  `json:"source"`
	TotalLeads  int64   `json:"total_leads"`
	Converted   int64   `json:"converted"`
	ConvertRate float64 `json:"convert_rate"`
}

func (r *ReportRepository) GetLeadSourceReport(startDate, endDate time.Time) ([]LeadSourceResult, error) {
	var results []LeadSourceResult

	err := r.dbgorm.Table("contacts").
		Select(`
			contacts.source,
			COUNT(DISTINCT leads.id) AS total_leads,
			COUNT(DISTINCT CASE WHEN deals.id IS NOT NULL THEN leads.id END) AS converted
		`).
		Joins("LEFT JOIN leads ON leads.contact_id = contacts.id AND leads.deleted_at IS NULL AND leads.created_at BETWEEN ? AND ?", startDate, endDate).
		Joins("LEFT JOIN deals ON deals.lead_id = leads.id AND deals.deleted_at IS NULL").
		Where("contacts.deleted_at IS NULL").
		Group("contacts.source").
		Order("total_leads DESC").
		Scan(&results).Error

	for i := range results {
		if results[i].TotalLeads > 0 {
			results[i].ConvertRate = float64(results[i].Converted) / float64(results[i].TotalLeads) * 100
		}
	}

	return results, err
}

type ActivityReportResult struct {
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	ActivityType string `json:"activity_type"`
	TotalCount   int64  `json:"total_count"`
}

type ActivitySummaryResult struct {
	UserID          string           `json:"user_id"`
	UserName        string           `json:"user_name"`
	UserEmail       string           `json:"user_email"`
	TotalActivities int64            `json:"total_activities"`
	ByType          map[string]int64 `json:"by_type"`
}

func (r *ReportRepository) GetActivityReport(startDate, endDate time.Time) ([]ActivityReportResult, error) {
	var results []ActivityReportResult

	err := r.dbgorm.Table("activities").
		Select(`
			users.id AS user_id,
			users.name AS user_name,
			activities.type AS activity_type,
			COUNT(activities.id) AS total_count
		`).
		Joins("JOIN users ON users.id = activities.assigned_to").
		Where("activities.deleted_at IS NULL AND activities.created_at BETWEEN ? AND ?", startDate, endDate).
		Group("users.id, users.name, activities.type").
		Order("users.name, total_count DESC").
		Scan(&results).Error

	return results, err
}
