package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subtitle struct {
	ID          string    `gorm:"type:uuid;primaryKey" json:"id"`
	MovieID     string    `gorm:"type:uuid;not null;index" json:"movie_id"`
	Language    string    `gorm:"type:varchar(20);not null" json:"language"`
	Label       string    `gorm:"type:varchar(50)" json:"label"`
	SubtitleURL string    `gorm:"type:text;not null" json:"subtitle_url"`
	Format      string    `gorm:"type:varchar(10);default:'vtt'" json:"format"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Subtitle) BeforeCreate(_ *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	return nil
}
