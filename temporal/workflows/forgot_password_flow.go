package workflows

import (
	"time"

	"github.com/Srivastava-samarth/sampay/dto"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func ForgotPasswordWorkflow(
	ctx workflow.Context,
	request *dto.ForgotPasswordRequest,
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

	err := workflow.ExecuteActivity(
		ctx,
		"ForgotPassword",
		request,
	).Get(ctx, nil)

	if err != nil {
		return err
	}

	return nil
}