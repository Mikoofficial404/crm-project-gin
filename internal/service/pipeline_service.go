package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"errors"
)

type PipelineService struct {
	pipelineRepo      *postgres.PipelineRepository
	pipelineStageRepo *postgres.PipelineStageRepository
}

func NewPipelineService(pipelineRepo *postgres.PipelineRepository, pipelineStageRepo *postgres.PipelineStageRepository) *PipelineService {
	return &PipelineService{
		pipelineRepo:      pipelineRepo,
		pipelineStageRepo: pipelineStageRepo,
	}
}

func (s *PipelineService) CreatePipeline(name, description, createdBy string, isDefault bool) (*entity.Pipeline, error) {
	if name == "" || createdBy == "" {
		return nil, errors.New("name dan created_by wajib diisi")
	}

	if isDefault {
		if err := s.pipelineRepo.UnsetAllDefaults(); err != nil {
			return nil, err
		}
	}

	pipeline := entity.Pipeline{
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		IsDefault:   isDefault,
		IsActive:    true,
	}
	return s.pipelineRepo.CreatePipeline(&pipeline)
}

func (s *PipelineService) GetAllPipelines() ([]entity.Pipeline, error) {
	return s.pipelineRepo.GetAllPipelines()
}

func (s *PipelineService) GetPipelineByID(pipelineID string) (*entity.Pipeline, error) {
	if pipelineID == "" {
		return nil, errors.New("pipeline_id wajib diisi")
	}
	return s.pipelineRepo.GetPipelineByID(pipelineID)
}

func (s *PipelineService) GetDefaultPipeline() (*entity.Pipeline, error) {
	return s.pipelineRepo.GetDefaultPipeline()
}

func (s *PipelineService) UpdatePipeline(pipelineID string, updates map[string]interface{}, setDefault bool) error {
	if pipelineID == "" {
		return errors.New("pipeline_id wajib diisi")
	}

	_, err := s.pipelineRepo.GetPipelineByID(pipelineID)
	if err != nil {
		return errors.New("pipeline tidak ditemukan")
	}

	if setDefault {
		if err := s.pipelineRepo.UnsetAllDefaults(); err != nil {
			return err
		}
		updates["is_default"] = true
	}

	return s.pipelineRepo.UpdatePipeline(pipelineID, updates)
}

func (s *PipelineService) DeletePipeline(pipelineID string) error {
	if pipelineID == "" {
		return errors.New("pipeline_id wajib diisi")
	}

	_, err := s.pipelineRepo.GetPipelineByID(pipelineID)
	if err != nil {
		return errors.New("pipeline tidak ditemukan")
	}

	count, err := s.pipelineRepo.CountDealsInPipeline(pipelineID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("pipeline tidak bisa dihapus karena masih memiliki deal aktif")
	}

	return s.pipelineRepo.DeletePipeline(pipelineID)
}

func (s *PipelineService) AddStage(pipelineID, name, color string, isClosedWon, isClosedLost bool, probability int, slaHours int) (*entity.PipelineStage, error) {
	if pipelineID == "" || name == "" {
		return nil, errors.New("pipeline_id dan name wajib diisi")
	}

	if probability < 0 || probability > 100 {
		return nil, errors.New("probability harus antara 0 dan 100")
	}

	_, err := s.pipelineRepo.GetPipelineByID(pipelineID)
	if err != nil {
		return nil, errors.New("pipeline tidak ditemukan")
	}

	existingStages, err := s.pipelineStageRepo.GetStagesByPipelineID(pipelineID)
	if err != nil {
		return nil, err
	}
	nextOrder := len(existingStages)

	stage := entity.PipelineStage{
		PipelineID:   pipelineID,
		Name:         name,
		Color:        color,
		StageOrder:   nextOrder,
		IsClosedWon:  isClosedWon,
		IsClosedLost: isClosedLost,
		Probability:  probability,
	}
	return s.pipelineStageRepo.CreatePipelineStage(&stage)
}

func (s *PipelineService) GetStagesByPipeline(pipelineID string) ([]entity.PipelineStage, error) {
	if pipelineID == "" {
		return nil, errors.New("pipeline_id wajib diisi")
	}
	return s.pipelineStageRepo.GetStagesByPipelineID(pipelineID)
}

func (s *PipelineService) UpdateStage(stageID string, updates map[string]interface{}) error {
	if stageID == "" {
		return errors.New("stage_id wajib diisi")
	}

	_, err := s.pipelineStageRepo.GetStageByID(stageID)
	if err != nil {
		return errors.New("stage tidak ditemukan")
	}
	slaHours, ok := updates["sla_hours"].(int)
	if ok {
		updates["sla_hours"] = slaHours
	}

	return s.pipelineStageRepo.UpdatePipelineStage(stageID, updates)
}

func (s *PipelineService) DeleteStage(stageID string) error {
	if stageID == "" {
		return errors.New("stage_id wajib diisi")
	}

	_, err := s.pipelineStageRepo.GetStageByID(stageID)
	if err != nil {
		return errors.New("stage tidak ditemukan")
	}

	return s.pipelineStageRepo.DeletePipelineStage(stageID)
}

func (s *PipelineService) ReorderStages(pipelineID string, stageOrders []entity.StageOrder) error {
	if pipelineID == "" {
		return errors.New("pipeline_id wajib diisi")
	}
	if len(stageOrders) == 0 {
		return errors.New("stage_orders tidak boleh kosong")
	}

	_, err := s.pipelineRepo.GetPipelineByID(pipelineID)
	if err != nil {
		return errors.New("pipeline tidak ditemukan")
	}

	return s.pipelineStageRepo.ReorderStages(pipelineID, stageOrders)
}
