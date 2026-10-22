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

const BankToWalletTransferSignal = "BankToWalletTransferResult"

func TopUpFlow(
	ctx workflow.Context,
	request *dto.TopupWalletBalanceRequest,
	merchantID uuid.UUID,
) (*models.Wallets, error) {
	workflowID := "bank-to-wallet-" + merchantID.String()
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

	var wallet *models.Wallets
	err := workflow.ExecuteActivity(
		ctx,
		"GetWalletByMerchantID",
		merchantID,
	).Get(ctx, &wallet)

	if err != nil {
		return nil, err
	}

	var linkedBankAccount *models.LinkedBankAccount

	errCF := workflow.ExecuteActivity(
		ctx,
		"GetBankAccountByMerchantID",
		merchantID,
	).Get(ctx, &linkedBankAccount)

	if errCF != nil {
		return nil, errCF
	}

	bankToWalletTransferEventPayload := &dto.CreateBankToWalletDebitEvent{
		EventID: uuid.New(),
		EventType: string(constants.EventTypeBankToWalletDebitRequested),
		Payload: &dto.CreateBankToWalletRequest{
			ReferenecID: uuid.New(),
			WalletID: wallet.ID,
			BankAccountID: linkedBankAccount.BankAccountID,
			Amount: request.Amount,
			Currency: "INR",
		},
	}

	var event *models.Event
	err = workflow.ExecuteActivity(
		ctx,
		"PublishBankToWalletTransferEvent",
		bankToWalletTransferEventPayload,
		workflowID,
	).Get(ctx, &event)
	if err != nil {
		return nil, err
	}

	signalCh := workflow.GetSignalChannel(
		ctx,
		BankToWalletTransferSignal,
	)

	timer := workflow.NewTimer(ctx, 10*time.Minute)
	var bankToWalletResponse *dto.BankToWalletTransferCompletedPayload
	var timedOut bool
	var receivedMatchingResponse bool

	for !receivedMatchingResponse && !timedOut {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(
			signalCh,
			func(c workflow.ReceiveChannel, more bool) {
				var response dto.BankToWalletTransferCompletedPayload
				c.Receive(ctx, &response)

				if response.Payload.RequestedEventID == event.EventID {
					bankToWalletResponse = &response
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

	var updatedWallet *models.Wallets
	errW := workflow.ExecuteActivity(
		ctx,
		"ExecuteBankToWallet",
		bankToWalletResponse.Event,
		merchantID,
		request,
	).Get(ctx, &updatedWallet)

	if errW == nil{
		return updatedWallet, nil 
	}

	failedStatus := constants.EventStatusFailed

	var finalEvent *models.Event
	failErr := workflow.ExecuteActivity(
		ctx,
		"UpdateEventStatus",
		bankToWalletResponse.Event.EventID,
		failedStatus,
	).Get(ctx, &finalEvent)

	if failErr != nil {
		return nil, fmt.Errorf(
			"event failed: %w; failed to mark event as failed: %v",
			errW,
			failErr,
		)
	}

	return wallet, nil
}
