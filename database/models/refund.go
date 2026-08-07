package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Refund struct {
	ID                uuid.UUID       `json:"id" gorm:"primaryKey"`
	PaymentID         uuid.UUID       `json:"payment_id" gorm:"not null"`
	MerchantID        uuid.UUID       `json:"merchant_id" gorm:"not null"`
	RefundReference   string          `json:"refund_reference" gorm:"unique;not null"`
	ExternalReference *string         `json:"external_reference"`
	Amount            decimal.Decimal `json:"amount" gorm:"not null"`
	Currency          string          `json:"currency" gorm:"not null"`
	Status            string          `json:"status" gorm:"default:pending;not null"`
	Reason            *string         `json:"reason"`
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
