package postgres

import (
	"crm-project/internal/models/entity"
	"errors"

	"gorm.io/gorm"
)

type DealCommentRepository struct {
	gormDB *gorm.DB
}

func NewDealCommentRepository(gormDB *gorm.DB) *DealCommentRepository {
	return &DealCommentRepository{gormDB: gormDB}
}

func (r *DealCommentRepository) CreateComment(dealID, userID, content string, mentionedUserIDs string) (*entity.DealComment, error) {
	if dealID == "" || userID == "" || content == "" {
		return nil, errors.New("dealID, userID, dan content wajib diisi")
	}
	comment := entity.DealComment{
		DealID:           dealID,
		UserID:           userID,
		Content:          content,
		MentionedUserIDs: mentionedUserIDs,
	}
	if err := r.gormDB.Create(&comment).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *DealCommentRepository) GetCommentsByDealID(dealID string, page, limit int) ([]entity.DealComment, int64, error) {
	var comments []entity.DealComment
	var total int64

	base := r.gormDB.Model(&entity.DealComment{}).Where("deal_id = ?", dealID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := base.
		Preload("User").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&comments).Error
	return comments, total, err
}

func (r *DealCommentRepository) DeleteComment(commentID string) error {
	if commentID == "" {
		return errors.New("commentID wajib diisi")
	}
	return r.gormDB.Where("id = ?", commentID).Delete(&entity.DealComment{}).Error
}
