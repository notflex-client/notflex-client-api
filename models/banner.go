package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Banner struct {
	ID          string    `gorm:"type:uuid;primaryKey" json:"id"`
	MovieID     *string   `gorm:"type:uuid;index" json:"movie_id"`
	Movie       *Movie    `gorm:"foreignKey:MovieID" json:"movie,omitempty"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	ImageURL    string    `gorm:"type:text;not null" json:"image_url"`
	LinkURL     string    `gorm:"type:text" json:"link_url"`
	Position    int       `gorm:"not null;default:0" json:"position"`
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

func (b *Banner) BeforeCreate(_ *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	return nil
}
