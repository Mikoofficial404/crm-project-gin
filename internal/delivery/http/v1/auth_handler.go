package v1

import (
	"crm-project/internal/service"
	"crm-project/pkg/jwt"
	"crm-project/pkg/response"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
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

type ForgetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// Register godoc
// @Summary     Register user baru
// @Description Daftarkan user baru ke sistem
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body RegisterRequest true "Data registrasi"
// @Success     201 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	data, err := h.userService.Register(req.Email, req.Name, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success("Registrasi berhasil", data))
}

// Login godoc
// @Summary     Login user
// @Description Login dan dapatkan JWT token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body LoginRequest true "Data login"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	ip := c.ClientIP()
	reqUser := c.Request.UserAgent()

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	token, userID, requires2FA, err := h.userService.Login(req.Email, req.Password, ip, reqUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	if requires2FA {
		c.JSON(http.StatusAccepted, gin.H{
			"success": false,
			"message": "OTP Diperlukan! Silakan verifikasi.",
			"data":    gin.H{"user_id": userID},
		})
		return
	}

	c.JSON(http.StatusOK, response.Success("Login berhasil", gin.H{"token": token}))
}

// Logout godoc
// @Summary     Logout user
// @Description Logout dan blacklist token JWT
// @Tags        auth
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response
// @Failure     401 {object} response.Response
// @Router      /logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	token, err := jwt.GetBearerToken(c.Request.Header)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error("Unauthorized: Token tidak valid"))
		return
	}
	userID := c.MustGet("user_id").(string)
	if err := h.userService.Logout(userID, token); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal logout"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Logout berhasil", nil))
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword godoc
// @Summary     Ganti password
// @Description Ganti password user yang sedang login
// @Tags        auth
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       body body map[string]string true "old_password dan new_password"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /profile/password [patch]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	userID := c.MustGet("user_id").(string)
	if err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Password berhasil diubah", nil))
}

// ForgotPassword godoc
// @Summary     Lupa password
// @Description Kirim email reset password
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body ForgetPasswordRequest true "Email user"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /auth/forgot-password [post]
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req ForgetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	if err := h.userService.ForgotPassword(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Email reset password telah dikirim", nil))
}

// ResetPassword godoc
// @Summary     Reset password
// @Description Reset password dengan token dari email
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body ResetPasswordRequest true "Token dan password baru"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /auth/reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	if err := h.userService.ResetPassword(req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Reset password berhasil", nil))
}

// VerifyOTP godoc
// @Summary     Verifikasi OTP 2FA
// @Description Verifikasi kode OTP untuk 2FA
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body VerifyOTPRequest true "UserID dan OTP code"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /login/verify-otp [post]
func (h *UserHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	data, err := h.userService.VerifyOTP(req.UserID, req.OTPCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Verifikasi OTP berhasil", gin.H{"token": data}))
}

// Setup2FA godoc
// @Summary     Setup 2FA
// @Description Setup two-factor authentication
// @Tags        auth
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /profile/2fa/setup [get]
func (h *UserHandler) Setup2FA(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	data, err := h.userService.SetUp2FA(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Setup 2FA berhasil", data))
}

func (h *UserHandler) LoginGoogle(c *gin.Context) {
	url := h.userService.GetGoogleLoginURL()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *UserHandler) CallbackGoogle(c *gin.Context) {
	code := c.Query("code")
	token, err := h.userService.GoogleCallback(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success("Login Google berhasil", gin.H{"token": token}))
}

// GetMe godoc
// @Summary     Get profil user
// @Description Ambil data profil user yang sedang login
// @Tags        auth
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response
// @Failure     401 {object} response.Response
// @Router      /me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	user, err := h.userService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("User tidak ditemukan"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil profil", gin.H{
		"id":                    user.ID,
		"name":                  user.Name,
		"email":                 user.Email,
		"role":                  user.Role,
		"is_two_factor_enabled": user.IsTwoFactorEnabled,
	}))
}

// GetUsers godoc
// @Summary     Get semua users
// @Description Ambil daftar semua user (admin only)
// @Tags        admin
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} response.Response
// @Failure     403 {object} response.Response
// @Router      /admin/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal mengambil data users"))
		return
	}
	c.JSON(http.StatusOK, response.Success("Berhasil mengambil data users", users))
}

// GetLoginHistory godoc
// @Summary     Riwayat login
// @Description Ambil riwayat login user yang sedang login
// @Tags        auth
// @Security    BearerAuth
// @Produce     json
// @Param       limit query int false "Jumlah data (default 10)"
// @Success     200 {object} response.Response
// @Failure     500 {object} response.Response
// @Router      /me/login-history [get]
func (h *UserHandler) GetLoginHistory(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var limit int
	if limitStr := c.Query("limit"); limitStr != "" {
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || limit <= 0 {
			limit = 10
		}
	} else {
		limit = 10
	}

	histories, err := h.userService.GetLoginHistory(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Gagal mengambil riwayat login"))
		return
	}

	c.JSON(http.StatusOK, response.Success("Berhasil mengambil riwayat login", histories))
}
