package workflows

import (
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/services"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func MerchantUpdateKycFlow(
	ctx workflow.Context,
	request dto.UpdateKYCRequest,
	merchantID uuid.UUID,
) (*models.Merchant, error) {

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

	complianceRequest := &dto.CreateMerchantOnboardingRequest{
		Individual: request.Individual,
		Company:    request.Company,
	}
	err := workflow.ExecuteActivity(
		ctx,
		"PerformComplianceCheck",
		complianceRequest,
		merchantID,
	).Get(ctx, &complianceResponse)

	if err != nil {
		return nil, err
	}

	// 2. Save compliance result
	err = workflow.ExecuteActivity(
		ctx,
		"UpdateMerchantCompliance",
		merchantID,
		complianceResponse,
	).Get(ctx, nil)

	if err != nil {
		return nil, err
	}

	var merchant *models.Merchant
	errM := workflow.ExecuteActivity(
		ctx,
		"GetMerchantById",
		merchantID,
	).Get(ctx, &merchant)

	if errM != nil {
		return nil, err
	}

	// 3. Stop if rejected
	if complianceResponse.ComplianceStatus ==
		constants.ComplianceStatusRejected {
		// Send KYC re-attempt mail
		err = workflow.ExecuteActivity(
			ctx,
			"SendKYCReattemptEmail",
			merchant.Email,
			merchantID,
			merchant.MerchantType,
		).Get(ctx, nil)

		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	// 4. Provision everything
	var provisionedUser services.MerchantProvisioningResult
	merchantOnboardingRequest := &dto.CreateMerchantOnboardingRequest{
		MerchantType: merchant.MerchantType,
		MerchantName: merchant.MerchantName,
		Email:        merchant.Email,
		PhoneNumber:  merchant.PhoneNumber,
		Individual:   request.Individual,
		Company:      request.Company,
	}

	err = workflow.ExecuteActivity(
		ctx,
		"ProvisionMerchant",
		merchantOnboardingRequest,
		merchantID,
	).Get(ctx, &provisionedUser)

	if err != nil {
		return nil, err
	}

	// 5. Send welcome email
	err = workflow.ExecuteActivity(
		ctx,
		"SendWelcomeEmail",
		provisionedUser.User,
		provisionedUser.TemporaryPassword,
	).Get(ctx, nil)

	if err != nil {
		return nil, err
	}

	// 6. Get updated merchant
	var updatedMerchant *models.Merchant
	errUM := workflow.ExecuteActivity(
		ctx,
		"GetMerchantById",
		merchantID,
	).Get(ctx, &updatedMerchant)

	if errUM != nil {
		return nil, errUM
	}

	return updatedMerchant, nil
}
