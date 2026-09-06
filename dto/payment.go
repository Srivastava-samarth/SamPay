package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreatePaymentRequest struct {
	ReceiverMerchantID uuid.UUID       `json:"receiver_merchant_id" gorm:"not null"`
	Amount             decimal.Decimal `json:"amount" gorm:"not null"`
	Currency           string         `json:"currency" gorm:"not null"`
	Description        string         `json:"description"`
}
