package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateWalletToBankRequest struct {
	SenderWalletID           uuid.UUID       `json:"sender_wallet_id"`
	DestinationBankAccountID uuid.UUID       `json:"destination_bank_account_id"`
	ExternalReference        *string         `json:"external_reference"`
	Amount                   decimal.Decimal `json:"amount"`
	Currency                 string          `json:"currency"`
	Description              *string         `json:"description"`
}

type CreateBankToBankRequest struct {
	SourceBankAccountID      uuid.UUID       `json:"source_bank_account_id"`
	DestinationBankAccountID uuid.UUID       `json:"destination_bank_account_id"`
	ExternalReference        *string         `json:"external_reference"`
	Amount                   decimal.Decimal `json:"amount"`
	Currency                 string          `json:"currency"`
	Description              *string         `json:"description"`
}
