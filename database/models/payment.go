package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Payment struct {
	ID                uuid.UUID       `json:"id" gorm:"primaryKey"`
	MerchantID        uuid.UUID       `json:"merchant_id" gorm:"not null"`
	PaymentReference  string          `json:"payment_reference" gorm:"unique;not null"`
	Amount            decimal.Decimal `json:"amount" gorm:"not null"`
	Currency          string          `json:"currency" gorm:"not null"`
	Status            string          `json:"status" gorm:"not null"`
	Description       *string         `json:"description"`
	CustomerReference *string         `json:"customer_reference"`
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"autoUpdateTime;not null"`
}