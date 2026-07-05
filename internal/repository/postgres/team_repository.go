package postgres

import (
	"crm-project/internal/models/entity"

	"gorm.io/gorm"
)

type TeamRepository struct {
	dbGorm *gorm.DB
}

func NewTeamRepository(dbGorm *gorm.DB) *TeamRepository {
	return &TeamRepository{dbGorm: dbGorm}
}

func (r *TeamRepository) CreateTeam(team *entity.Team) error {
	newTeam := r.dbGorm.Create(team)
	err := newTeam.Error
	if err != nil {
		return err
	}
	return nil
}

func (r *TeamRepository) GetTeamByID(teamID string) (*entity.Team, error) {
	var team entity.Team
	err := r.dbGorm.Preload("Members").Preload("Manager").Where("id = ?", teamID).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) GetAllTeams() ([]entity.Team, error) {
	var teams []entity.Team
	err := r.dbGorm.Preload("Members").Preload("Manager").Find(&teams).Error
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *TeamRepository) UpdateTeam(teamID string, updates map[string]interface{}) error {
	return r.dbGorm.Model(&entity.Team{}).Where("id = ?", teamID).Updates(updates).Error
}

func (r *TeamRepository) DeleteTeam(teamID string) error {
	return r.dbGorm.Where("id = ?", teamID).Delete(&entity.Team{}).Error
}

func (r *TeamRepository) AddMember(teamID, userID string) error {
	return r.dbGorm.Model(&entity.User{}).Where("id = ?", userID).Update("team_id", teamID).Error
}

func (r *TeamRepository) RemoveMember(teamID, userID string) error {
	return r.dbGorm.Model(&entity.User{}).
		Where("id = ? AND team_id = ?", userID, teamID).
		Update("team_id", nil).Error
}

func (r *TeamRepository) CountMembers(teamID string) (int64, error) {
	var count int64
	err := r.dbGorm.Model(&entity.User{}).Where("team_id = ?", teamID).Count(&count).Error
	return count, err
}
