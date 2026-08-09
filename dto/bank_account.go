package dto

import "github.com/google/uuid"

type CreateBankAccountRequest struct {
	BankName      string `json:"bank_name" validate:"required"`
	AccountNumber string `json:"account_number" validate:"required"`
	AccountName   string `json:"account_name" validate:"required"`
	AccountType   string `json:"account_type" validate:"required"`
	MerchantID    uuid.UUID `json:"merchant_id" validate:"required"`
}