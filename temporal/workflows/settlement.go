package workflows

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

func SettlementFlow(ctx workflow.Context) error {
	var payments []*models.Payment

	err := workflow.ExecuteActivity(
		ctx,
		"GetUnsettledPayments",
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