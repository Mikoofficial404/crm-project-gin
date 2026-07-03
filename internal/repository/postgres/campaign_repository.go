package postgres

import (
	"crm-project/internal/models/entity"
	"errors"

	"gorm.io/gorm"
)

type CampaignRepository struct {
	dbgorm *gorm.DB
}

func NewCampaignRepository(db *gorm.DB) *CampaignRepository {
	return &CampaignRepository{dbgorm: db}
}

func (r *CampaignRepository) CreateCampaign(campign *entity.Campaign) (*entity.Campaign, error) {
	err := r.dbgorm.Create(campign).Error
	if err != nil {
		return nil, err
	}
	return campign, nil
}

func (r *CampaignRepository) GetAllCampaigns(status string) ([]entity.Campaign, error) {
	var campaigns []entity.Campaign
	query := r.dbgorm.Preload("CreatedByUser").Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&campaigns).Error
	if err != nil {
		return nil, err
	}
	return campaigns, nil
}

func (r *CampaignRepository) UpdateCampaign(campaignID string, updates map[string]interface{}) error {
	return r.dbgorm.Model(&entity.Campaign{}).Where("id = ?", campaignID).Updates(updates).Error
}

func (r *CampaignRepository) DeleteCampaign(campaignID string) error {
	return r.dbgorm.Where("id = ?", campaignID).Delete(&entity.Campaign{}).Error
}

func (r *CampaignRepository) GetCampaignByID(campaignID string) (*entity.Campaign, error) {
	var campaign entity.Campaign
	err := r.dbgorm.
		Preload("CreatedByUser").
		Preload("Recipients").
		Where("id = ?", campaignID).
		First(&campaign).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("campaign not found")
	}
	if err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (r *CampaignRepository) GetScheduledCampaigns() ([]entity.Campaign, error) {
	var campaigns []entity.Campaign
	err := r.dbgorm.
		Where("status = ? AND scheduled_at <= NOW()", "SCHEDULED").
		Find(&campaigns).Error
	return campaigns, err
}
