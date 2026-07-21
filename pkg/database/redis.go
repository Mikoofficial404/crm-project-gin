package database

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func ConnectRedis() (*redis.Client, error) {
	err := godotenv.Load()
	if err != nil {
		logrus.Info("⚠️  Warning: .env file not found, using system env")
	}

	redisPort := os.Getenv("REDIS_PORT")
	redisHost := os.Getenv("REDIS_HOST")

	redisconst := fmt.Sprintf("%s:%s", redisHost, redisPort)

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisconst,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("gagal konek ke Redis: %v", err)
	}
	logrus.Info("Berhasil konek ke Redis!")
	return rdb, nil
}
