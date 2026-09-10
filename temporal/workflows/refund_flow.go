package workflows

import (
	"errors"
	"fmt"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func RefundFlow(
	ctx workflow.Context,
	request *dto.CreateRefundRequest,
	merchantID *uuid.UUID,
) (*models.Refund, error) {

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

	// 1. Validate refund request
	err := workflow.ExecuteActivity(
		ctx,
		"ValidateRefundRequest",
		request,
		merchantID,
	).Get(ctx, nil)

	if err != nil {
		return nil, err
	}

	// 2. Get previous refunds
	var refunds []*models.Refund

	err = workflow.ExecuteActivity(
		ctx,
		"GetRefundsByPaymentID",
		request.PaymentID,
	).Get(ctx, &refunds)

	if err != nil {
		return nil, err
	}

	// 3. Create refund
	var refund *models.Refund

	err = workflow.ExecuteActivity(
		ctx,
		"CreateRefund",
		request,
		refunds,
	).Get(ctx, &refund)

	if err != nil {
		return nil, err
	}

	// 4. Get payment
	var payment *models.Payment

	err = workflow.ExecuteActivity(
		ctx,
		"GetPaymentByID",
		refund.PaymentID,
	).Get(ctx, &payment)

	if err != nil {
		var updatedRefund *models.Refund

		statusErr := workflow.ExecuteActivity(
			ctx,
			"UpdateRefundStatus",
			refund.ID,
			constants.TransactionStatusFailed,
		).Get(ctx, &updatedRefund)

		if statusErr != nil {
			return nil, fmt.Errorf(
				"failed to get payment: %w; failed to mark refund as failed: %v",
				err,
				statusErr,
			)
		}

		return updatedRefund, err
	}

	if payment == nil {
		var updatedRefund *models.Refund

		statusErr := workflow.ExecuteActivity(
			ctx,
			"UpdateRefundStatus",
			refund.ID,
			constants.TransactionStatusFailed,
		).Get(ctx, &updatedRefund)

		if statusErr != nil {
			return nil, fmt.Errorf(
				"payment not found: %w; failed to mark refund as failed: %v",
				errors.New("payment not found"),
				statusErr,
			)
		}

		return updatedRefund, errors.New("payment not found")
	}

	if payment.SettlementStatus == nil {
		var updatedRefund *models.Refund

		statusErr := workflow.ExecuteActivity(
			ctx,
			"UpdateRefundStatus",
			refund.ID,
			constants.TransactionStatusFailed,
		).Get(ctx, &updatedRefund)

		if statusErr != nil {
			return nil, fmt.Errorf(
				"payment settlement status not found: %w; failed to mark refund as failed: %v",
				errors.New("payment settlement status not found"),
				statusErr,
			)
		}

		return updatedRefund, errors.New("payment settlement status not found")
	}

	var updatedRefund *models.Refund

	if *payment.SettlementStatus == constants.LedgerSettlementSettled {

		err = workflow.ExecuteActivity(
			ctx,
			"RefundFromMerchantWallet",
			refund,
			payment,
		).Get(ctx, &updatedRefund)

	} else {

		err = workflow.ExecuteActivity(
			ctx,
			"RefundFromPaymentVault",
			refund,
			payment,
		).Get(ctx, &updatedRefund)
	}

	if err != nil {
		return updatedRefund, err
	}

	return updatedRefund, nil
}
