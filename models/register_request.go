package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	ID               string    `gorm:"type:uuid;primaryKey" json:"id"`
	Email            string    `gorm:"type:varchar(255);not null;index" json:"email"`
	ConfirmationCode string    `gorm:"type:varchar(6);not null" json:"-"`
	Verified         bool      `gorm:"not null;default:false" json:"verified"`
	ExpireAt         time.Time `gorm:"not null" json:"expire_at"`
	CreatedAt        time.Time `json:"created_at"`
}

func (rr *RegisterRequest) BeforeCreate(_ *gorm.DB) error {
	if rr.ID == "" {
		rr.ID = uuid.NewString()
	}
	return nil
}
