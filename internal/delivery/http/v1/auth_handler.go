package v1

import (
	"context"
	"crm-project/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type UserHandler struct {
	userService *service.UserService
	redisClient *redis.Client
}

func NewUserHandler(userService *service.UserService, redisClient *redis.Client) *UserHandler {
	return &UserHandler{
		userService: userService,
		redisClient: redisClient,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type VerifyOTPRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	OTPCode string `json:"otp_code" binding:"required"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.userService.Register(req.Email, req.Name, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "registration successful", "data": data})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, userID, requires2FA, err := h.userService.Login(req.Email, req.Password)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requires2FA {
		c.JSON(http.StatusAccepted, gin.H{
			"message": "OTP Diperlukan! Silakan verifikasi.",
			"user_id": userID,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "login sukses", "token": token})
}

func (h *UserHandler) Logout(c *gin.Context) {
	token := c.MustGet("token").(string)
	ctx := context.Background()

	err := h.redisClient.Set(ctx, "blacklist:"+token, "true", 24*time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Logout berhasil"})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(string)

	err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Password berhasil diubah"})
}

func (h *UserHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := h.userService.VerifyOTP(req.UserID, req.OTPCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "Verifikasi OTP Sukses!",
		"token":  data,
	})
}

func (h *UserHandler) Setup2FA(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	data, err := h.userService.SetUp2FA(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "SetUp2FA", "data": data})
}

func (h *UserHandler) LoginGoogle(c *gin.Context) {
	url := h.userService.GetGoogleLoginURL()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *UserHandler) CallbackGoogle(c *gin.Context) {
	code := c.Query("code")
	token, err := h.userService.GoogleCallback(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Token", "token": token})

}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	user, err := h.userService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":                    user.ID,
		"name":                  user.Name,
		"email":                 user.Email,
		"role":                  user.Role,
		"is_two_factor_enabled": user.IsTwoFactorEnabled,
	})
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil data users",
		"data":    users,
	})
}
