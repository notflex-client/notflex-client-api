package models

import "time"

type Genre struct {
	ID   int    `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(50);unique;not null" json:"name"`
}

type Movie struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	Title        string    `gorm:"type:varchar(255);not null" json:"title"`
	Description  *string   `gorm:"type:text" json:"description"`
	PosterURL    *string   `gorm:"type:text" json:"poster_url"`
	VideoURL     *string   `gorm:"type:text" json:"video_url,omitempty"`
	TrailerURL   *string   `gorm:"type:text" json:"trailer_url"`
	DurationMins *int      `json:"duration_mins"`
	ReleaseYear  *int16    `json:"release_year"`
	AvgRating    float64   `gorm:"type:numeric(3,1);default:0" json:"avg_rating"`
	IsPremium    bool      `gorm:"default:true" json:"is_premium"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Genres       []Genre   `gorm:"many2many:movie_genres;" json:"genres,omitempty"`
}

type WatchHistory struct {
	ID            string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        string    `gorm:"type:uuid;not null;index" json:"user_id"`
	MovieID       string    `gorm:"type:uuid;not null;index" json:"movie_id"`
	Movie         *Movie    `gorm:"foreignKey:MovieID" json:"movie,omitempty"`
	WatchedAt     time.Time `gorm:"not null;default:now()" json:"watched_at"`
	WatchDuration int       `gorm:"not null;default:0" json:"watch_duration"`
	IsCompleted   bool      `gorm:"not null;default:false" json:"is_completed"`
}

type MovieRating struct {
	UserID  string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	MovieID string    `gorm:"type:uuid;primaryKey" json:"movie_id"`
	Rating  int       `gorm:"type:smallint;not null" json:"rating"`
	RatedAt time.Time `gorm:"not null;default:now()" json:"rated_at"`
}
