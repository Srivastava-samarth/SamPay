package workflows

import (
	"fmt"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/temporal/activities"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func ReconFlow(ctx workflow.Context) error {

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

	var timeRange activities.ReconTimeRange

	err := workflow.ExecuteActivity(
		ctx,
		"GetReconTimeRange",
	).Get(ctx, &timeRange)

	if err != nil {
		return fmt.Errorf("failed to get recon time range: %w", err)
	}

	var transactions []*models.LedgerTransaction

	err = workflow.ExecuteActivity(
		ctx,
		"GetLedgerTransactionsForRecon",
		&timeRange,
	).Get(ctx, &transactions)

	if err != nil {
		return fmt.Errorf("failed to get ledger transactions: %w", err)
	}

	transactionIDs := make([]uuid.UUID, 0, len(transactions))

	for _, transaction := range transactions {
		if transaction == nil {
			continue
		}

		transactionIDs = append(transactionIDs, transaction.ID)
	}

	var entries []*models.LedgerEntry

	err = workflow.ExecuteActivity(
		ctx,
		"GetLedgerEntriesForRecon",
		transactionIDs,
	).Get(ctx, &entries)

	if err != nil {
		return fmt.Errorf("failed to get ledger entries: %w", err)
	}

	var reconResult dto.ReconResult

	err = workflow.ExecuteActivity(
		ctx,
		"ReconcileLedgerTransactions",
		transactions,
		entries,
	).Get(ctx, &reconResult)

	if err != nil {
		return fmt.Errorf("failed to reconcile transactions: %w", err)
	}

	var reconReport *dto.ReconReport

	err = workflow.ExecuteActivity(
		ctx,
		"GenerateReconReport",
		&reconResult,
	).Get(ctx, &reconReport)

	if err != nil {
		return fmt.Errorf("failed to generate recon report: %w", err)
	}

	var emailBody string

	err = workflow.ExecuteActivity(
		ctx,
		"GenerateReportEmail",
		reconReport,
	).Get(ctx, &emailBody)

	if err != nil {
		return fmt.Errorf("failed to generate recon email: %w", err)
	}

	err = workflow.ExecuteActivity(
		ctx,
		"SendReconEmail",
		&emailBody,
	).Get(ctx, nil)

	if err != nil {
		return fmt.Errorf("failed to send recon email: %w", err)
	}

	return nil
}
