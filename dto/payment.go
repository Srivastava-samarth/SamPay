package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreatePaymentRequest struct {
	ReceiverMerchantID uuid.UUID       `json:"receiver_merchant_id" binding:"required"`
	Amount             decimal.Decimal `json:"amount" binding:"required"`
	Currency           string          `json:"currency" binding:"required"`
	Description        string          `json:"description"`
}
