package database

import (
	"log"
	"log/slog"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"notflex_client_api/models"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// AutoMigrate chỉ cho các bảng mới (không có trong init_schema.sql gốc)
	// Không ảnh hưởng đến trigger và enum đã có
	if err := db.AutoMigrate(
		&models.UserToken{},
		&models.RegisterRequest{},
		&models.WatchHistory{},
	); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	DB = db
	slog.Info("connected to PostgreSQL")
}

func CloseDB() {
	if DB == nil {
		return
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return
	}
	sqlDB.Close()
	slog.Info("database connection closed")
}
