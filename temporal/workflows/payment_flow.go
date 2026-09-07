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

func PaymentFlow(
	ctx workflow.Context,
	request *dto.CreatePaymentRequest,
	merchantID uuid.UUID,
) (*models.Payment ,error) {
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
		"ValidatePaymentRequest",
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

	var payment *models.Payment
	errCP := workflow.ExecuteActivity(
		ctx,
		"CreatePayment",
		request,
		merchantID,
		constants.TransactionStatusPending,
	).Get(ctx, &payment)

	if errCP != nil {
		return nil, errCP
	}

	executePaymentPayload := &activities.PaymentWorkflowRequest{
		PaymentReference: *payment.PaymentReference,
		MerchantID: merchantID,
		Fee: fees,
		Request: *request,
	}
	var updatedPayment *models.Payment
	errEP := workflow.ExecuteActivity(
		ctx,
		"ExecutePayment",
		executePaymentPayload,
	).Get(ctx, &updatedPayment)

	if errEP == nil{
		return updatedPayment, nil
	}

	failedStatus := constants.TransactionStatusFailed

	var finalUpdatedPayment *models.Payment
	failErr := workflow.ExecuteActivity(
		ctx,
		"UpdatePaymentStatus",
		&failedStatus,
		payment.PaymentReference,
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
