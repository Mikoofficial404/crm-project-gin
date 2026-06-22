package service

import (
	"context"
	"crm-project/internal/repository/postgres"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type DashboardService struct {
	repo        *postgres.DashboardRepository
	redisClient *redis.Client
}

func NewDashboardService(repo *postgres.DashboardRepository, redisClient *redis.Client) *DashboardService {
	return &DashboardService{
		repo:        repo,
		redisClient: redisClient,
	}
}

func (s *DashboardService) GetStats() (*postgres.DashboardStats, error) {
	ctx := context.Background()
	cacheKey := "crm_dashboard_stats"

	data, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var stats postgres.DashboardStats
		if unmarshalErr := json.Unmarshal([]byte(data), &stats); unmarshalErr == nil {
			return &stats, nil
		}
	}

	stats, err := s.repo.GetStats()
	if err != nil {
		return nil, err
	}

	dataJson, _ := json.Marshal(stats)
	s.redisClient.Set(ctx, cacheKey, dataJson, 5*time.Minute)

	return stats, nil
}
