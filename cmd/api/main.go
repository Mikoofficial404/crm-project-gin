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
	"os"

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

	campaignRepo := postgres.NewCampaignRepository(database.GetDB())
	campaignRecipientRepo := postgres.NewCampaignRecipientRepository(database.GetDB())

	mux := asynq.NewServeMux()
	mux.HandleFunc("email:send", worker.HandleSendEmailTask)
	go func() {
		if err := srvAsynq.Run(mux); err != nil {
			logrus.Fatalf("Gagal menyalakan Server Asynq: %v", err)
		}
	}()

	userRepo := postgres.NewUserRepository(database.GetDB())
	authService := service.NewUserService(userRepo, rdb, clientAsynq)
	authHandler := v1.NewUserHandler(authService, rdb)

	dealRepo := postgres.NewDealRepository(database.GetDB())
	pipelineStageRepo := postgres.NewPipelineStageRepository(database.GetDB())
	dealProductRepo := postgres.NewDealProductRepository(database.GetDB())

	invoiceRepo := postgres.NewInvoiceRepository(database.GetDB())
	auditRepo := postgres.NewAuditRepository(database.GetDB())
	dealService := service.NewDealService(dealRepo, auditRepo, clientAsynq, invoiceRepo, rdb, pipelineStageRepo, dealProductRepo)
	dealHandler := v1.NewDealHandler(dealService)

	leadRepo := postgres.NewLeadRepository(database.GetDB())

	contactRepo := postgres.NewContactRepository(database.GetDB())
	contactService := service.NewContactService(contactRepo)
	contactHandler := v1.NewContactHandler(contactService)

	leadService := service.NewLeadService(leadRepo, userRepo, clientAsynq, rdb, auditRepo, contactRepo)

	activityRepo := postgres.NewActivityRepository(database.GetDB())
	activityService := service.NewActivityService(activityRepo)
	activityHandler := v1.NewAcitivyHandler(activityService)

	notifRepo := postgres.NewNotificationRepository(database.GetDB())
	notifService := service.NewNotificationService(notifRepo)
	notifHandler := v1.NewNotificationHandler(notifService)

	searchService := service.NewSearchService(leadRepo, dealRepo, userRepo, contactRepo)
	searchHandler := v1.NewServiceHandler(searchService)

	taskRepo := postgres.NewTaskRepository(database.GetDB())
	taskService := service.NewTaskService(taskRepo, userRepo, notifService)
	taskHandler := v1.NewTaskHandler(taskService)

	pipelineRepo := postgres.NewPipelineRepository(database.GetDB())
	pipelineService := service.NewPipelineService(pipelineRepo, pipelineStageRepo)
	pipelineHandler := v1.NewPipelineHandler(pipelineService)

	productRepo := postgres.NewProductRepository(database.GetDB())
	productService := service.NewProductService(productRepo)
	productHandler := v1.NewProductHandler(productService)

	reportRepo := postgres.NewReportRepository(database.GetDB())
	reportService := service.NewReportService(reportRepo, rdb)
	reportHandler := v1.NewReportHandler(reportService)

	campaignService := service.NewCampaignService(campaignRepo, campaignRecipientRepo, leadRepo, contactRepo, clientAsynq)
	campaignHandler := v1.NewCampaignHandler(campaignService)

	mux.HandleFunc("email:campaign", worker.NewHandleCampaignEmailTask(
		campaignRecipientRepo.UpdateRecipientStatus,
		campaignService.MarkCampaignSentIfDone,
	))

	r.POST("/api/v1/auth/forgot-password", authHandler.ForgotPassword)
	r.POST("/api/v1/auth/reset-password", authHandler.ResetPassword)
	r.POST("/api/v1/register", authHandler.Register)
	r.POST("/api/v1/login", middleware.RateLimitMiddleware(rdb), authHandler.Login)
	r.POST("/api/v1/login/verify-otp", authHandler.VerifyOTP)
	r.GET("/api/v1/auth/google/login", authHandler.LoginGoogle)
	r.GET("/api/v1/auth/google/callback", authHandler.CallbackGoogle)

	waService := service.NewWhatsAppService(
		os.Getenv("WA_GOWA_URL"),
		os.Getenv("WA_DEVICE_ID"),
		os.Getenv("WA_BASIC_AUTH"),
	)

	aiService := service.NewAIService(os.Getenv("GEMINI_API_KEY"))

	leadHandle := v1.NewLeadHandler(leadService, leadRepo, waService, activityService)

	webhookHandler := v1.NewWebhookHandler(leadService, waService, userRepo, leadRepo, activityService, notifService, rdb, aiService, contactService)
	r.POST("/api/v1/webhook/whatsapp", webhookHandler.ReceiveWhatsApp)
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(rdb))

	protected.POST("/logout", authHandler.Logout)
	protected.GET("/me", authHandler.GetMe)
	protected.PATCH("/profile/password", authHandler.ChangePassword)
	protected.GET("/profile/2fa/setup", authHandler.Setup2FA)

	protected.POST("/leads", leadHandle.CreateLeader)
	protected.POST("/leads/import", leadHandle.ImportCSV)
	protected.GET("/leads", leadHandle.GetLeads)
	protected.GET("/leads/:id", leadHandle.GetLeadByID)
	protected.PATCH("/leads/:id/status", leadHandle.UpdateStatusLeads)
	protected.PUT("/leads/:id", leadHandle.UpdateLead)
	protected.DELETE("/leads/:id", middleware.RoleMiddleware("admin"), leadHandle.DeleteLead)
	protected.GET("/leads/trash", middleware.RoleMiddleware("admin"), leadHandle.GetTrashedLeads)
	protected.POST("/leads/trash/:id/restore", middleware.RoleMiddleware("admin"), leadHandle.RestoreLead)

	protected.POST("/deals", dealHandler.CreateDeal)
	protected.GET("/deals", dealHandler.GetDeals)
	protected.PATCH("/deals/:id/stage", dealHandler.UpdateStage)
	protected.PUT("/deals/:id", dealHandler.UpdateDeal)
	protected.DELETE("/deals/:id", middleware.RoleMiddleware("admin"), dealHandler.DeleteDeal)
	protected.GET("/deals/export/pdf", dealHandler.ExportPDF)
	protected.GET("/deals/export/excel", dealHandler.ExportExcel)
	protected.GET("/deals/:id/invoice", dealHandler.DownloadInvoice)
	protected.POST("/deals/:id/products", dealHandler.AssignProduct)
	protected.DELETE("/deals/:id/products/:productId", dealHandler.RemoveProduct)
	protected.GET("/deals/:id/products", dealHandler.GetDealProducts)

	protected.PATCH("/invoices/:id/status", dealHandler.UpdateInvoiceStatus)

	protected.POST("/activities", activityHandler.CreateActivity)
	protected.GET("/activities/:lead_id", activityHandler.GetActivities)
	protected.PUT("/activities/:id", activityHandler.UpdateActivity)
	protected.DELETE("/activities/:id", middleware.RoleMiddleware("admin"), activityHandler.DeleteActivity)
	protected.POST("/upload", activityHandler.UploadFile)

	protected.GET("/search", searchHandler.GlobalSearch)

	protected.GET("/ws", websocket.ConnectWs)
	dashboardRepo := postgres.NewDashboardRepository(database.GetDB())
	dashboardService := service.NewDashboardService(dashboardRepo, rdb)
	dashboardHandler := v1.NewDashboardHandler(dashboardService)

	protected.GET("/deals/export", dealHandler.ExportCSC)
	protected.GET("/dashboard", dashboardHandler.GetDashboardStats)
	protected.GET("/analytics/forecasting", dashboardHandler.GetAnalytics)

	protected.GET("/notifications", notifHandler.GetMyNotifications)
	protected.PATCH("/notifications/read-all", notifHandler.MarkAllAsRead)
	protected.PATCH("/notifications/:id/read", notifHandler.MarkAsRead)

	protected.POST("/leads/:id/reply", leadHandle.ReplyWhatsApp)

	protected.POST("/tasks", taskHandler.CreateTask)
	protected.GET("/tasks", taskHandler.GetAllTasks)
	protected.GET("/tasks/:id", taskHandler.GetTaskByID)
	protected.PUT("/tasks/:id", taskHandler.UpdateTask)
	protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
	protected.PATCH("/tasks/:id/done", taskHandler.MarkAsDone)

	protected.POST("/contacts", contactHandler.CreateContact)
	protected.GET("/contacts", contactHandler.GetAllContacts)
	protected.GET("/contacts/:id", contactHandler.GetContactByID)
	protected.PATCH("/contacts/:id", contactHandler.UpdateContact)
	protected.DELETE("/contacts/:id", contactHandler.DeleteContact)

	adminPipeline := protected.Group("/pipelines")
	adminPipeline.Use(middleware.RoleMiddleware("admin"))
	adminPipeline.POST("", pipelineHandler.CreatePipeline)
	adminPipeline.GET("", pipelineHandler.GetAllPipelines)
	adminPipeline.GET("/:id", pipelineHandler.GetPipelineByID)
	adminPipeline.PATCH("/:id", pipelineHandler.UpdatePipeline)
	adminPipeline.DELETE("/:id", pipelineHandler.DeletePipeline)
	adminPipeline.POST("/:id/stages", pipelineHandler.AddStage)
	adminPipeline.PATCH("/:id/stages/:stageId", pipelineHandler.UpdateStage)
	adminPipeline.DELETE("/:id/stages/:stageId", pipelineHandler.DeleteStage)
	adminPipeline.PATCH("/:id/stages/reorder", pipelineHandler.ReorderStages)

	protected.GET("/reports/sales-summary", middleware.RoleMiddleware("admin"), reportHandler.GetSalesSummary)
	protected.GET("/reports/pipeline", middleware.RoleMiddleware("admin"), reportHandler.GetPipelineReport)
	protected.GET("/reports/sales-performance", middleware.RoleMiddleware("admin"), reportHandler.GetSalesPerformance)
	protected.GET("/reports/lead-source", middleware.RoleMiddleware("admin"), reportHandler.GetLeadSourceReport)
	protected.GET("/reports/activity", middleware.RoleMiddleware("admin"), reportHandler.GetActivityReport)

	adminProduct := protected.Group("/products")
	adminProduct.Use(middleware.RoleMiddleware("admin"))
	adminProduct.POST("", productHandler.CreateProduct)
	adminProduct.GET("", productHandler.GetAllProducts)
	adminProduct.GET("/:id", productHandler.GetProductByID)
	adminProduct.PATCH("/:id", productHandler.UpdateProduct)
	adminProduct.DELETE("/:id", productHandler.DeleteProduct)

	adminGroup := protected.Group("/admin")
	adminGroup.Use(middleware.RoleMiddleware("admin"))
	adminGroup.GET("/users", authHandler.GetUsers)

	adminCampaign := protected.Group("/campaigns")
	adminCampaign.Use(middleware.RoleMiddleware("admin"))
	adminCampaign.POST("", campaignHandler.CreateCampaign)
	adminCampaign.GET("", campaignHandler.GetAllCampaigns)
	adminCampaign.GET("/:id", campaignHandler.GetCampaignByID)
	adminCampaign.PATCH("/:id", campaignHandler.UpdateCampaign)
	adminCampaign.DELETE("/:id", campaignHandler.DeleteCampaign)
	adminCampaign.POST("/:id/recipients", campaignHandler.AddRecipients)
	adminCampaign.GET("/:id/recipients", campaignHandler.GetRecipients)
	adminCampaign.POST("/:id/send", campaignHandler.SendCampaign)
	adminCampaign.POST("/:id/schedule", campaignHandler.ScheduleCampaign)
	adminCampaign.GET("/:id/stats", campaignHandler.GetCampaignStats)

	c := cron.New()

	c.AddFunc("* * * * *", func() {
		leadService.CheckStaleLeads()
	})
	c.AddFunc("* * * * *", func() {
		campaignService.ProcessScheduledCampaigns()
	})
	c.AddFunc("0 8 * * *", func() {
		taskService.SendDueDateReminders()
	})
	c.Start()

	r.Run(":8080")
}
