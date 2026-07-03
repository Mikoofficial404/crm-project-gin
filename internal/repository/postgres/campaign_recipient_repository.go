package postgres

import (
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type CampaignRecipientRepository struct {
	dbgorm *gorm.DB
}

func NewCampaignRecipientRepository(db *gorm.DB) *CampaignRecipientRepository {
	return &CampaignRecipientRepository{dbgorm: db}
}

func (r *CampaignRecipientRepository) BulkInsertRecipients(recipients []entity.CampaignRecipient) error {
	return r.dbgorm.CreateInBatches(recipients, 100).Error
}

func (r *CampaignRecipientRepository) UpdateRecipientStatus(recipientID string, status string, errMsg *string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == "SENT" {
		updates["sent_at"] = gorm.Expr("NOW()")
	}
	if errMsg != nil {
		updates["error_msg"] = errMsg
	}
	return r.dbgorm.Model(&entity.CampaignRecipient{}).
		Where("id = ?", recipientID).
		Updates(updates).Error
}

func (r *CampaignRecipientRepository) GetPendingRecipients(campaignID string) ([]entity.CampaignRecipient, error) {
	var recipients []entity.CampaignRecipient
	err := r.dbgorm.
		Where("campaign_id = ? AND status = ?", campaignID, "PENDING").
		Find(&recipients).Error
	return recipients, err
}

func (r *CampaignRecipientRepository) GetRecipientsByCampaign(campaignID string, page, limit int) ([]entity.CampaignRecipient, int64, error) {
	var recipients []entity.CampaignRecipient
	var total int64

	baseQuery := r.dbgorm.Model(&entity.CampaignRecipient{}).Where("campaign_id = ?", campaignID)
	baseQuery.Count(&total)

	err := baseQuery.
		Preload("Lead").
		Preload("Contact").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&recipients).Error
	return recipients, total, err
}

func (r *CampaignRecipientRepository) CountPendingRecipients(campaignID string) (int64, error) {
	var count int64
	err := r.dbgorm.Model(&entity.CampaignRecipient{}).
		Where("campaign_id = ? AND status = ?", campaignID, "PENDING").
		Count(&count).Error
	return count, err
}

type CampaignStats struct {
	Total   int64 `json:"total"`
	Sent    int64 `json:"sent"`
	Failed  int64 `json:"failed"`
	Pending int64 `json:"pending"`
	Bounced int64 `json:"bounced"`
}

func (r *CampaignRecipientRepository) GetCampaignStats(campaignID string) (*CampaignStats, error) {
	var stats CampaignStats

	err := r.dbgorm.Model(&entity.CampaignRecipient{}).
		Where("campaign_id = ?", campaignID).
		Count(&stats.Total).Error
	if err != nil {
		return nil, err
	}

	counts := []struct {
		Status string
		Count  int64
	}{}

	err = r.dbgorm.Model(&entity.CampaignRecipient{}).
		Select("status, COUNT(*) as count").
		Where("campaign_id = ?", campaignID).
		Group("status").
		Scan(&counts).Error
	if err != nil {
		return nil, err
	}

	for _, c := range counts {
		switch c.Status {
		case "SENT":
			stats.Sent = c.Count
		case "FAILED":
			stats.Failed = c.Count
		case "PENDING":
			stats.Pending = c.Count
		case "BOUNCED":
			stats.Bounced = c.Count
		}
	}

	return &stats, nil
}
