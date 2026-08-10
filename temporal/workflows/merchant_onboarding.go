package workflows

import (
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/services"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func MerchantONboardingWorkflow(
	ctx workflow.Context,
	request dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
) error {

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

	var complianceResponse dto.ComplianceCheckResponse

	// 1. Compliance check
	err := workflow.ExecuteActivity(
		ctx,
		"PerformComplianceCheck",
		request,
		merchantID,
	).Get(ctx, &complianceResponse)

	if err != nil{
		return err;
	}

	// 2. Save compliance result
	err = workflow.ExecuteActivity(
		ctx,
		"UpdateMerchantCompliance",
		merchantID,
		complianceResponse,
	).Get(ctx, nil)

	if err != nil {
		return err
	}

	// 3. Stop if rejected
	if complianceResponse.ComplianceStatus ==
		constants.ComplianceStatusRejected {
		return nil
	}

	// 4. Provision everything
	var provisionedUser services.MerchantProvisioningResult

	err = workflow.ExecuteActivity(
		ctx,
		"ProvisionMerchant",
		request,
		merchantID,
	).Get(ctx, &provisionedUser)

	if err != nil {
		return err
	}

	// 5. Send welcome email
	err = workflow.ExecuteActivity(
		ctx,
		"SendWelcomeEmail",
		provisionedUser.User,
		provisionedUser.TemporaryPassword,
	).Get(ctx, nil)

	if err != nil {
		return err
	}

	return nil
}
