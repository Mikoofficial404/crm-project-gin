package postgres

import (
	"crm-project/internal/models/entity"
	"errors"

	"gorm.io/gorm"
)

type PipelineStageRepository struct {
	dbgorm *gorm.DB
}

func NewPipelineStageRepository(db *gorm.DB) *PipelineStageRepository {
	return &PipelineStageRepository{dbgorm: db}
}

func (r *PipelineStageRepository) CreatePipelineStage(pipelineStage *entity.PipelineStage) (*entity.PipelineStage, error) {
	err := r.dbgorm.Create(pipelineStage).Error
	if err != nil {
		return nil, err
	}
	return pipelineStage, nil
}

func (r *PipelineStageRepository) GetStageByID(stageID string) (*entity.PipelineStage, error) {
	var stage entity.PipelineStage
	err := r.dbgorm.Where("id = ?", stageID).First(&stage).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("stage not found")
	}
	if err != nil {
		return nil, err
	}
	return &stage, nil
}

func (r *PipelineStageRepository) GetStagesByPipelineID(pipelineID string) ([]entity.PipelineStage, error) {
	var stages []entity.PipelineStage
	err := r.dbgorm.
		Where("pipeline_id = ?", pipelineID).
		Order("stage_order ASC").
		Find(&stages).Error
	if err != nil {
		return nil, err
	}
	return stages, nil
}

func (r *PipelineStageRepository) UpdatePipelineStage(stageID string, updates map[string]interface{}) error {
	return r.dbgorm.Model(&entity.PipelineStage{}).Where("id = ?", stageID).Updates(updates).Error
}

func (r *PipelineStageRepository) DeletePipelineStage(stageID string) error {
	return r.dbgorm.Where("id = ?", stageID).Delete(&entity.PipelineStage{}).Error
}

func (r *PipelineStageRepository) ReorderStages(pipelineID string, stageOrders []entity.StageOrder) error {
	tx := r.dbgorm.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, item := range stageOrders {
		err := tx.Model(&entity.PipelineStage{}).
			Where("id = ? AND pipeline_id = ?", item.StageID, pipelineID).
			Update("stage_order", item.Order).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
