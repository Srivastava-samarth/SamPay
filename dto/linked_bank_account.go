package dto

import "github.com/google/uuid"

type CreateLinkedBankAccountRequest struct {
	MerchantID    uuid.UUID `json:"merchant_id" gorm:"not null"`
	BankAccountID uuid.UUID `json:"bank_account_id" gorm:"not null"`
	Type          string    `json:"type" gorm:"not null"`
	Status        string    `json:"status" gorm:"not null"`
}

type CreateLinkedBankAccountResponse struct {
	LinkedBankAccountID uuid.UUID `json:"linked_bank_account_id" gorm:"not null"`
	MerchantID          uuid.UUID `json:"merchant_id" gorm:"not null"`
	BankAccountID       uuid.UUID `json:"bank_account_id" gorm:"not null"`
	Type                string    `json:"type" gorm:"not null"`
	Status              string    `json:"status" gorm:"not null"`
}

type UpdateLinkedBankAccountRequest struct {
	Type                string    `json:"type"`
	Status              string    `json:"status"`
}
