package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserToken struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

func (t *UserToken) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	return nil
}
