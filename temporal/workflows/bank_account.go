package workflows

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func CreateBankAccount(
	ctx workflow.Context,
	request *dto.CreateBankAccountRequest,
	merchantID uuid.UUID,
) (*models.BankAccount,error){
	workflowID := "bank-account-" + merchantID.String()
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    5,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	signalCh := workflow.GetSignalChannel(
		ctx,
		BankAccountCreationResultSignal,
	)

	bankAccoutnCreationEventPayload := &dto.BankAccountCreationRequestEvent{
		EventID: uuid.New(),
		EventType: string(constants.EventTypeSampayToBankAccountCreateRequested),
		Payload: &dto.CreateBankAccountRequest{
			AccountName: request.AccountName,
			AccountType: request.AccountType,
		},
	}

	var event *models.Event
	err := workflow.ExecuteActivity(
		ctx,
		"PublishBankAccountCreationRequest",
		bankAccoutnCreationEventPayload,
		workflowID,
	).Get(ctx, &event)
	if err != nil {
		return nil, err
	}


	timer := workflow.NewTimer(ctx, 10*time.Minute)
	var bankResponse dto.BankAccountCreatedPayload
	var timedOut bool
	var receivedMatchingResponse bool

	for !receivedMatchingResponse && !timedOut {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(
			signalCh,
			func(c workflow.ReceiveChannel, more bool) {
				var response dto.BankAccountCreatedPayload
				c.Receive(ctx, &response)

				if response.Payload.RequestedEventID == event.EventID {
					bankResponse = response
					receivedMatchingResponse = true
				}
			},
		)

		selector.AddFuture(
			timer,
			func(workflow.Future) {
				timedOut = true
			},
		)

		selector.Select(ctx)
	}

	if timedOut && !receivedMatchingResponse {
		return nil, fmt.Errorf(
			"timed out waiting for bank account creation: %s",
			event.EventID,
		)
	}

	var bankAccount *models.BankAccount
	if err := json.Unmarshal(bankResponse.Payload.Payload, &bankAccount); err != nil {
		return nil, fmt.Errorf("unmarshal bank account model: %w", err)
	}

	err = workflow.ExecuteActivity(
		ctx,
		"CreateLinkedBankAccount",
		bankAccount,
		merchantID,
	).Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	return bankAccount, nil
}