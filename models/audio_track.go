package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AudioTrack struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	MovieID   string    `gorm:"type:uuid;not null;index" json:"movie_id"`
	Language  string    `gorm:"type:varchar(20);not null" json:"language"`
	Label     string    `gorm:"type:varchar(50)" json:"label"`
	AudioURL  string    `gorm:"type:text;not null" json:"audio_url"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *AudioTrack) BeforeCreate(_ *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}
