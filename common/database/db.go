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

	if err := db.AutoMigrate(
		&models.Genre{},
		&models.Tag{},
		&models.Movie{},
		&models.Episode{},
		&models.UserToken{},
		&models.RegisterRequest{},
		&models.SubscriptionPlan{},
		&models.UserSubscription{},
		&models.Payment{},
		&models.WatchHistory{},
		&models.MovieRating{},
		&models.LoginRequest{},
		&models.ProfileTransfer{},
	); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	seedTags(db)
	seedSubscriptionPlans(db)

	DB = db
	slog.Info("connected to PostgreSQL")
}

func seedTags(db *gorm.DB) {
	tags := []models.Tag{
		{Name: "Trending", Slug: "trending"},
		{Name: "Top 10", Slug: "top-10"},
		{Name: "New on Netflix", Slug: "new-on-netflix"},
		{Name: "Korean", Slug: "korean"},
		{Name: "Netflix Originals", Slug: "netflix-originals"},
		{Name: "Weekend", Slug: "weekend"},
		{Name: "Critically Acclaimed", Slug: "critically-acclaimed"},
		{Name: "Fresh Picks", Slug: "fresh-picks"},
		{Name: "Animation", Slug: "animation"},
		{Name: "Action", Slug: "action"},
		{Name: "Romance", Slug: "romance"},
	}

	for _, tag := range tags {
		db.FirstOrCreate(&tag, models.Tag{Slug: tag.Slug})
	}
}

func seedSubscriptionPlans(db *gorm.DB) {
	descMonthly := "Xem phim premium không giới hạn trong 30 ngày"
	descAnnual := "Gói năm tiết kiệm cho người xem thường xuyên"
	plans := []models.SubscriptionPlan{
		{Name: "Monthly", Price: 79000, DurationDays: 30, Description: &descMonthly, IsActive: true},
		{Name: "Annual", Price: 699000, DurationDays: 365, Description: &descAnnual, IsActive: true},
	}

	for _, plan := range plans {
		db.Where(models.SubscriptionPlan{Name: plan.Name}).FirstOrCreate(&plan)
	}
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
