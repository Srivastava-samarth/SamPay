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

const BankAccountCreationResultSignal = "BankAccountCreationResult"

func MerchantONboardingWorkflow(
	ctx workflow.Context,
	request *dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
) error {
	workflowID := "merchant-onboarding-" + merchantID.String()
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

	// 1. Compliance check.
	var complianceResponse dto.ComplianceCheckResponse

	err := workflow.ExecuteActivity(
		ctx,
		"PerformComplianceCheck",
		request,
		merchantID,
	).Get(ctx, &complianceResponse)
	if err != nil {
		return err
	}

	// 2. Persist compliance result.
	err = workflow.ExecuteActivity(
		ctx,
		"UpdateMerchantCompliance",
		merchantID,
		&complianceResponse,
	).Get(ctx, nil)
	if err != nil {
		return err
	}

	// 3. Stop if compliance is rejected.
	if complianceResponse.ComplianceStatus ==
		constants.ComplianceStatusRejected {
		return workflow.ExecuteActivity(
			ctx,
			"SendKYCReattemptEmail",
			request.Email,
			merchantID,
			request.MerchantType,
		).Get(ctx, nil)
	}

	// Register the signal channel before requesting account creation.
	signalCh := workflow.GetSignalChannel(
		ctx,
		BankAccountCreationResultSignal,
	)

	// 4. Create wallet in its initial inactive state.
	err = workflow.ExecuteActivity(
		ctx,
		"CreateWallet",
		merchantID,
	).Get(ctx, nil)
	if err != nil {
		return err
	}
	// 5. Persist and publish the bank-account creation request.
	// This Activity must return the original request EventID.
	bankAccoutnCreationEventPayload := &dto.BankAccountCreationRequestEvent{
		EventID: uuid.New(),
		EventType: string(constants.EventTypeSampayToBankAccountCreateRequested),
		Payload: &dto.CreateBankAccountRequest{
			AccountName: request.MerchantName,
			AccountType: constants.BankAccountTypePrimary,
		},
	}

	var event *models.Event
	err = workflow.ExecuteActivity(
		ctx,
		"PublishBankAccountCreationRequest",
		bankAccoutnCreationEventPayload,
		workflowID,
	).Get(ctx, &event)
	if err != nil {
		return err
	}

	// 6. Wait for the NATS consumer to signal the bank response.

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
		return fmt.Errorf(
			"timed out waiting for bank account creation: %s",
			event.EventID,
		)
	}

	// 7. Decode the raw bank account model returned by sampay-bank.
	var bankAccount *models.BankAccount
	if err := json.Unmarshal(bankResponse.Payload.Payload, &bankAccount); err != nil {
		return fmt.Errorf("unmarshal bank account model: %w", err)
	}

	// 8. Create the SamPay merchant-bank account association.
	err = workflow.ExecuteActivity(
		ctx,
		"CreateLinkedBankAccount",
		bankAccount,
		merchantID,
	).Get(ctx, nil)
	if err != nil {
		return err
	}

	// 9. Create the owner user and merchant-user association.
	// Both DB writes should share one transaction inside this Activity.
	var provisionedUser dto.MerchantProvisioningResult

	err = workflow.ExecuteActivity(
		ctx,
		"ProvisionUser",
		merchantID,
		request,
	).Get(ctx, &provisionedUser)
	if err != nil {
		return err
	}

	// 10. Activate the merchant, wallet and owner user.
	err = workflow.ExecuteActivity(
		ctx,
		"ActivateMerchantOnboarding",
		merchantID,
		provisionedUser.User.ID,
	).Get(ctx, nil)
	if err != nil {
		return err
	}

	// 11. Send the welcome email only after activation.
	return workflow.ExecuteActivity(
		ctx,
		"SendWelcomeEmail",
		provisionedUser.User,
		provisionedUser.TemporaryPassword,
	).Get(ctx, nil)
}