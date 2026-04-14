package database

import (
	"fmt"
	"log"
	"os"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// Lấy DSN từ biến môi trường
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Println("WARNING: DB_DSN is not set in .env")
		return
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = db
	fmt.Println("Connected to PostgreSQL successfully")
	
	// Vì CSDL đã được sinh bằng file SQL chuẩn nên không chạy AutoMigrate nữa
	// Tránh thao tác ghi đè làm mất Trigger và Enum của PostgreSQL
	// DB.AutoMigrate(&models.User{})
}
