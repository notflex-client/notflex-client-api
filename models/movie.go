package models

import (
	"time"
)

type Genre struct {
	ID   int    `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(50);unique;not null" json:"name"`
}

type Movie struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	Title        string    `gorm:"type:varchar(255);not null" json:"title"`
	Description  *string   `json:"description"`
	PosterURL    *string   `json:"poster_url"`
	VideoURL     *string   `json:"video_url"`
	TrailerURL   *string   `json:"trailer_url"`
	DurationMins *int      `json:"duration_mins"`
	ReleaseYear  *int16    `json:"release_year"`
	AvgRating    float64   `gorm:"type:numeric(3,1);default:0" json:"avg_rating"`
	IsPremium    bool      `gorm:"default:true" json:"is_premium"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Genres       []Genre   `gorm:"many2many:movie_genres;" json:"genres"`
}
