package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	models "github.com/Srivastava-samarth/sampay/database/models"
)

type CreateBankAccountRequest struct {
	AccountName string `json:"account_name"`
	AccountType string `json:"account_type"`
}

type CreateBankAccountResponse struct {
	ID            uuid.UUID `json:"id" gorm:"primaryKey"`
	MerchantID    uuid.UUID `json:"merchant_id" gorm:"not null"`
	AccountNumber string    `json:"-" gorm:"not null"`
	AccountName   string    `json:"account_name" gorm:"not null"`
	BankName      string    `json:"bank_name" gorm:"not null"`
	IFSCCode      string    `json:"ifsc_code" gorm:"not null"`
	AccountType   string    `json:"account_type" gorm:"not null"`
	Status        string    `json:"status" gorm:"default:active;not null"`
}

type UpdateBankAccountRequest struct {
	ID          uuid.UUID       `json:"id" gorm:"not null"`
	AccountName string          `json:"account_name"`
	AccountType string          `json:"account_type"`
	Status      string          `json:"status"`
	Balance     decimal.Decimal `json:"balance"`
}

type BankAccountCreatedPayload struct {
	Event   *models.Event
	Payload *BankAccountEventResponsePayload
}
