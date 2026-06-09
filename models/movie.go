package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Genre struct {
	ID   int    `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(50);unique;not null" json:"name"`
}

type Tag struct {
	ID   int    `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(50);unique;not null" json:"name"`
	Slug string `gorm:"type:varchar(80);unique;not null" json:"slug"`
}

type Movie struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	Title        string    `gorm:"type:varchar(255);not null" json:"title"`
	Type         string    `gorm:"type:varchar(20);default:'movie'" json:"type"`
	Description  *string   `gorm:"type:text" json:"description"`
	PosterURL    *string   `gorm:"type:text" json:"poster_url"`
	BannerURL    *string   `gorm:"type:text" json:"banner_url"`
	VideoURL     *string   `gorm:"type:text" json:"video_url,omitempty"`
	TrailerURL   *string   `gorm:"type:text" json:"trailer_url"`
	DurationMins *int      `json:"duration_mins"`
	ReleaseYear  *int16    `json:"release_year"`
	Rating       *string   `gorm:"type:varchar(20)" json:"rating"`
	Language     *string   `gorm:"type:varchar(20);index" json:"language"`
	AvgRating    float64   `gorm:"type:numeric(3,1);default:0" json:"avg_rating"`
	IsPremium    bool      `gorm:"default:true" json:"is_premium"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Genres       []Genre      `gorm:"many2many:movie_genres;" json:"genres,omitempty"`
	Tags         []Tag        `gorm:"many2many:movie_tags;" json:"tags,omitempty"`
	Episodes     []Episode    `gorm:"foreignKey:MovieID" json:"episodes,omitempty"`
	Subtitles    []Subtitle   `gorm:"foreignKey:MovieID" json:"subtitles,omitempty"`
	AudioTracks  []AudioTrack `gorm:"foreignKey:MovieID" json:"audio_tracks,omitempty"`
}

func (m *Movie) BeforeCreate(_ *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

func (e *Episode) BeforeCreate(_ *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}

func (w *WatchHistory) BeforeCreate(_ *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.NewString()
	}
	return nil
}

type Episode struct {
	ID            string    `gorm:"type:uuid;primaryKey" json:"id"`
	MovieID       string    `gorm:"type:uuid;not null;index" json:"movie_id"`
	SeasonNumber  int       `gorm:"not null;default:1" json:"season_number"`
	EpisodeNumber int       `gorm:"not null;default:1" json:"episode_number"`
	Title         string    `gorm:"type:varchar(255);not null" json:"title"`
	Description   *string   `gorm:"type:text" json:"description"`
	VideoURL      *string   `gorm:"type:text" json:"video_url,omitempty"`
	DurationMins  *int      `json:"duration_mins"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type WatchHistory struct {
	ID            string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        string    `gorm:"type:uuid;not null;index" json:"user_id"`
	ProfileID     string    `gorm:"type:uuid;index" json:"profile_id"`
	MovieID       string    `gorm:"type:uuid;not null;index" json:"movie_id"`
	Movie         *Movie    `gorm:"foreignKey:MovieID" json:"movie,omitempty"`
	WatchedAt     time.Time `gorm:"not null;default:now()" json:"watched_at"`
	WatchDuration int       `gorm:"not null;default:0" json:"watch_duration"`
	IsCompleted   bool      `gorm:"not null;default:false" json:"is_completed"`
}

type MovieRating struct {
	ProfileID string    `gorm:"type:uuid;primaryKey" json:"profile_id"`
	MovieID   string    `gorm:"type:uuid;primaryKey" json:"movie_id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Rating    int       `gorm:"type:smallint;not null" json:"rating"`
	RatedAt   time.Time `gorm:"not null;default:now()" json:"rated_at"`
}
