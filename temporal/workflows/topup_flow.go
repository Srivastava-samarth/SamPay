package workflows

import (
	"fmt"
	"log"
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

func TopUpFlow(
	ctx workflow.Context,
	request *dto.TopupWalletBalanceRequest,
	merchantID uuid.UUID,
) (*models.Wallets, error) {
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
			ReferenecID: payout.ID,
			SenderAccountID: request.SourceBankAccountID,
			ReceiverAccountID: request.DestinationBankAccountID,
			Amount: request.Amount,
			Currency: request.Currency,
		},
	}

	var event *models.Event
	err = workflow.ExecuteActivity(
		ctx,
		"PublishBankToBankTransferEvent",
		bankToBankTransferEventPayload,
		workflowID,
	).Get(ctx, &event)
	if err != nil {
		return nil, err
	}

	var updatedWallet *models.Wallets
	errW := workflow.ExecuteActivity(
		ctx,
		"ExecuteBankToWallet",
		event,
		merchantID,
		request,
	).Get(ctx, &updatedWallet)

	if errW == nil{
		return updatedWallet, nil 
	}

	var updated
}
