package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionPlan struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(50);not null" json:"name"`
	Price        float64   `gorm:"type:numeric(10,2);not null" json:"price"`
	DurationDays int       `gorm:"not null" json:"duration_days"`
	Description  *string   `gorm:"type:text" json:"description"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserSubscription struct {
	ID        string           `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    string           `gorm:"type:uuid;not null;index" json:"user_id"`
	PlanID    int              `gorm:"not null" json:"plan_id"`
	Plan      SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
	StartDate time.Time        `gorm:"not null" json:"start_date"`
	EndDate   time.Time        `gorm:"not null;index" json:"end_date"`
	Status    string           `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
	CreatedAt time.Time        `json:"created_at"`
}

func (s *UserSubscription) BeforeCreate(_ *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	if s.StartDate.IsZero() {
		s.StartDate = time.Now()
	}
	return nil
}

type Payment struct {
	ID             string            `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         string            `gorm:"type:uuid;not null;index" json:"user_id"`
	SubscriptionID *string           `gorm:"type:uuid;index" json:"subscription_id"`
	Subscription   *UserSubscription `gorm:"foreignKey:SubscriptionID" json:"subscription,omitempty"`
	Amount         float64           `gorm:"type:numeric(10,2);not null" json:"amount"`
	PaymentMethod  string            `gorm:"type:varchar(50)" json:"payment_method"`
	Status         string            `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	TransactionID  string            `gorm:"type:varchar(255)" json:"transaction_id"`
	CreatedAt      time.Time         `json:"created_at"`
}

func (p *Payment) BeforeCreate(_ *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}
