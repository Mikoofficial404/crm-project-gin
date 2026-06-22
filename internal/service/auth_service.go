package service

import (
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"crm-project/pkg/jwt"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	user *postgres.UserRepository
}

func NewUserService(userRepo *postgres.UserRepository) *UserService {
	return &UserService{
		user: userRepo,
	}
}

func (s *UserService) Register(email string, name string, password string) (*entity.User, error) {
	HashPass, err := jwt.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := entity.User{
		Email:    email,
		Name:     name,
		Password: HashPass,
	}
	result, err := s.user.CreateUser(&user)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *UserService) Login(email string, password string) (string, error) {
	isUser, err := s.user.FindByEmail(email)
	if err != nil {
		return "", fmt.Errorf("email not registered")
	}

	jwtCheck, err := jwt.CheckPasswordHash(password, isUser.Password)
	if err != nil {
		return "", fmt.Errorf("password check failed")
	}
	if !jwtCheck {
		return "", fmt.Errorf("incorrect password")
	}

	parseUUID, err := uuid.Parse(isUser.ID)
	if err != nil {
		return "", fmt.Errorf("failed to parse UUID")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtMake, err := jwt.MakeJWT(parseUUID, isUser.Role, jwtSecret, time.Hour*24)
	if err != nil {
		return "", fmt.Errorf("failed to create JWT token")
	}
	return jwtMake, nil
}

func (s *UserService) ChangePassword(userID string, oldPassword string, newPassword string) error {
	user, err := s.user.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user tidak ditemukan")
	}

	match, err := jwt.CheckPasswordHash(oldPassword, user.Password)
	if err != nil || !match {
		return fmt.Errorf("password lama salah")
	}

	newHash, err := jwt.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("gagal mengenkripsi password baru")
	}

	return s.user.UpdatePassword(userID, newHash)
}
