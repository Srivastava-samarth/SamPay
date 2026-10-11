package dto

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type CreateEventRequest struct {
	EventID   uuid.UUID      `json:"event_id"`
	EventType string         `json:"event_type"`
	Direction string         `json:"direction"`
	Payload   datatypes.JSON `json:"payload"`
	Status    string         `json:"status"`
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

