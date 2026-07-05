package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	RateLimit  = 100
	RateWindow = time.Minute
)

func RateLimitUser(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}
		key := fmt.Sprintf("rate:user:%v", userId)
		ctx := context.Background()
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "rate limiter unavailable",
			})
			return
		}
		if count == 1 {
			redisClient.Expire(ctx, key, RateWindow)
		}
		if count > RateLimit {
			ttl, _ := redisClient.TTL(ctx, key).Result()
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", RateLimit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", fmt.Sprintf("%.0f", ttl.Seconds()))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests",
				"retry_after": fmt.Sprintf("%.0f seconds", ttl.Seconds()),
			})
			return
		}
		remaining := RateLimit - count
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", RateLimit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}
