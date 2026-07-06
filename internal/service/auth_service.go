package service

import (
	"context"
	"crm-project/internal/models/entity"
	"crm-project/internal/repository/postgres"
	"crm-project/internal/worker"
	"crm-project/pkg/jwt"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
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
	user         *postgres.UserRepository
	rdb          *redis.Client
	asynq        *asynq.Client
	loginHistory *postgres.LoginHistoryRepository
}

func NewUserService(userRepo *postgres.UserRepository, rdb *redis.Client, clientAsynq *asynq.Client, loginHistoryRepo *postgres.LoginHistoryRepository) *UserService {
	GoogleOAuthConfig.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleOAuthConfig.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	return &UserService{
		user:         userRepo,
		rdb:          rdb,
		asynq:        clientAsynq,
		loginHistory: loginHistoryRepo,
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
		Role:     "sales",
	}
	result, err := s.user.CreateUser(&user)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *UserService) Login(email string, password string, ip string, reqUser string) (token string, userID string, requires2FA bool, err error) {
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

	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return "", "", false, fmt.Errorf("failed to generate refresh token")
	}
	refreshToken := hex.EncodeToString(rawToken)

	hashBytes := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	expiry := time.Duration(time.Hour * 24 * 7)
	ctx := context.Background()
	if err := s.rdb.Set(ctx, "refresh:"+tokenHash, isUser.ID, expiry).Err(); err != nil {
		return "", "", false, fmt.Errorf("failed to save refresh token")
	}

	history := &entity.LoginHistory{
		UserID:    isUser.ID,
		IPAddress: ip,
		UserAgent: reqUser,
		Success:   true,
	}

	if err := s.loginHistory.CreateLoginHistory(history); err != nil {
		return "", "", false, fmt.Errorf("failed to save login history")
	}

	return jwtMake, refreshToken, false, nil
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

func (s *UserService) ForgotPassword(email string) error {
	ctx := context.Background()
	user, err := s.user.FindByEmail(email)
	if err != nil {
		return fmt.Errorf("user tidak ditemukan")
	}
	token, err := jwt.MakeRefreshToken()
	if err != nil {
		return fmt.Errorf("Token gagal")
	}
	err = s.rdb.Set(ctx, "reset:"+token, user.ID, 30*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("gagal menyimpan token reset")
	}

	subject := "Reset Password CRM"
	body := fmt.Sprintf(`
		<h2>Reset Password</h2>
		<p>Halo %s,</p>
		<p>Kami menerima permintaan reset password untuk akun Anda.</p>
		<p>Gunakan token berikut untuk mereset password Anda (berlaku 30 menit):</p>
		<p><strong>%s</strong></p>
		<p>Jika Anda tidak meminta reset password, abaikan email ini.</p>
	`, user.Name, token)

	task, err := worker.NewEmailDeliveryTask(user.Email, subject, body)
	if err != nil {
		return fmt.Errorf("gagal membuat email task")
	}
	_, err = s.asynq.Enqueue(task)
	if err != nil {
		return fmt.Errorf("gagal mengirim email")
	}

	return nil
}

func (s *UserService) ResetPassword(token string, newPassword string) error {
	ctx := context.Background()
	tokens, err := s.rdb.Get(ctx, "reset:"+token).Result()
	if err != nil {
		return fmt.Errorf("token invalid or expired")
	}
	hashPassword, err := jwt.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("Hash password gagal")
	}
	err = s.user.UpdatePassword(tokens, hashPassword)
	if err != nil {
		return fmt.Errorf("Update Password Gagal")
	}
	delTokens := s.rdb.Del(ctx, "reset:"+token)
	return delTokens.Err()
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

func (s *UserService) GetProfile(userID string) (*entity.User, error) {
	return s.user.FindByID(userID)
}

func (s *UserService) GetAllUsers() ([]entity.User, error) {
	return s.user.GetAllUsers()
}

func (s *UserService) Logout(userID string, token string) error {
	claims, err := jwt.ExtractClaims(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}

	hash := sha256.Sum256([]byte(token))
	redisKey := "blacklist:" + hex.EncodeToString(hash[:])
	return s.rdb.Set(context.Background(), redisKey, "1", ttl).Err()
}

func (s *UserService) GetLoginHistory(userID string, limit int) ([]entity.LoginHistory, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.loginHistory.GetLoginHistoriesByUserID(userID, limit)
}
