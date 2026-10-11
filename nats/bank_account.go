package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Srivastava-samarth/sampay/constants"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/security"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func (c *NatsClient) handleBankAccountCreated(
    ctx context.Context,
    event *security.SignedEvent,
) (*dto.BankAccountCreatedPayload, error) {

	newEvent, errE := c.repo.CreateEvent(context.Background(), &dto.CreateEventRequest{
		EventID: event.EventID,
		EventType: string(constants.EventTypeBankToSampayBankAccountCreated),
		Direction: string(constants.EventDirectionInbound),
		Payload: datatypes.JSON(event.Payload),
		Status: constants.ComplianceStatusPending,
	})

	if errE != nil{
		return nil, errE
	}

    var response dto.BankAccountEventResponsePayload

    if err := json.Unmarshal(event.Payload, &response); err != nil {
        return nil, fmt.Errorf(
            "unmarshal bank account response payload: %w",
            err,
        )
    }

    if response.EventID == uuid.Nil {
        return nil, fmt.Errorf("bank account response event ID is missing")
    }

    if response.RequestedEventID == uuid.Nil {
        return nil, fmt.Errorf("requested event ID is missing")
    }

    if response.BankAccountID == uuid.Nil {
        return nil, fmt.Errorf("bank account ID is missing")
    }

    if len(response.Payload) == 0 {
        return nil, fmt.Errorf("bank account model payload is empty")
    }

	
    return &dto.BankAccountCreatedPayload{
		Event: newEvent,
		Payload: &response,
	}, nil
}