package models

import "time"

type Favorite struct {
	ProfileID string    `gorm:"type:uuid;primaryKey" json:"profile_id"`
	MovieID   string    `gorm:"type:uuid;primaryKey" json:"movie_id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Movie     *Movie    `gorm:"foreignKey:MovieID" json:"movie,omitempty"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (Favorite) TableName() string {
	return "favorites"
}
