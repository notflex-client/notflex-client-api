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
		&models.User{},
		&models.Profile{},
		&models.UserToken{},
		&models.RegisterRequest{},
		&models.LoginRequest{},
		&models.SubscriptionPlan{},
		&models.UserSubscription{},
		&models.Payment{},
		&models.WatchHistory{},
		&models.MovieRating{},
		&models.Favorite{},
		&models.Subtitle{},
		&models.AudioTrack{},
		&models.Banner{},
		&models.ProfileTransfer{},
		&models.MovieSimilarity{},
	); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	seedTags(db)
	seedSubscriptionPlans(db)
	migrateProfileScoping(db)

	DB = db
	slog.Info("connected to PostgreSQL")
}

// migrateProfileScoping evolves a pre-existing database to the per-profile data
// model: it bootstraps a default profile for every legacy account, backfills
// profile_id on watch history / favorites / ratings, and swaps the favorites and
// movie_ratings primary keys to (profile_id, movie_id). All steps are idempotent
// and best-effort so a fresh database (already in the target shape) is unaffected.
func migrateProfileScoping(db *gorm.DB) {
	// 1. Bootstrap a default profile for any account that has none.
	var legacyUsers []models.User
	db.Where("id NOT IN (SELECT user_id FROM profiles)").Find(&legacyUsers)
	for _, u := range legacyUsers {
		name := u.FullName
		if name == "" {
			name = "Profile 1"
		}
		db.Create(&models.Profile{UserID: u.ID, Name: name, AvatarURL: u.AvatarURL})
	}

	// 2. Backfill profile_id from each account's earliest profile.
	for _, table := range []string{"watch_histories", "favorites", "movie_ratings"} {
		if err := db.Exec(`
			UPDATE `+table+` t SET profile_id = p.id
			FROM (SELECT DISTINCT ON (user_id) user_id, id FROM profiles ORDER BY user_id, created_at ASC) p
			WHERE t.user_id = p.user_id AND t.profile_id IS NULL
		`).Error; err != nil {
			slog.Warn("backfill profile_id failed", "table", table, "error", err)
		}
	}

	// 3. Swap composite primary keys to (profile_id, movie_id) on the two tables
	//    that were previously keyed by (user_id, movie_id).
	for _, table := range []string{"favorites", "movie_ratings"} {
		stmts := []string{
			`ALTER TABLE ` + table + ` ALTER COLUMN profile_id SET NOT NULL`,
			`ALTER TABLE ` + table + ` DROP CONSTRAINT IF EXISTS ` + table + `_pkey`,
			`ALTER TABLE ` + table + ` ADD CONSTRAINT ` + table + `_pkey PRIMARY KEY (profile_id, movie_id)`,
		}
		for _, s := range stmts {
			if err := db.Exec(s).Error; err != nil {
				slog.Warn("profile-scoping PK migration step failed (safe to ignore on fresh DB)", "table", table, "error", err)
			}
		}
	}
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
		{Name: "Monthly", Price: 79000, DurationDays: 30, Description: &descMonthly, MaxProfiles: 4, IsActive: true},
		{Name: "Annual", Price: 699000, DurationDays: 365, Description: &descAnnual, MaxProfiles: 5, IsActive: true},
	}

	for _, plan := range plans {
		// Assign keeps existing rows up to date with the seeded MaxProfiles value.
		db.Where(models.SubscriptionPlan{Name: plan.Name}).
			Assign(models.SubscriptionPlan{MaxProfiles: plan.MaxProfiles}).
			FirstOrCreate(&plan)
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
