package workflows

import (
	"fmt"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const CompanyBankToVaultSignal = "CompanyBankToVaultTransferResult"

func CompanyVaultTopUpFlow(
	ctx workflow.Context,
	request *dto.VaultTopupRequest,
) (*models.Vault, error) {
	workflowID := "bank-to-vault-" + request.VaultID.String()
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 2,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Second * 30,
			MaximumAttempts:    5,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var vault *models.Vault
	err := workflow.ExecuteActivity(
		ctx,
		"GetVault",
		constants.CompanyVault,
	).Get(ctx, &vault)

	if err != nil {
		return nil, err
	}

	bankToVaultTransferEventPayload := &dto.CreateVaultTopupEvent{
		EventID: uuid.New(),
		EventType: string(constants.EventTypeVaultTopupRequested),
		Payload: &dto.VaultTopupRequest{
			ReferenecID: request.ReferenecID,
			VaultID: vault.ID,
			Amount: request.Amount,
			Currency: "INR",
		},
	}

	var event *models.Event
	err = workflow.ExecuteActivity(
		ctx,
		"PublishVaultTopupEvent",
		bankToVaultTransferEventPayload,
		workflowID,
	).Get(ctx, &event)
	if err != nil {
		return nil, err
	}

	signalCh := workflow.GetSignalChannel(
		ctx,
		CompanyBankToVaultSignal,
	)

	timer := workflow.NewTimer(ctx, 10*time.Minute)
	var bankToVaultResponse *dto.CompanyBankToVaultTransferCompletedPayload
	var timedOut bool
	var receivedMatchingResponse bool

	for !receivedMatchingResponse && !timedOut {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(
			signalCh,
			func(c workflow.ReceiveChannel, more bool) {
				var response dto.CompanyBankToVaultTransferCompletedPayload
				c.Receive(ctx, &response)

				if response.Payload.RequestedEventID == event.EventID {
					bankToVaultResponse = &response
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

	var updatedVault *models.Vault
	errW := workflow.ExecuteActivity(
		ctx,
		"ExecuteVaultTopUp",
		vault,
		bankToVaultResponse.Event,
		request,
	).Get(ctx, &updatedVault)

	if errW == nil{
		return updatedVault, nil 
	}

	failedStatus := constants.EventStatusFailed

	var finalEvent *models.Event
	failErr := workflow.ExecuteActivity(
		ctx,
		"UpdateEventStatus",
		bankToVaultResponse.Event.EventID,
		failedStatus,
	).Get(ctx, &finalEvent)

	if failErr != nil {
		return nil, fmt.Errorf(
			"event failed: %w; failed to mark event as failed: %v",
			errW,
			failErr,
		)
	}

	return vault, nil
}
