package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ReconResult struct {
	ReconDate          time.Time
	MatchedCount       *int
	FailedCount        *int
	TotalDebit         decimal.Decimal
	TotalCredit        decimal.Decimal
	TotalTransactions  *int
	FailedTransactions []ReconFailure
}

type ReconFailure struct {
	LedgerTransactionID uuid.UUID       `json:"ledger_transaction_id"`
	ReferenceID         string          `json:"reference_id"`
	TotalDebit          decimal.Decimal `json:"total_debit"`
	TotalCredit         decimal.Decimal `json:"total_credit"`
	Difference          decimal.Decimal `json:"difference"`
}

type ReconReport struct {
	ReconDate          string
	StartTimestamp     string
	EndTimestamp       string
	TotalTransactions  int
	MatchedCount       int
	FailedCount        int
	TotalDebit         string
	TotalCredit        string
	Difference         string
	Status             string
	FailedTransactions []ReconFailure
}
type CreateRefundRequest struct {
	PaymentID         uuid.UUID       `json:"payment_id"`
	MerchantID        uuid.UUID       `json:"merchant_id"`
	ExternalReference *string         `json:"external_reference"`
	Amount            decimal.Decimal `json:"amount"`
	Currency          string          `json:"currency"`
	Reason            *string         `json:"reason"`
}
