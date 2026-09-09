package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateRefundRequest struct {
	PaymentID         uuid.UUID       `json:"payment_id"`
	MerchantID        uuid.UUID       `json:"merchant_id"`
	ExternalReference *string         `json:"external_reference"`
	Amount            decimal.Decimal `json:"amount"`
	Currency          string          `json:"currency"`
	Reason            *string         `json:"reason"`
}