package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginRequest struct {
	ID               string    `gorm:"type:uuid;primaryKey" json:"id"`
	Email            string    `gorm:"type:varchar(255);not null;index" json:"email"`
	ConfirmationCode string    `gorm:"type:varchar(6);not null" json:"-"`
	ExpireAt         time.Time `gorm:"not null" json:"expire_at"`
	CreatedAt        time.Time `json:"created_at"`
}

func (lr *LoginRequest) BeforeCreate(_ *gorm.DB) error {
	if lr.ID == "" {
		lr.ID = uuid.NewString()
	}
	return nil
}
