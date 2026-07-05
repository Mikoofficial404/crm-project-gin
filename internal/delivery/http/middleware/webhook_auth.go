package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// WebhookAuthMiddleware memvalidasi request webhook WhatsApp menggunakan
// header X-Webhook-Token atau query param ?token= yang dicocokkan dengan WEBHOOK_SECRET di .env
func WebhookAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := os.Getenv("WEBHOOK_SECRET")

		// Kalau WEBHOOK_SECRET tidak di-set, lewati validasi (development mode)
		if secret == "" {
			c.Next()
			return
		}

		// Cek header dulu, fallback ke query param
		token := c.GetHeader("X-Webhook-Token")
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Webhook token wajib diisi (header X-Webhook-Token atau query ?token=)"})
			c.Abort()
			return
		}

		if token != secret {
			c.JSON(http.StatusForbidden, gin.H{"error": "Webhook token tidak valid"})
			c.Abort()
			return
		}

		c.Next()
	}
}
