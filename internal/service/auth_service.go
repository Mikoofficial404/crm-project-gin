package service

import (
	"context"
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"crm-project/pkg/jwt"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

var GoogleOAuthConfig = &oauth2.Config{
	RedirectURL: "http://localhost:8080/api/v1/auth/google/callback",
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	},
	Endpoint: google.Endpoint,
}

type UserService struct {
	user *postgres.UserRepository
}

func NewUserService(userRepo *postgres.UserRepository) *UserService {
	GoogleOAuthConfig.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleOAuthConfig.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
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

func (s *UserService) Login(email string, password string) (token string, userID string, requires2FA bool, err error) {
	isUser, err := s.user.FindByEmail(email)
	if err != nil {
		return "", "", false, fmt.Errorf("user not found")
	}

	jwtCheck, err := jwt.CheckPasswordHash(password, isUser.Password)
	if err != nil {
		return "", "", false, fmt.Errorf("password check failed")
	}
	if !jwtCheck {
		return "", "", false, fmt.Errorf("incorrect password")
	}

	if isUser.IsTwoFactorEnabled {
		return "", isUser.ID, true, nil
	}

	parseUUID, err := uuid.Parse(isUser.ID)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to parse UUID")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtMake, err := jwt.MakeJWT(parseUUID, isUser.Role, jwtSecret, time.Hour*24)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to create JWT token")
	}

	return jwtMake, isUser.ID, false, nil
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

func (s *UserService) VerifyOTP(userID string, otpCode string) (string, error) {
	user, err := s.user.FindByID(userID)
	if err != nil {
		return "", err
	}
	valid := totp.Validate(otpCode, user.TwoFactorSecret)
	if !valid {
		return "", fmt.Errorf("kode OTP salah atau kedaluwarsa")
	}
	parseUUID, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("failed to parse UUID")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtMake, err := jwt.MakeJWT(parseUUID, user.Role, jwtSecret, time.Hour*24)
	if err != nil {
		return "", fmt.Errorf("failed to create JWT token")
	}

	return jwtMake, nil
}

func (s *UserService) SetUp2FA(userID string) (string, error) {

	user, errUser := s.user.FindByID(userID)
	if errUser != nil {
		return "", errUser
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "CRM Project", AccountName: user.Email})
	if err != nil {
		return "", err
	}
	secretData := key.Secret()

	errUpdate := s.user.Update2FA(userID, secretData)
	if errUpdate != nil {
		return "", errUpdate
	}

	return key.URL(), nil
}

func (s *UserService) GetGoogleLoginURL() string {
	return GoogleOAuthConfig.AuthCodeURL("random-state-123")
}

func (s *UserService) GoogleCallback(code string) (string, error) {

	goggleTokens, err := GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return "", err
	}
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + goggleTokens.AccessToken)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	var req GoogleUserInfo
	if err := json.Unmarshal(data, &req); err != nil {
		return "", err
	}

	userDB, err := s.user.FindByEmail(req.Email)
	if err != nil {
		newUser := entity.User{
			Email:    req.Email,
			Name:     req.Name,
			Password: "google-oauth-random-password",
		}

		createdUser, errCreate := s.user.CreateUser(&newUser)
		if errCreate != nil {
			return "", fmt.Errorf("gagal mendaftarkan user baru: %v", errCreate)
		}
		userDB = createdUser
	}
	parseUUID, err := uuid.Parse(userDB.ID)
	if err != nil {
		return "", fmt.Errorf("failed to parse UUID")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtMake, err := jwt.MakeJWT(parseUUID, userDB.Role, jwtSecret, time.Hour*24)
	if err != nil {
		return "", fmt.Errorf("failed to create JWT token")
	}

	return jwtMake, nil

}
