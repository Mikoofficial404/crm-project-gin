package service

import (
	"context"
	"crm-project/internal/models/entity"
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

func (s *DashboardService) GetForecastingAnalytics() (*entity.AnalyticsResponse, error) {
	addMoney, err := s.repo.GetTotalRevenue()
	if err != nil {
		return nil, err
	}
	predictioMoney, err := s.repo.GetProjectedRevenue()
	if err != nil {
		return nil, err
	}
	totalDealWon, err := s.repo.GetTotalWonDeals()
	if err != nil {
		return nil, err
	}
	totalLeads, err := s.repo.GetTotalLeads()
	if err != nil {
		return nil, err
	}

	var rasio float64 = 0
	if totalLeads > 0 {
		rasio = float64(totalDealWon) / float64(totalLeads) * 100
	}
	analyticts := entity.AnalyticsResponse{
		TotalRevenue:     addMoney,
		ProjectedRevenue: predictioMoney,
		ConversionRate:   rasio,
	}
	return &analyticts, err
}
