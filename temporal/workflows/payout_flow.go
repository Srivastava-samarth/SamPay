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

func PayoutFlow(
	ctx workflow.Context,
	request *dto.CreateWalletToBankRequest,
	merchantID uuid.UUID,
) (*models.Payout ,error) {
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
		MerchantID: merchantID,
		Fee: fees,
		Request: *request,
	}
	var updatedPayout *models.Payout
	errEP := workflow.ExecuteActivity(
		ctx,
		"ExecutePayoutWalletToBank",
		executePaymentPayload,
	).Get(ctx, &updatedPayout)

	if errEP == nil{
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

func PayoutFlowBankToBank(
	ctx workflow.Context,
	request *dto.CreateBankToBankRequest,
	merchantID uuid.UUID,
) (*models.Payout ,error) {
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
		MerchantID: merchantID,
		Fee: decimal.Zero,
		Request: *request,
	}
	var updatedPayout *models.Payout
	errEP := workflow.ExecuteActivity(
		ctx,
		"ExecutePayoutBankToBank",
		executePaymentPayload,
	).Get(ctx, &updatedPayout)

	if errEP == nil{
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
