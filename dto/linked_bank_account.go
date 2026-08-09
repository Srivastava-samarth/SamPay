package dto

import "github.com/google/uuid"

type CreateLinkedBankAccountRequest struct {
	MerchantID    uuid.UUID `json:"merchant_id" validate:"required"`
	BankAccountID uuid.UUID `json:"bank_account_id" validate:"required"`
	Type          string    `json:"type" validate:"required"`
	Status        string    `json:"status" validate:"required"`
}
