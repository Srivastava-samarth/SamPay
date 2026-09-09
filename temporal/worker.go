package temporal

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/Srivastava-samarth/sampay/temporal/activities"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
)

const (
	MerchantOnboardingTaskQueue = "MERCHANT_ONBOARDING"
	ForgotPasswordTaskQueue     = "FORGOT_PASSWORD"
	UserOnboardingTaskQueue     = "USER_ONBOARDING"
	PaymentFlowTaskQueue        = "PAYMENT_FLOW"
	PayoutFlowTaskQueue         = "PAYOUT_FLOW"
	RefundFlowTaskQueue         = "REFUND_FLOW"
)

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
		{
			TaskQueue: PayoutFlowTaskQueue,
			Register: func(w worker.Worker) {
				w.RegisterWorkflow(workflows.PayoutFlow)
				w.RegisterActivity(activityRegistry.CreatePayoutWalletToBank)
				w.RegisterActivity(activityRegistry.CreatePayoutBankToBank)
				w.RegisterActivity(activityRegistry.ExecutePayoutWalletToBank)
				w.RegisterActivity(activityRegistry.ExecutePayoutBankToBank)
				w.RegisterActivity(activityRegistry.UpdatePayoutStatus)
				w.RegisterActivity(activityRegistry.CalculateFees)
				w.RegisterActivity(activityRegistry.ValidatePayoutWalletToBankRequest)
				w.RegisterActivity(activityRegistry.ValidatePayoutBankToBankRequest)
			},
		},
		{
			TaskQueue: PayoutFlowTaskQueue,
			Register: func(w worker.Worker) {
				w.RegisterWorkflow(workflows.PayoutFlowBankToBank)
				w.RegisterActivity(activityRegistry.CreatePayoutWalletToBank)
				w.RegisterActivity(activityRegistry.CreatePayoutBankToBank)
				w.RegisterActivity(activityRegistry.ExecutePayoutWalletToBank)
				w.RegisterActivity(activityRegistry.ExecutePayoutBankToBank)
				w.RegisterActivity(activityRegistry.UpdatePayoutStatus)

				w.RegisterActivity(activityRegistry.CalculateFees)
				w.RegisterActivity(activityRegistry.ValidatePayoutWalletToBankRequest)
				w.RegisterActivity(activityRegistry.ValidatePayoutBankToBankRequest)
			},
		},
		{
			TaskQueue: RefundFlowTaskQueue,
			Register: func(w worker.Worker) {
				w.RegisterWorkflow(workflows.RefundFlow)
				w.RegisterActivity(activityRegistry.ValidateRefundRequest)
				w.RegisterActivity(activityRegistry.UpdateRefundStatus)
				w.RegisterActivity(activityRegistry.GetRefundByPaymentID)
				w.RegisterActivity(activityRegistry.GetPaymentByID)
				w.RegisterActivity(activityRegistry.CreateRefund)
				w.RegisterActivity(activityRegistry.RefundFromMerchantWallet)
				w.RegisterActivity(activityRegistry.RefundFromPaymentVault)
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
