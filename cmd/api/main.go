package main

import (
	"crm-project/internal/delivery/http/middleware"
	v1 "crm-project/internal/delivery/http/v1"
	"crm-project/internal/repository/postgres"
	"crm-project/internal/service"
	"crm-project/pkg/database"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

	userRepo := postgres.NewUserRepository(database.GetDB())
	authService := service.NewUserService(userRepo)
	authHandler := v1.NewUserHandler(authService)

	leadRepo := postgres.NewLeadRepository(database.GetDB())
	leadService := service.NewLeadService(leadRepo)
	leadHandle := v1.NewLeadHandler(leadService)

	r.POST("/api/v1/register", authHandler.Register)
	r.POST("/api/v1/login", authHandler.Login)
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	protected.POST("/leads", leadHandle.CreateLeader)
	protected.GET("/leads", leadHandle.GetLeads)
	protected.PATCH("/leads/:id/status", leadHandle.UpdateStatusLeads)
	protected.GET("/dashboard", func(c *gin.Context) {
		userID := c.MustGet("user_id")
		c.JSON(200, gin.H{"message": "Selamat datang di area rahasia!", "user_id": userID})
	})

	r.Run(":8080")

}
