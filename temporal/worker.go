package temporal

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/Srivastava-samarth/sampay/temporal/activities"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
)

const MerchantOnboardingTaskQueue = "MERCHANT_ONBOARDING"
const ForgotPasswordTaskQueue = "FORGOT_PASSWORD"
const UserOnboardingTaskQueue = "USER_ONBOARDING"
const PaymentFlowTaskQueue = "PAYMENT_FLOW"

type WorkerConfig struct {
    TaskQueue string
    Register  func(worker.Worker)
}

func StartWorkers(
    temporalClient client.Client,
    activityRegistry *activities.Registry,
) []worker.Worker {

    configs := []WorkerConfig{
        {
            TaskQueue: MerchantOnboardingTaskQueue,
            Register: func(w worker.Worker) {
                w.RegisterWorkflow(workflows.MerchantONboardingWorkflow)

                w.RegisterActivity(activityRegistry.PerformComplianceCheck)
                w.RegisterActivity(activityRegistry.ProvisionMerchant)
                w.RegisterActivity(activityRegistry.SendWelcomeEmail)
                w.RegisterActivity(activityRegistry.UpdateMerchantCompliance)
            },
        },
        {
            TaskQueue: ForgotPasswordTaskQueue,
            Register: func(w worker.Worker) {
                w.RegisterWorkflow(workflows.ForgotPasswordWorkflow)

                w.RegisterActivity(activityRegistry.ForgotPassword)
            },
        },
        {
            TaskQueue: UserOnboardingTaskQueue,
            Register: func(w worker.Worker) {
                w.RegisterWorkflow(workflows.UserOnboardingFlow)
                w.RegisterActivity(activityRegistry.SendUserWelcomeEmail)
                w.RegisterActivity(activityRegistry.GetMerchantById)
                w.RegisterActivity(activityRegistry.CreateMerchantUser)
                w.RegisterActivity(activityRegistry.CreateUser)
            },
        },
         {
            TaskQueue: PaymentFlowTaskQueue,
            Register: func(w worker.Worker) {
                w.RegisterWorkflow(workflows.PaymentFlow)
                w.RegisterActivity(activityRegistry.CreatePayment)
                w.RegisterActivity(activityRegistry.CalculateFees)
                w.RegisterActivity(activityRegistry.ExecutePayment)
                w.RegisterActivity(activityRegistry.UpdatePaymentStatus)
                w.RegisterActivity(activityRegistry.ValidatePaymentRequest)
            },
        },
    }

    workers := make([]worker.Worker, 0, len(configs))

    for _, cfg := range configs {
        w := worker.New(
            temporalClient,
            cfg.TaskQueue,
            worker.Options{},
        )

        cfg.Register(w)
        workers = append(workers, w)
    }

    return workers
}