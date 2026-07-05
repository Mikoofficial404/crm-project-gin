package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Check(c *gin.Context) {
	type ComponentStatus struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	}

	type HealthResponse struct {
		Status     string                     `json:"status"`
		Components map[string]ComponentStatus `json:"components"`
	}

	components := map[string]ComponentStatus{}
	overallOK := true

	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		components["database"] = ComponentStatus{Status: "degraded", Message: "cannot connect to database"}
		overallOK = false
	} else {
		components["database"] = ComponentStatus{Status: "ok"}
	}

	if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
		components["redis"] = ComponentStatus{Status: "degraded", Message: "cannot connect to redis"}
		overallOK = false
	} else {
		components["redis"] = ComponentStatus{Status: "ok"}
	}

	status := "ok"
	httpCode := http.StatusOK
	if !overallOK {
		status = "degraded"
		httpCode = http.StatusServiceUnavailable
	}

	c.JSON(httpCode, HealthResponse{
		Status:     status,
		Components: components,
	})
}
