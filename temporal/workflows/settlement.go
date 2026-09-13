package workflows

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func SettlementFlow(ctx workflow.Context) error {
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
	var payments []*models.Payment

	err := workflow.ExecuteActivity(
		ctx,
		"GetUnsettledPayment",
	).Get(ctx, &payments)

	if err != nil {
		return err
	}

	paymentsByMerchant := make(map[uuid.UUID][]*models.Payment)

	for _, payment := range payments {
		if payment == nil {
			continue
		}

		paymentsByMerchant[payment.ReceiverMerchantID] =
			append(
				paymentsByMerchant[payment.ReceiverMerchantID],
				payment,
			)
	}

	for merchantID, merchantPayments := range paymentsByMerchant {

		err := workflow.ExecuteActivity(
			ctx,
			"ExecuteSettlementByMerchant",
			merchantID,
			merchantPayments,
		).Get(ctx, nil)

		if err != nil {
			return err
		}
	}

	return nil
}