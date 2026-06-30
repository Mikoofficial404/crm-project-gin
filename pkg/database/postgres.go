package database

import (
	"crm-project/internal/config"
	"crm-project/internal/models/entity"
	"log"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func DBConn() *gorm.DB {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	dsn := cfg.GetDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("Gagal membuka database: %v", err)
	}
	logrus.Info("Berhasil terkoneksi ke Database PostgreSQL!")
	DB = db
	return db
}

func MigrateDB() error {
	err := DB.AutoMigrate(&entity.User{}, &entity.Lead{}, &entity.Deal{}, &entity.Activity{}, &entity.Notification{}, &entity.AuditLog{}, &entity.Invoice{})
	if err != nil {
		return err
	}
	logrus.Info("Migrasi database berhasil!")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
