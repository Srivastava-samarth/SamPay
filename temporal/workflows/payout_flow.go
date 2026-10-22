package workflows

import (
	"fmt"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/temporal/activities"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const WalletToBankTransferSignal = "WalletToBankTransferResult"

func PayoutFlow(
	ctx workflow.Context,
	request *dto.CreateWalletToBankRequest,
	merchantID uuid.UUID,
) (*models.Payout, error) {
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 2,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Second * 30,
			MaximumAttempts:    5,
		},
	}

	workflowID := "payout-flow-wallet-to-bank" + merchantID.String()
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	err := workflow.ExecuteActivity(
		ctx,
		"ValidatePayoutWalletToBankRequest",
		request,
		merchantID,
	).Get(ctx, nil)

	if err != nil {
		return nil, err
	}

	var fees decimal.Decimal

	errCF := workflow.ExecuteActivity(
		ctx,
		"CalculateFees",
		request.Amount,
	).Get(ctx, &fees)

	if errCF != nil {
		return nil, errCF
	}

	var payout *models.Payout
	errCP := workflow.ExecuteActivity(
		ctx,
		"CreatePayoutWalletToBank",
		request,
		merchantID,
		constants.TransactionStatusPending,
	).Get(ctx, &payout)

	if errCP != nil {
		return nil, errCP
	}

	executePaymentPayload := &activities.PayoutWorkflowWalletToBanRequest{
		PayoutReference: payout.PayoutReference,
		MerchantID:      merchantID,
		Fee:             fees,
		Request:         *request,
	}
	var response *activities.ExecutePayoutWalletToVault
	errEP := workflow.ExecuteActivity(
		ctx,
		"ExecutePayoutWalletToVault",
		executePaymentPayload,
	).Get(ctx, &response)

	walletToBankTransferEventPaylaod := &dto.CreateWalletToBankCreditEvent{
		EventID:   uuid.New(),
		EventType: string(constants.EventTypeSampayToBankCreditRequested),
		Payload: &dto.CreateWalletToBankEventRequest{
			ReferenecID:   payout.ID,
			WalletID:      *payout.SourceWalletID,
			BankAccountID: payout.DestinationBankAccountID,
			Amount:        payout.Amount,
			Currency:      payout.Currency,
		},
	}
	var event *models.Event
	err = workflow.ExecuteActivity(
		ctx,
		"PublishWalletToBankTransferEvent",
		walletToBankTransferEventPaylaod,
		workflowID,
	).Get(ctx, &event)
	if err != nil {
		return nil, err
	}

	signalCh := workflow.GetSignalChannel(
		ctx,
		WalletToBankTransferSignal,
	)

	timer := workflow.NewTimer(ctx, 10*time.Minute)
	var walletToBankResponse *dto.WalletToBankTransferCompletedPayload
	var timedOut bool
	var receivedMatchingResponse bool

	for !receivedMatchingResponse && !timedOut {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(
			signalCh,
			func(c workflow.ReceiveChannel, more bool) {
				var response dto.WalletToBankTransferCompletedPayload
				c.Receive(ctx, &response)

				if response.Payload.RequestedEventID == event.EventID {
					walletToBankResponse = &response
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

	var finalPayout *models.Payout
	err = workflow.ExecuteActivity(
		ctx,
		"ExecuteVaultToBankAccount",
		walletToBankResponse.Event,
		request.DestinationBankAccountID,
		response.LedgerTransaction,
		response.Payout,
		fees,
	).Get(ctx, &finalPayout)

	if err == nil{
		return finalPayout, nil
	}

	failedStatus := constants.TransactionStatusFailed

	var finalUpdatedPayout *models.Payout
	failErr := workflow.ExecuteActivity(
		ctx,
		"UpdatePayoutStatus",
		&failedStatus,
		payout.PayoutReference,
	).Get(ctx, &finalUpdatedPayout)

	if failErr != nil {
		return nil, fmt.Errorf(
			"payment failed: %w; failed to mark payment as failed: %v",
			errEP,
			failErr,
		)
	}

	var finalEvent *models.Event
	failErr = workflow.ExecuteActivity(
		ctx,
		"UpdateEventStatus",
		walletToBankResponse.Event.EventID,
		constants.EventStatusFailed,
	).Get(ctx, &finalEvent)

	if failErr != nil {
		return nil, fmt.Errorf(
			"event failed: %w; failed to mark event as failed: %v",
			errEP,
			failErr,
		)
	}

	return finalUpdatedPayout, nil
}

func PayoutFlowBankToBank(
	ctx workflow.Context,
	request *dto.CreateBankToBankRequest,
	merchantID uuid.UUID,
) (*models.Payout, error) {
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

	err := workflow.ExecuteActivity(
		ctx,
		"ValidatePayoutBankToBankRequest",
		request,
		merchantID,
	).Get(ctx, nil)

	if err != nil {
		return nil, err
	}

	var payout *models.Payout
	errCP := workflow.ExecuteActivity(
		ctx,
		"CreatePayoutBankToBank",
		request,
		merchantID,
		constants.TransactionStatusPending,
	).Get(ctx, &payout)

	if errCP != nil {
		return nil, errCP
	}

	executePaymentPayload := &activities.PayoutWorkflowBankToBanRequest{
		PayoutReference: payout.PayoutReference,
		MerchantID:      merchantID,
		Fee:             decimal.Zero,
		Request:         *request,
	}
	var updatedPayout *models.Payout
	errEP := workflow.ExecuteActivity(
		ctx,
		"ExecutePayoutBankToBank",
		executePaymentPayload,
	).Get(ctx, &updatedPayout)

	if errEP == nil {
		return updatedPayout, nil
	}

	failedStatus := constants.TransactionStatusFailed

	var finalUpdatedPayment *models.Payout
	failErr := workflow.ExecuteActivity(
		ctx,
		"UpdatePayoutStatus",
		&failedStatus,
		payout.PayoutReference,
	).Get(ctx, &finalUpdatedPayment)

	if failErr != nil {
		return nil, fmt.Errorf(
			"payment failed: %w; failed to mark payment as failed: %v",
			errEP,
			failErr,
		)
	}

	return finalUpdatedPayment, nil
}
