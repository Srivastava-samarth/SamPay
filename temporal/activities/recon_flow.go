package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ReconTimeRange struct {
	StartTimestamp time.Time
	EndTimestamp   time.Time
}

func (a *Registry) GetReconTimeRange(
	ctx context.Context,
) (*ReconTimeRange, error) {

	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return nil, err
	}

	now := time.Now().In(location)

	yesterday := now.AddDate(0, 0, -1)

	startTimestamp := time.Date(
		yesterday.Year(),
		yesterday.Month(),
		yesterday.Day(),
		0, 0, 0, 0,
		location,
	)

	endTimestamp := startTimestamp.AddDate(0, 0, 1)

	return &ReconTimeRange{
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
	}, nil
}

func (a *Registry) GetLedgerTransactionsForRecon(
	ctx context.Context,
	timeRange *ReconTimeRange,
) ([]*models.LedgerTransaction, error) {

	if timeRange == nil {
		return nil, errors.New("recon time range is required")
	}

	transactions, err := a.Repo.
		GetTransactionsByCreatedAtRange(
			timeRange.StartTimestamp,
			timeRange.EndTimestamp,
		)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger transactions for recon: %w", err)
	}

	return transactions, nil
}

func (a *Registry) GetLedgerEntriesForRecon(
	ctx context.Context,
	transactionIDs []uuid.UUID,
) ([]*models.LedgerEntry, error) {

	if len(transactionIDs) == 0 {
		return []*models.LedgerEntry{}, nil
	}

	entries, err := a.Repo.
		GetEntriesByTransactionIDs(transactionIDs)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get ledger entries for recon: %w",
			err,
		)
	}

	return entries, nil
}

func (a *Registry) ReconcileLedgerTransactions(
	ctx context.Context,
	transactions []*models.LedgerTransaction,
	entries []*models.LedgerEntry,
) (*dto.ReconResult, error) {

	entriesByTransactionID := make(
		map[uuid.UUID][]*models.LedgerEntry,
	)

	for _, entry := range entries {
		if entry == nil {
			continue
		}

		entriesByTransactionID[entry.LedgerTransactionID] =
			append(
				entriesByTransactionID[entry.LedgerTransactionID],
				entry,
			)
	}

	result := &dto.ReconResult{
		ReconDate:          time.Now(),
		TotalTransactions:  new(int),
		MatchedCount:       new(int),
		FailedCount:        new(int),
		FailedTransactions: make([]dto.ReconFailure, 0),
	}

	*result.TotalTransactions = len(transactions)

	for _, transaction := range transactions {
		if transaction == nil {
			continue
		}

		transactionEntries := entriesByTransactionID[transaction.ID]

		transactionDebit := decimal.Zero
		transactionCredit := decimal.Zero

		for _, entry := range transactionEntries {
			if entry == nil {
				continue
			}

			switch entry.EntryType {
			case constants.LedgerEntryTypeDebit:
				transactionDebit = transactionDebit.Add(entry.Amount)

			case constants.LedgerEntryTypeCredit:
				transactionCredit = transactionCredit.Add(entry.Amount)
			}
		}

		result.TotalDebit = result.TotalDebit.Add(transactionDebit)
		result.TotalCredit = result.TotalCredit.Add(transactionCredit)

		if transactionDebit.Equal(transactionCredit) {
			*result.MatchedCount++
			continue
		}

		*result.FailedCount++

		result.FailedTransactions = append(
			result.FailedTransactions,
			dto.ReconFailure{
				LedgerTransactionID: transaction.ID,
				ReferenceID:         transaction.ReferenceID,
				TotalDebit:          transactionDebit,
				TotalCredit:         transactionCredit,
				Difference:          transactionDebit.Sub(transactionCredit),
			},
		)
	}

	return result, nil
}

func (a *Registry) GenerateReconReport(
	ctx context.Context,
	result *dto.ReconResult,
) (*dto.ReconReport, error) {
	return a.Services.GenerateReconReport(result)
}

func (a *Registry) GenerateReportEmail(
	ctx context.Context,
	report *dto.ReconReport,
) (string, error) {
	return a.Services.GenerateReconEmailBody(report)
}

func (a *Registry) SendReconEmail(
	ctx context.Context,
	emailBody *string,
) error {

	if emailBody == nil {
		return errors.New("recon email body is required")
	}

	return a.NotificationService.SendReconEmail(
		*emailBody,
	)
}
