package middleware

import (
	"context"
	"crm-project/pkg/jwt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func AuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtBearer, err := jwt.GetBearerToken(c.Request.Header)
		if err != nil {
			jwtBearer = c.Query("token")
			if jwtBearer == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Token tidak valid"})
				return
			}
		}

		ctx := context.Background()
		_, redisErr := rdb.Get(ctx, "blacklist:"+jwtBearer).Result()
		if redisErr == nil {

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Token sudah tidak berlaku (logout)"})
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
		c.Set("token", jwtBearer)
		c.Next()
	}
}
