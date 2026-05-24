package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProfileTransfer struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	FromUserID   string    `gorm:"type:uuid;not null" json:"from_user_id"`
	ToUserID     string    `gorm:"type:uuid;not null" json:"to_user_id"`
	ProfileName  string    `gorm:"type:varchar(100)" json:"profile_name"`
	Status       string    `gorm:"type:varchar(20);default:'pending'" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (pt *ProfileTransfer) BeforeCreate(tx *gorm.DB) (err error) {
	if pt.ID == "" {
		pt.ID = uuid.NewString()
	}
	return
}
