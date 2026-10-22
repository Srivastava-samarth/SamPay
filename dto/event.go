package dto

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type CreateEventRequest struct {
	EventID    uuid.UUID      `json:"event_id"`
	EventType  string         `json:"event_type"`
	Direction  string         `json:"direction"`
	WorkflowID string         `json:"workflow_id"`
	Payload    datatypes.JSON `json:"payload"`
	Status     string         `json:"status"`
}

type BankAccountCreationRequestEvent struct {
	EventID   uuid.UUID                 `json:"event_id"`
	EventType string                    `json:"event_type"`
	Payload   *CreateBankAccountRequest `json:"payload"`
}

type BankAccountEventResponsePayload struct {
	EventID          uuid.UUID      `json:"event_id"`
	RequestedEventID uuid.UUID      `json:"requested_event_id"`
	BankAccountID    uuid.UUID      `json:"bank_account_id"`
	Payload          datatypes.JSON `json:"payload"`
}

type CreateWalletToBankCreditEvent struct {
	EventID   uuid.UUID                       `json:"event_id" binding:"required"`
	EventType string                          `json:"event_type" binding:"required"`
	Payload   *CreateWalletToBankEventRequest `josn:"payload" binding:"required"`
}

type CreateWalletToBankEventRequest struct {
	ReferenecID   uuid.UUID       `json:"reference_id" binding:"required"`
	WalletID      uuid.UUID       `json:"wallet_id" binding:"required"`
	BankAccountID uuid.UUID       `json:"bank_account_id" binding:"required"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
	Currency      string          `json:"currency" binding:"required"`
}

type WalletToBankTransferCompletedPayload struct {
	Event   *models.Event
	Payload *WalletToBankEventResponsePayload
}

type BankToWalletTransferCompletedPayload struct {
	Event   *models.Event
	Payload *BankToWalletEventResponsePayload
}

type BankToWalletEventResponsePayload struct {
	EventID          uuid.UUID      `json:"event_id"`
	RequestedEventID uuid.UUID      `json:"requested_event_id"`
	TransferID       uuid.UUID      `json:"transfer_id"`
	Payload          datatypes.JSON `json:"payload"`
}

type WalletToBankEventResponsePayload struct {
	EventID          uuid.UUID      `json:"event_id"`
	RequestedEventID uuid.UUID      `json:"requested_event_id"`
	TransferID       uuid.UUID      `json:"transfer_id"`
	Payload          datatypes.JSON `json:"payload"`
}
type CreateBankToBankTransferEvent struct {
	EventID   uuid.UUID                        `json:"event_id" binding:"required"`
	EventType string                           `json:"event_type" binding:"required"`
	Payload   *CreateBankToBankTransferRequest `josn:"payload" binding:"required"`
}

type CreateBankToBankTransferRequest struct {
	ReferenecID       uuid.UUID       `json:"reference_id" binding:"required"`
	SenderAccountID   uuid.UUID       `json:"sender_account_id" binding:"required"`
	ReceiverAccountID uuid.UUID       `json:"receiver_account_id" binding:"required"`
	Amount            decimal.Decimal `json:"amount" binding:"required"`
	Currency          string          `json:"currency" binding:"required"`
}

type BankToBankTransferCompletedPayload struct {
	Event   *models.Event
	Payload *BankToBankEventResponsePayload
}

type BankToBankEventResponsePayload struct {
	EventID          uuid.UUID      `json:"event_id"`
	RequestedEventID uuid.UUID      `json:"requested_event_id"`
	TransferID       uuid.UUID      `json:"transfer_id"`
	Payload          datatypes.JSON `json:"payload"`
}

type CreateBankToWalletDebitEvent struct {
	EventID   uuid.UUID                  `json:"event_id" binding:"required"`
	EventType string                     `json:"event_type" binding:"required"`
	Payload   *CreateBankToWalletRequest `josn:"payload" binding:"required"`
}

type CreateBankToWalletRequest struct {
	ReferenecID   uuid.UUID       `json:"reference_id" binding:"required"`
	WalletID      uuid.UUID       `json:"wallet_id" binding:"required"`
	BankAccountID uuid.UUID       `json:"bank_account_id" binding:"required"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
	Currency      string          `json:"currency" binding:"required"`
}
