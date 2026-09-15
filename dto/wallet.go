package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WalletTransactionResponse struct {
	ID               uuid.UUID       `json:"id"`
	TransactionRef   string          `json:"transaction_ref"`
	Type             string          `json:"type"`
	ReferenceID      string          `json:"reference_id"`
	Amount           decimal.Decimal `json:"amount"`
	Currency         string          `json:"currency"`
	EntryType        string          `json:"entry_type"`
	Status           string          `json:"status"`
	SettlementStatus string          `json:"settlement_status"`
	CreatedAt        time.Time       `json:"created_at"`
}

type UpdateWalletBalanceRequest struct {
	AvailableBalance decimal.Decimal `json:"available_balance"`
	ReservedBalance  decimal.Decimal `json:"reserved_balance"`
}

type TopupWalletBalanceRequest struct {
	Amount decimal.Decimal `json:"amount" binding:"required"`
}
