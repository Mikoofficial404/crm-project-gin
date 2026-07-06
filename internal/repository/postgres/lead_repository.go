package postgres

import (
	"context"
	"crm-project/internal/models/entity"
	"sort"
	"time"

	"gorm.io/gorm"
)

type LeadRepository struct {
	dbGorm *gorm.DB
}

func NewLeadRepository(db *gorm.DB) *LeadRepository {
	return &LeadRepository{dbGorm: db}
}

type TimelineItem struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

func (r *LeadRepository) CreateLead(lead *entity.Lead) (*entity.Lead, error) {
	isCreate := r.dbGorm.Create(lead)
	err := isCreate.Error
	if err != nil {
		return nil, err
	}
	return lead, nil
}

func (r *LeadRepository) GetAllLeads(page int, limit int, search string, status string, startDate string, endDate string) ([]entity.Lead, int64, error) {
	var leads []entity.Lead
	var total int64

	query := r.dbGorm.Model(&entity.Lead{}).Preload("Contact")
	if search != "" {
		query = query.Where("to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(email,'') || ' ' || coalesce(phone,'')) @@ plainto_tsquery('simple', ?)", search)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&leads).Error
	if err != nil {
		return nil, 0, err
	}
	return leads, total, nil
}

func (r *LeadRepository) GetLeadsByUserId(userID string, page int, limit int, search string, status string, startDate string, endDate string) ([]entity.Lead, int64, error) {
	ctx := context.Background()
	var leads []entity.Lead
	var total int64

	query := r.dbGorm.WithContext(ctx).Model(&entity.Lead{}).Preload("Contact").Where("assigned_to = ?", userID)
	if search != "" {
		query = query.Where("to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(email,'') || ' ' || coalesce(phone,'')) @@ plainto_tsquery('simple', ?)", search)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&leads).Error
	if err != nil {
		return nil, 0, err
	}
	return leads, total, nil
}

func (r *LeadRepository) UpdateStatus(leadID string, status string) error {
	ctx := context.Background()
	var leads entity.Lead
	err := r.dbGorm.WithContext(ctx).Model(&leads).Where("id = ?", leadID).Update("status", status).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *LeadRepository) GetLeadByID(leadID string) (*entity.Lead, error) {
	ctx := context.Background()
	var lead entity.Lead
	err := r.dbGorm.WithContext(ctx).Preload("Contact").Where("id = ?", leadID).First(&lead).Error
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

func (r *LeadRepository) GetLeadByPhone(phone string) (*entity.Lead, error) {
	ctx := context.Background()
	var lead entity.Lead
	err := r.dbGorm.WithContext(ctx).Preload("Contact").Where("phone = ?", phone).First(&lead).Error
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

func (r *LeadRepository) SoftDeleteLead(leadID string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Where("id = ?", leadID).Delete(&entity.Lead{}).Error
	return err
}

func (r *LeadRepository) GetLeadTimeline(leadID string) (*[]TimelineItem, error) {
	var timeLines []TimelineItem
	var activities []entity.Activity
	if err := r.dbGorm.Where("lead_id = ?", leadID).Find(&activities).Error; err != nil {
		return nil, err
	}
	for _, activity := range activities {
		timeLines = append(timeLines, TimelineItem{
			Type:      "activity",
			Data:      activity,
			Timestamp: activity.CreatedAt,
		})
	}

	var deals []entity.Deal
	if err := r.dbGorm.Where("lead_id = ?", leadID).Find(&deals).Error; err != nil {
		return nil, err
	}
	for _, deal := range deals {
		timeLines = append(timeLines, TimelineItem{
			Type:      "deal",
			Data:      deal,
			Timestamp: deal.CreatedAt,
		})
	}

	var tasks []entity.Task
	if err := r.dbGorm.Where("lead_id = ?", leadID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	for _, task := range tasks {
		timeLines = append(timeLines, TimelineItem{
			Type:      "task",
			Data:      task,
			Timestamp: task.CreatedAt,
		})
	}
	sort.Slice(timeLines, func(i, j int) bool {
		return timeLines[i].Timestamp.After(timeLines[j].Timestamp)
	})
	return &timeLines, nil
}

func (r *LeadRepository) CreateBulkLeads(leadEntity *[]entity.Lead) ([]entity.Lead, error) {
	csvInsert := r.dbGorm.Create(&leadEntity)
	err := csvInsert.Error
	if err != nil {
		return nil, err
	}
	return *leadEntity, nil
}

func (r *LeadRepository) GetTrashedLeads(userID string, role string) (*[]entity.Lead, error) {
	var leads []entity.Lead
	query := r.dbGorm.Unscoped().Where("deleted_at IS NOT NULL")
	if role == "sales" {
		query = query.Where("assigned_to = ?", userID)
	}
	err := query.Find(&leads).Error
	if err != nil {
		return nil, err
	}
	return &leads, nil
}

func (r *LeadRepository) RestoreLead(leadID string) error {
	var leads entity.Lead
	err := r.dbGorm.Unscoped().Model(&leads).Where("id = ?", leadID).Update("deleted_at", nil).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *LeadRepository) SearchLeads(keyword string) ([]entity.Lead, error) {
	ctx := context.Background()
	var leads []entity.Lead
	err := r.dbGorm.WithContext(ctx).Preload("Contact").Where("to_tsvector('simple', name || ' ' || COALESCE(email, '') || ' ' || COALESCE(phone, '')) @@ plainto_tsquery('simple', ?)", keyword).Find(&leads).Error
	if err != nil {
		return nil, err
	}
	return leads, nil
}

func (r *LeadRepository) UpdateLead(leadID string, name string, email string, phone string) error {
	ctx := context.Background()
	err := r.dbGorm.WithContext(ctx).Model(&entity.Lead{}).Where("id = ?", leadID).Updates(map[string]interface{}{
		"name":  name,
		"email": email,
		"phone": phone,
	}).Error
	return err
}

func (r *LeadRepository) GetAgingLeads(userID, role string, days int) ([]entity.Lead, error) {
	var leads []entity.Lead
	query := r.dbGorm.Model(&entity.Lead{}).
		Preload("Contact").
		Where("status NOT IN ? AND updated_at < NOW() - (? * INTERVAL '1 day')", []string{"WON", "LOST"}, days).
		Order("updated_at ASC")

	if role == "sales" {
		query = query.Where("assigned_to = ?", userID)
	}

	err := query.Find(&leads).Error
	if err != nil {
		return nil, err
	}
	return leads, nil
}
