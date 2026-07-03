package postgres

import (
	"crm-project/internal/models/entity"
	"errors"

	"gorm.io/gorm"
)

type PipelineRepository struct {
	dbgorm *gorm.DB
}

func NewPipelineRepository(db *gorm.DB) *PipelineRepository {
	return &PipelineRepository{dbgorm: db}
}

func (r *PipelineRepository) CreatePipeline(pipeline *entity.Pipeline) (*entity.Pipeline, error) {
	newPipeLine := r.dbgorm.Create(pipeline)
	err := newPipeLine.Error
	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

func (r *PipelineRepository) GetPipelineByID(pipelineID string) (*entity.Pipeline, error) {
	var pipeline entity.Pipeline
	err := r.dbgorm.
		Preload("Stages", func(db *gorm.DB) *gorm.DB {
			return db.Order("stage_order ASC").Where("deleted_at IS NULL")
		}).
		Preload("CreatedByUser").
		Where("id = ?", pipelineID).
		First(&pipeline).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("pipeline tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}
	return &pipeline, nil
}

func (r *PipelineRepository) GetAllPipelines() ([]entity.Pipeline, error) {
	var pipelines []entity.Pipeline
	err := r.dbgorm.
		Preload("Stages", func(db *gorm.DB) *gorm.DB {
			return db.Order("stage_order ASC").Where("deleted_at IS NULL")
		}).
		Order("created_at DESC").
		Find(&pipelines).Error

	if err != nil {
		return nil, err
	}
	return pipelines, nil
}

func (r *PipelineRepository) GetDefaultPipeline() (*entity.Pipeline, error) {
	var pipeline entity.Pipeline
	err := r.dbgorm.
		Preload("Stages", func(db *gorm.DB) *gorm.DB {
			return db.Order("stage_order ASC").Where("deleted_at IS NULL")
		}).
		Where("is_default = ? AND is_active = ?", true, true).
		First(&pipeline).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("default pipeline tidak ditemukan")
	}
	if err != nil {
		return nil, err
	}
	return &pipeline, nil
}

func (r *PipelineRepository) UnsetAllDefaults() error {
	return r.dbgorm.Model(&entity.Pipeline{}).Where("is_default = ?", true).Update("is_default", false).Error
}

func (r *PipelineRepository) CountDealsInPipeline(pipelineID string) (int64, error) {
	var count int64
	err := r.dbgorm.Model(&entity.Deal{}).Where("pipeline_id = ?", pipelineID).Count(&count).Error
	return count, err
}

func (r *PipelineRepository) UpdatePipeline(pipeLineID string, updates map[string]interface{}) error {
	return r.dbgorm.Model(&entity.Pipeline{}).Where("id = ?", pipeLineID).Updates(updates).Error
}

func (r *PipelineRepository) DeletePipeline(pipeLineID string) error {
	return r.dbgorm.Where("id = ?", pipeLineID).Delete(&entity.Pipeline{}).Error
}
