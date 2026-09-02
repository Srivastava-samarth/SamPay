package workflows

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/services"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func UserOnboardingFlow(
	ctx workflow.Context,
	request dto.UserOnboardingRequest,
) (*dto.CreateUserResponse,error) {

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

	var merchant *models.Merchant
	errM := workflow.ExecuteActivity(
		ctx,
		"GetMerchantById",
		request.MerchantID,
	).Get(ctx, &merchant)

	if errM != nil {
		return nil, errM
	}

	var user *services.CreatedUserResult
	createUserPayload := &dto.CreateUserRequest{
		Email: request.Email,
		FirstName: request.FirstName,
		LastName: request.LastName,
	} 

	errCU := workflow.ExecuteActivity(
		ctx,
		"CreateUser",
		createUserPayload,
	).Get(ctx, &user)

	if errCU != nil {
		return nil, errCU
	}

	var merchantUser dto.CreateMerchantUserResponse
	createMerchantUserPayload := &dto.CreateMerchantUserRequest{
		MerchantID: request.MerchantID,
		UserID: user.User.ID,
		Role: request.Role,
	}

	errMU := workflow.ExecuteActivity(
		ctx,
		"CreateMerchantUser",
		createMerchantUserPayload,
	).Get(ctx, &merchantUser)

	if errMU != nil {
		return nil, errMU
	}

	// 5. Send welcome email
	errN := workflow.ExecuteActivity(
		ctx,
		"SendUserWelcomeEmail",
		user.User.FirstName,
		user.User.Email,
		merchant.MerchantName,
		merchantUser.Role,
		user.TemporaryPassword,
	).Get(ctx, nil)

	if errN != nil {
		return nil, errN
	}

	return user.User, nil
}