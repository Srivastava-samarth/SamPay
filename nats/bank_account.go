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

	newEvent, errE := c.repo.CreateEvent(ctx, &dto.CreateEventRequest{
		EventID:   event.EventID,
		EventType: string(constants.EventTypeBankToSampayBankAccountCreated),
		Direction: string(constants.EventDirectionInbound),
		WorkflowID: "bank-to-sampay",
		Payload:   datatypes.JSON(event.Payload),
		Status:    constants.ComplianceStatusPending,
	})

	if errE != nil {
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
		Event:   newEvent,
		Payload: &response,
	}, nil
}

func (c *NatsClient) handleBankAccountCredit(
	ctx context.Context,
	event *security.SignedEvent,
) (*dto.WalletToBankTransferCompletedPayload, error) {

	newEvent, errE := c.repo.CreateEvent(ctx, &dto.CreateEventRequest{
		EventID:   event.EventID,
		EventType: string(constants.EventTypeBankTransferCompleted),
		Direction: string(constants.EventDirectionInbound),
		WorkflowID: "bank-to-sampay",
		Payload:   datatypes.JSON(event.Payload),
		Status:    string(constants.EventStatusPending),
	})

	if errE != nil {
		return nil, errE
	}

	var response dto.WalletToBankEventResponsePayload

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

	if response.TransferID == uuid.Nil {
		return nil, fmt.Errorf("bank account ID is missing")
	}

	if len(response.Payload) == 0 {
		return nil, fmt.Errorf("bank account model payload is empty")
	}

	return &dto.WalletToBankTransferCompletedPayload{
		Event:   newEvent,
		Payload: &response,
	}, nil
}

func (c *NatsClient) handleBankAccountDebit(
	ctx context.Context,
	event *security.SignedEvent,
) (*dto.BankToWalletTransferCompletedPayload, error){
	newEvent, errE := c.repo.CreateEvent(ctx, &dto.CreateEventRequest{
		EventID:   event.EventID,
		EventType: string(constants.EventTypeBankToWalletTransferCompleted),
		Direction: string(constants.EventDirectionInbound),
		WorkflowID: "bank-to-sampay",
		Payload:   datatypes.JSON(event.Payload),
		Status:    string(constants.EventStatusPending),
	})

	if errE != nil {
		return nil, errE
	}

	var response dto.BankToWalletEventResponsePayload	
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

	if response.TransferID == uuid.Nil {
		return nil, fmt.Errorf("bank account ID is missing")
	}

	if len(response.Payload) == 0 {
		return nil, fmt.Errorf("bank account model payload is empty")
	}

	return &dto.BankToWalletTransferCompletedPayload{
		Event:   newEvent,
		Payload: &response,
	}, nil
}

		
func (c *NatsClient) handleBankTransfer(
	ctx context.Context,
	event *security.SignedEvent,
) (*dto.BankToBankTransferCompletedPayload, error) {

	newEvent, errE := c.repo.CreateEvent(ctx, &dto.CreateEventRequest{
		EventID:   event.EventID,
		EventType: string(constants.EventTypeBankTransferCompleted),
		Direction: string(constants.EventDirectionInbound),
		WorkflowID: "bank-to-sampay",
		Payload:   datatypes.JSON(event.Payload),
		Status:    constants.ComplianceStatusPending,
	})

	if errE != nil {
		return nil, errE
	}

	var response dto.BankToBankEventResponsePayload

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

	if response.TransferID == uuid.Nil {
		return nil, fmt.Errorf("bank account ID is missing")
	}

	if len(response.Payload) == 0 {
		return nil, fmt.Errorf("bank account model payload is empty")
	}

	return &dto.BankToBankTransferCompletedPayload{
		Event:   newEvent,
		Payload: &response,
	}, nil
}
