package service

import (
	"context"
	"crm-project/internal/repository/postgres"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ReportService struct {
	reportRepo  *postgres.ReportRepository
	redisClient *redis.Client
}

func NewReportService(reportRepo *postgres.ReportRepository, redisClient *redis.Client) *ReportService {
	return &ReportService{
		reportRepo:  reportRepo,
		redisClient: redisClient,
	}
}

func parseDateRange(startStr, endStr string) (time.Time, time.Time, error) {
	var startDate, endDate time.Time
	var err error

	if startStr == "" {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	} else {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("format start_date tidak valid, gunakan YYYY-MM-DD")
		}
	}

	if endStr == "" {

		now := time.Now()
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)
	} else {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("format end_date tidak valid, gunakan YYYY-MM-DD")
		}

		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.Local)
	}

	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date tidak boleh sebelum start_date")
	}

	return startDate, endDate, nil
}

func (s *ReportService) GetSalesSummary(startStr, endStr string) (*postgres.SalesSummaryResult, error) {
	startDate, endDate, err := parseDateRange(startStr, endStr)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("report:sales_summary:%s:%s", startDate.Format("20060102"), endDate.Format("20060102"))

	cached, err := s.redisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var result postgres.SalesSummaryResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			return &result, nil
		}
	}

	result, err := s.reportRepo.GetSalesSummary(startDate, endDate)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(result); err == nil {
		s.redisClient.Set(context.Background(), cacheKey, data, 5*time.Minute)
	}

	return result, nil
}

func (s *ReportService) GetPipelineReport(startStr, endStr, pipelineID string) ([]postgres.PipelineStageReport, error) {
	startDate, endDate, err := parseDateRange(startStr, endStr)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("report:pipeline:%s:%s:%s", startDate.Format("20060102"), endDate.Format("20060102"), pipelineID)

	cached, err := s.redisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var result []postgres.PipelineStageReport
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result, nil
		}
	}

	result, err := s.reportRepo.GetPipelineReport(startDate, endDate, pipelineID)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(result); err == nil {
		s.redisClient.Set(context.Background(), cacheKey, data, 5*time.Minute)
	}

	return result, nil
}

func (s *ReportService) GetSalesPerformance(startStr, endStr string) ([]postgres.SalesPerformanceResult, error) {
	startDate, endDate, err := parseDateRange(startStr, endStr)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("report:sales_performance:%s:%s", startDate.Format("20060102"), endDate.Format("20060102"))

	cached, err := s.redisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var result []postgres.SalesPerformanceResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result, nil
		}
	}

	result, err := s.reportRepo.GetSalesPerformance(startDate, endDate)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(result); err == nil {
		s.redisClient.Set(context.Background(), cacheKey, data, 5*time.Minute)
	}

	return result, nil
}

func (s *ReportService) GetLeadSourceReport(startStr, endStr string) ([]postgres.LeadSourceResult, error) {
	startDate, endDate, err := parseDateRange(startStr, endStr)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("report:lead_source:%s:%s", startDate.Format("20060102"), endDate.Format("20060102"))

	cached, err := s.redisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var result []postgres.LeadSourceResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result, nil
		}
	}

	result, err := s.reportRepo.GetLeadSourceReport(startDate, endDate)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(result); err == nil {
		s.redisClient.Set(context.Background(), cacheKey, data, 5*time.Minute)
	}

	return result, nil
}

func (s *ReportService) GetActivityReport(startStr, endStr string) ([]postgres.ActivityReportResult, error) {
	startDate, endDate, err := parseDateRange(startStr, endStr)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("report:activity:%s:%s", startDate.Format("20060102"), endDate.Format("20060102"))

	cached, err := s.redisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var result []postgres.ActivityReportResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result, nil
		}
	}

	result, err := s.reportRepo.GetActivityReport(startDate, endDate)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(result); err == nil {
		s.redisClient.Set(context.Background(), cacheKey, data, 5*time.Minute)
	}

	return result, nil
}
