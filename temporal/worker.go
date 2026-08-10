package temporal

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/Srivastava-samarth/sampay/temporal/activities"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
)

const MerchantOnboardingTaskQueue = "MERCHANT_ONBOARDING"

func StartWorker(
	temporalClient client.Client,
	activityRegistry *activities.Registry,
) worker.Worker {

	w := worker.New(
		temporalClient,
		MerchantOnboardingTaskQueue,
		worker.Options{},
	)

	w.RegisterWorkflow(workflows.MerchantONboardingWorkflow)

	w.RegisterActivity(activityRegistry.PerformComplianceCheck)
	w.RegisterActivity(activityRegistry.ProvisionMerchant)
	w.RegisterActivity(activityRegistry.SendWelcomeEmail)
	w.RegisterActivity(activityRegistry.UpdateMerchantCompliance)
	return w
}