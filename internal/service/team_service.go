package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"errors"
)

type TeamService struct {
	teamRepo *postgres.TeamRepository
	userRepo *postgres.UserRepository
}

func NewTeamService(teamRepo *postgres.TeamRepository, userRepo *postgres.UserRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo, userRepo: userRepo}
}

func (s *TeamService) CreateTeam(name, description, managerID string) (*entity.Team, error) {
	if name == "" || managerID == "" {
		return nil, errors.New("name dan manager_id wajib diisi")
	}

	_, err := s.userRepo.FindByID(managerID)
	if err != nil {
		return nil, errors.New("manager tidak ditemukan")
	}

	team := &entity.Team{
		Name:        name,
		Description: description,
		ManagerID:   managerID,
	}

	err = s.teamRepo.CreateTeam(team)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *TeamService) GetAllTeams() ([]entity.Team, error) {
	return s.teamRepo.GetAllTeams()
}

func (s *TeamService) GetTeamByID(teamID string) (*entity.Team, error) {
	if teamID == "" {
		return nil, errors.New("team_id wajib diisi")
	}
	team, err := s.teamRepo.GetTeamByID(teamID)
	if err != nil {
		return nil, errors.New("team tidak ditemukan")
	}
	return team, nil
}

func (s *TeamService) UpdateTeam(teamID string, name, description *string, managerID *string) error {
	if teamID == "" {
		return errors.New("team_id wajib diisi")
	}

	_, err := s.teamRepo.GetTeamByID(teamID)
	if err != nil {
		return errors.New("team tidak ditemukan")
	}

	updates := map[string]interface{}{}
	if name != nil {
		if *name == "" {
			return errors.New("name tidak boleh kosong")
		}
		updates["name"] = *name
	}
	if description != nil {
		updates["description"] = *description
	}
	if managerID != nil {
		_, err := s.userRepo.FindByID(*managerID)
		if err != nil {
			return errors.New("manager baru tidak ditemukan")
		}
		updates["manager_id"] = *managerID
	}

	if len(updates) == 0 {
		return errors.New("tidak ada data yang diupdate")
	}

	return s.teamRepo.UpdateTeam(teamID, updates)
}

func (s *TeamService) DeleteTeam(teamID string) error {
	if teamID == "" {
		return errors.New("team_id wajib diisi")
	}

	_, err := s.teamRepo.GetTeamByID(teamID)
	if err != nil {
		return errors.New("team tidak ditemukan")
	}

	count, err := s.teamRepo.CountMembers(teamID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("team tidak bisa dihapus karena masih memiliki anggota")
	}

	return s.teamRepo.DeleteTeam(teamID)
}

func (s *TeamService) AddMember(teamID, userID string) error {
	if teamID == "" || userID == "" {
		return errors.New("team_id dan user_id wajib diisi")
	}

	_, err := s.teamRepo.GetTeamByID(teamID)
	if err != nil {
		return errors.New("team tidak ditemukan")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	if user.TeamID != nil && *user.TeamID != "" && *user.TeamID != teamID {
		return errors.New("user sudah terdaftar di team lain")
	}

	return s.teamRepo.AddMember(teamID, userID)
}

func (s *TeamService) RemoveMember(teamID, userID string) error {
	if teamID == "" || userID == "" {
		return errors.New("team_id dan user_id wajib diisi")
	}

	_, err := s.teamRepo.GetTeamByID(teamID)
	if err != nil {
		return errors.New("team tidak ditemukan")
	}

	return s.teamRepo.RemoveMember(teamID, userID)
}
