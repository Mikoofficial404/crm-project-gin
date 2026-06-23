package main

import (
	"crm-project/internal/delivery/http/middleware"
	v1 "crm-project/internal/delivery/http/v1"
	"crm-project/internal/delivery/websocket"
	"crm-project/internal/repository/postgres"
	"crm-project/internal/service"
	"crm-project/internal/worker"
	"crm-project/pkg/database"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env file")
	}

	db := database.DBConn()
	sqlDb, err := db.DB()
	if err != nil {
		logrus.Fatal("Gagal menyambung ke database: ", err)
	}
	defer sqlDb.Close()

	rdb, err := database.ConnectRedis()
	if err != nil {
		logrus.Fatal(err)
	}
	defer rdb.Close()

	if err := database.MigrateDB(); err != nil {
		log.Fatal("Gagal migrasi database:", err)
	}

	r := gin.Default()

	r.Static("/uploads", "./uploads")
	clientAsynq := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6380"})
	srvAsynq := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6380"},
		asynq.Config{Concurrency: 10},
	)
	mux := asynq.NewServeMux()
	mux.HandleFunc("email:send", worker.HandleSendEmailTask)
	go func() {
		if err := srvAsynq.Run(mux); err != nil {
			logrus.Fatalf("Gagal menyalakan Server Asynq: %v", err)
		}
	}()

	userRepo := postgres.NewUserRepository(database.GetDB())
	authService := service.NewUserService(userRepo)
	authHandler := v1.NewUserHandler(authService, rdb)

	leadRepo := postgres.NewLeadRepository(database.GetDB())
	leadService := service.NewLeadService(leadRepo)
	leadHandle := v1.NewLeadHandler(leadService)

	dealRepo := postgres.NewDealRepository(database.GetDB())

	auditRepo := postgres.NewAuditRepository(database.GetDB())
	dealService := service.NewDealService(dealRepo, auditRepo, clientAsynq)
	dealHandler := v1.NewDealHandler(dealService)

	activityRepo := postgres.NewActivityRepository(database.GetDB())
	activityService := service.NewActivityService(activityRepo)
	activityHandler := v1.NewAcitivyHandler(activityService)

	r.POST("/api/v1/register", authHandler.Register)
	r.POST("/api/v1/login", middleware.RateLimitMiddleware(rdb), authHandler.Login)
	r.POST("/api/v1/login/verify-otp", authHandler.VerifyOTP)
	r.GET("/api/v1/auth/google/login", authHandler.LoginGoogle)
	r.GET("/api/v1/auth/google/callback", authHandler.CallbackGoogle)
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(rdb))

	protected.POST("/logout", authHandler.Logout)
	protected.PATCH("/profile/password", authHandler.ChangePassword)
	protected.GET("/profile/2fa/setup", authHandler.Setup2FA)

	protected.POST("/leads", leadHandle.CreateLeader)
	protected.POST("/leads/import", leadHandle.ImportCSV)
	protected.GET("/leads", leadHandle.GetLeads)
	protected.PATCH("/leads/:id/status", leadHandle.UpdateStatusLeads)
	protected.DELETE("/leads/:id", leadHandle.DeleteLead)
	protected.GET("/leads/trash", leadHandle.GetTrashedLeads)
	protected.POST("/leads/trash/:id/restore", leadHandle.RestoreLead)

	protected.POST("/deals", dealHandler.CreateDeal)
	protected.GET("/deals", dealHandler.GetDeals)
	protected.PATCH("/deals/:id/stage", dealHandler.UpdateStage)
	protected.DELETE("/deals/:id", dealHandler.DeleteDeal)
	protected.GET("/deals/export/pdf", dealHandler.ExportPDF)

	protected.POST("/activities", activityHandler.CreateActivity)
	protected.GET("/activities/:lead_id", activityHandler.GetActivities)
	protected.POST("/upload", activityHandler.UploadFile)

	protected.GET("/ws", websocket.ConnectWs)
	dashboardRepo := postgres.NewDashboardRepository(database.GetDB())
	dashboardService := service.NewDashboardService(dashboardRepo, rdb)
	dashboardHandler := v1.NewDashboardHandler(dashboardService)

	protected.GET("/deals/export", dealHandler.ExportCSC)

	protected.GET("/dashboard", dashboardHandler.GetDashboardStats)

	// adminGroup := protected.Group("/admin")
	// adminGroup.Use(middleware.RoleMiddleware("admin"))
	// adminGroup.GET("/users", func(c *gin.Context) { ... })
	//

	c := cron.New()

	c.AddFunc("* * * * *", func() {
		leadService.CheckStaleLeads()
	})
	c.Start()

	r.Run(":8080")

}
