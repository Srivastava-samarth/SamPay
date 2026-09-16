package controllers

import (
	"os"
	"testing"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/Srivastava-samarth/sampay/notifications"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/services"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/activities"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/Srivastava-samarth/sampay/testutils"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"
)

var (
	db             *gorm.DB
	testRepo       *repositories.Repository
	testServices   *services.Services
	testController *Controller
	temporalClient client.Client
)

func TestMain(m *testing.M) {
	var err error

	db, err = testutils.SetupTestDB()
	if err != nil {
		panic(err)
	}

	temporalConfig := &config.TemporalConfig{
		Host: "localhost:7233",
	}

	temporalClient, err = temporal.NewClient(temporalConfig)
	if err != nil {
		panic(err)
	}

	testRepo = repositories.NewRepository(db)

	jwtService := middlewares.NewJwt(&config.JWTConfig{
		AccessExpiry:  "3600",
		RefreshExpiry: "86400",
	})

	notificationService, err := notifications.NewEmailService(
		config.SMTPConfig{
			Host: "localhost",
			Port: "1025",
			From: "no-reply@sampay.io",
		},
	)
	if err != nil {
		panic(err)
	}

	testServices = services.NewServices(
		db,
		testRepo,
		jwtService,
		notificationService,
	)

	activityRegistry := activities.NewRegistry(
		db,
		notificationService,
		testServices,
		testRepo,
	)

	// Forgot Password Worker
	forgotPasswordWorker := worker.New(
		temporalClient,
		temporal.ForgotPasswordTaskQueue,
		worker.Options{},
	)

	forgotPasswordWorker.RegisterWorkflow(
		workflows.ForgotPasswordWorkflow,
	)
	forgotPasswordWorker.RegisterActivity(
		activityRegistry.ForgotPassword,
	)

	if err := forgotPasswordWorker.Start(); err != nil {
		panic(err)
	}
	defer forgotPasswordWorker.Stop()

	// Payment Flow Worker
	paymentWorker := worker.New(
		temporalClient,
		temporal.PaymentFlowTaskQueue,
		worker.Options{},
	)

	paymentWorker.RegisterWorkflow(
		workflows.PaymentFlow,
	)
	paymentWorker.RegisterActivity(
		activityRegistry.CreatePayment,
	)
	paymentWorker.RegisterActivity(
		activityRegistry.CalculateFees,
	)
	paymentWorker.RegisterActivity(
		activityRegistry.ExecutePayment,
	)
	paymentWorker.RegisterActivity(
		activityRegistry.UpdatePaymentStatus,
	)
	paymentWorker.RegisterActivity(
		activityRegistry.ValidatePaymentRequest,
	)

	if err := paymentWorker.Start(); err != nil {
		panic(err)
	}
	defer paymentWorker.Stop()

	settlementWorker := worker.New(
		temporalClient,
		temporal.SettlementFlowTaskQueue,
		worker.Options{},
	)

	settlementWorker.RegisterWorkflow(workflows.SettlementFlow)
	settlementWorker.RegisterActivity(activityRegistry.GetUnsettledPayment)
	settlementWorker.RegisterActivity(activityRegistry.ExecuteSettlementByMerchant)

	if err := settlementWorker.Start(); err != nil {
		panic(err)
	}
	defer settlementWorker.Stop()

	payoutWorker := worker.New(
		temporalClient,
		temporal.PayoutFlowTaskQueue,
		worker.Options{},
	)

	payoutWorker.RegisterWorkflow(workflows.PayoutFlow)
	payoutWorker.RegisterWorkflow(workflows.PayoutFlowBankToBank)

	payoutWorker.RegisterActivity(activityRegistry.CreatePayoutWalletToBank)
	payoutWorker.RegisterActivity(activityRegistry.CreatePayoutBankToBank)
	payoutWorker.RegisterActivity(activityRegistry.ExecutePayoutWalletToBank)
	payoutWorker.RegisterActivity(activityRegistry.ExecutePayoutBankToBank)
	payoutWorker.RegisterActivity(activityRegistry.UpdatePayoutStatus)
	payoutWorker.RegisterActivity(activityRegistry.CalculateFees)
	payoutWorker.RegisterActivity(activityRegistry.ValidatePayoutWalletToBankRequest)
	payoutWorker.RegisterActivity(activityRegistry.ValidatePayoutBankToBankRequest)

	if err := payoutWorker.Start(); err != nil {
		panic(err)
	}
	defer payoutWorker.Stop()

	userOnboardingWorker := worker.New(
		temporalClient,
		temporal.UserOnboardingTaskQueue,
		worker.Options{},
	)

	userOnboardingWorker.RegisterWorkflow(workflows.UserOnboardingFlow)
	userOnboardingWorker.RegisterActivity(activityRegistry.SendUserWelcomeEmail)
	userOnboardingWorker.RegisterActivity(activityRegistry.GetMerchantById)
	userOnboardingWorker.RegisterActivity(activityRegistry.CreateMerchantUser)
	userOnboardingWorker.RegisterActivity(activityRegistry.CreateUser)

	if err := userOnboardingWorker.Start(); err != nil {
		panic(err)
	}
	defer userOnboardingWorker.Stop()

	refundWorker := worker.New(
		temporalClient,
		temporal.RefundFlowTaskQueue,
		worker.Options{},
	)

	refundWorker.RegisterWorkflow(workflows.RefundFlow)
	refundWorker.RegisterActivity(activityRegistry.ValidateRefundRequest)
	refundWorker.RegisterActivity(activityRegistry.UpdateRefundStatus)
	refundWorker.RegisterActivity(activityRegistry.GetRefundByPaymentID)
	refundWorker.RegisterActivity(activityRegistry.GetPaymentByID)
	refundWorker.RegisterActivity(activityRegistry.CreateRefund)
	refundWorker.RegisterActivity(activityRegistry.RefundFromMerchantWallet)
	refundWorker.RegisterActivity(activityRegistry.RefundFromPaymentVault)

	if err := refundWorker.Start(); err != nil {
		panic(err)
	}
	defer refundWorker.Stop()

	reconWorker := worker.New(
		temporalClient,
		temporal.ReconFlowTaskQueue,
		worker.Options{},
	)

	reconWorker.RegisterWorkflow(workflows.ReconFlow)
	reconWorker.RegisterActivity(activityRegistry.GetReconTimeRange)
	reconWorker.RegisterActivity(activityRegistry.GetLedgerTransactionsForRecon)
	reconWorker.RegisterActivity(activityRegistry.GetLedgerEntriesForRecon)
	reconWorker.RegisterActivity(activityRegistry.ReconcileLedgerTransactions)
	reconWorker.RegisterActivity(activityRegistry.GenerateReconReport)
	reconWorker.RegisterActivity(activityRegistry.GenerateReportEmail)
	reconWorker.RegisterActivity(activityRegistry.SendReconEmail)

	if err := reconWorker.Start(); err != nil {
		panic(err)
	}
	defer reconWorker.Stop()

	updateKYCWorker := worker.New(
		temporalClient,
		temporal.UpdateKycFlow,
		worker.Options{},
	)

	updateKYCWorker.RegisterWorkflow(workflows.MerchantUpdateKycFlow)
	updateKYCWorker.RegisterActivity(activityRegistry.GetMerchantById)
	updateKYCWorker.RegisterActivity(activityRegistry.ProvisionMerchant)
	updateKYCWorker.RegisterActivity(activityRegistry.PerformComplianceCheck)
	updateKYCWorker.RegisterActivity(activityRegistry.SendKYCReattemptEmail)
	updateKYCWorker.RegisterActivity(activityRegistry.SendWelcomeEmail)
	updateKYCWorker.RegisterActivity(activityRegistry.UpdateMerchantCompliance)

	if err := updateKYCWorker.Start(); err != nil {
		panic(err)
	}
	defer reconWorker.Stop()

	merchantOnboardingWorker := worker.New(
		temporalClient,
		temporal.ForgotPasswordTaskQueue,
		worker.Options{},
	)

	merchantOnboardingWorker.RegisterWorkflow(workflows.MerchantONboardingWorkflow)

	merchantOnboardingWorker.RegisterActivity(activityRegistry.PerformComplianceCheck)
	merchantOnboardingWorker.RegisterActivity(activityRegistry.ProvisionMerchant)
	merchantOnboardingWorker.RegisterActivity(activityRegistry.SendWelcomeEmail)
	merchantOnboardingWorker.RegisterActivity(activityRegistry.SendKYCReattemptEmail)
	merchantOnboardingWorker.RegisterActivity(activityRegistry.UpdateMerchantCompliance)

	if err := merchantOnboardingWorker.Start(); err != nil {
		panic(err)
	}
	defer merchantOnboardingWorker.Stop()

	testController = NewController(
		db,
		temporalClient,
		testServices,
	)

	code := m.Run()

	if err := testutils.CleanupTestDB(db); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func cleanTestDB(t *testing.T) {
	t.Helper()

	if err := testutils.CleanupTestDB(db); err != nil {
		t.Fatalf("failed to clean test DB: %v", err)
	}
}
