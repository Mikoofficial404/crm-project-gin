package v1

import (
	"context"
	"net/http"
	"time"

	"crm-project/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type CronHandler struct {
	redisClient *redis.Client
}

func NewCronHandler(redisClient *redis.Client) *CronHandler {
	return &CronHandler{redisClient: redisClient}
}

type CronJobStatus struct {
	Name        string `json:"name"`
	Schedule    string `json:"schedule"`
	Description string `json:"description"`
	LastRun     string `json:"last_run"`
}

func (h *CronHandler) GetCronStatus(c *gin.Context) {
	ctx := context.Background()

	cronJobs := []struct {
		Name        string
		Schedule    string
		Description string
		RedisKey    string
	}{
		{
			Name:        "check_stale_leads",
			Schedule:    "* * * * *",
			Description: "Cek leads yang tidak aktif lebih dari 72 jam",
			RedisKey:    "cron:last_run:check_stale_leads",
		},
		{
			Name:        "send_due_date_reminders",
			Schedule:    "0 8 * * *",
			Description: "Kirim reminder task yang jatuh tempo besok",
			RedisKey:    "cron:last_run:send_due_date_reminders",
		},
		{
			Name:        "process_scheduled_campaigns",
			Schedule:    "0 8 * * *",
			Description: "Proses email campaign yang sudah terjadwal",
			RedisKey:    "cron:last_run:process_scheduled_campaigns",
		},
	}

	var result []CronJobStatus
	for _, job := range cronJobs {
		lastRun := "belum pernah berjalan"
		val, err := h.redisClient.Get(ctx, job.RedisKey).Result()
		if err == nil && val != "" {
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				lastRun = t.Format("2006-01-02 15:04:05 WIB")
			} else {
				lastRun = val
			}
		}

		result = append(result, CronJobStatus{
			Name:        job.Name,
			Schedule:    job.Schedule,
			Description: job.Description,
			LastRun:     lastRun,
		})
	}

	c.JSON(http.StatusOK, response.Success("Berhasil mengambil status cron jobs", result))
}
