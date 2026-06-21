package middleware

import (
	"crm-project/pkg/jwt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtBearer, err := jwt.GetBearerToken(c.Request.Header)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Token tidak valid"})
			return
		}

		secret := os.Getenv("JWT_SECRET")
		validateJwt, role, err := jwt.ValidateJWT(jwtBearer, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Token tidak valid"})
			return
		}
		c.Set("user_id", validateJwt.String())
		c.Set("role", role)
		c.Next()
	}
}
