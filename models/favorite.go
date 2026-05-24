package models

import "time"

type Favorite struct {
	UserID    string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	MovieID   string    `gorm:"type:uuid;primaryKey" json:"movie_id"`
	Movie     *Movie    `gorm:"foreignKey:MovieID" json:"movie,omitempty"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (Favorite) TableName() string {
	return "favorites"
}
