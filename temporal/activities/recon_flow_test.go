package activities

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestGetReconTimeRange(t *testing.T) {
	activityRegistry := &Registry{}

	t.Run("returns yesterday's full day in Asia/Kolkata", func(t *testing.T) {
		result, err := activityRegistry.GetReconTimeRange(context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected recon time range, got nil")
		}

		location, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			t.Fatalf("failed to load timezone: %v", err)
		}

		now := time.Now().In(location)
		yesterday := now.AddDate(0, 0, -1)

		expectedStart := time.Date(
			yesterday.Year(),
			yesterday.Month(),
			yesterday.Day(),
			0, 0, 0, 0,
			location,
		)

		expectedEnd := expectedStart.AddDate(0, 0, 1)

		if !result.StartTimestamp.Equal(expectedStart) {
			t.Errorf(
				"expected start %v, got %v",
				expectedStart,
				result.StartTimestamp,
			)
		}

		if !result.EndTimestamp.Equal(expectedEnd) {
			t.Errorf(
				"expected end %v, got %v",
				expectedEnd,
				result.EndTimestamp,
			)
		}

		if result.EndTimestamp.Sub(result.StartTimestamp) != 24*time.Hour {
			t.Errorf(
				"expected 24 hour range, got %v",
				result.EndTimestamp.Sub(result.StartTimestamp),
			)
		}

		if result.StartTimestamp.Location().String() != "Asia/Kolkata" {
			t.Errorf(
				"expected location Asia/Kolkata, got %s",
				result.StartTimestamp.Location(),
			)
		}
	})
}

func TestGetLedgerTransactionsForRecon(t *testing.T) {
	activityRegistry := &Registry{
		Repo: testRepo,
	}

	t.Run("nil time range", func(t *testing.T) {
		transactions, err := activityRegistry.GetLedgerTransactionsForRecon(
			context.Background(),
			nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "recon time range is required" {
			t.Errorf(
				"expected error %q, got %q",
				"recon time range is required",
				err.Error(),
			)
		}

		if transactions != nil {
			t.Errorf("expected nil transactions, got %v", transactions)
		}
	})

	t.Run("returns transactions within time range", func(t *testing.T) {
		cleanTestDB(t)

		location, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			t.Fatalf("failed to load timezone: %v", err)
		}

		start := time.Date(
			2026,
			9,
			15,
			0, 0, 0, 0,
			location,
		)

		end := start.AddDate(0, 0, 1)

		transactionInside := &models.LedgerTransaction{
			ID:               utils.GenerateUUID(),
			TransactionRef:   "txn-recon-inside",
			ReferenceID:      "recon-inside",
			Type:             constants.LedgerTransactionTypePayment,
			Status:           constants.TransactionStatusCompleted,
			SettlementStatus: constants.LedgerSettlementSettled,
			CreatedAt:        start.Add(2 * time.Hour),
			UpdatedAt:        start.Add(2 * time.Hour),
		}

		transactionOutsideBefore := &models.LedgerTransaction{
			ID:               utils.GenerateUUID(),
			TransactionRef:   "txn-recon-before",
			ReferenceID:      "recon-before",
			Type:             constants.LedgerTransactionTypePayment,
			Status:           constants.TransactionStatusCompleted,
			SettlementStatus: constants.LedgerSettlementSettled,
			CreatedAt:        start.Add(-1 * time.Hour),
			UpdatedAt:        start.Add(-1 * time.Hour),
		}

		transactionOutsideAfter := &models.LedgerTransaction{
			ID:               utils.GenerateUUID(),
			TransactionRef:   "txn-recon-after",
			ReferenceID:      "recon-after",
			Type:             constants.LedgerTransactionTypePayment,
			Status:           constants.TransactionStatusCompleted,
			SettlementStatus: constants.LedgerSettlementSettled,
			CreatedAt:        end.Add(time.Hour),
			UpdatedAt:        end.Add(time.Hour),
		}

		if err := db.Create(transactionInside).Error; err != nil {
			t.Fatalf("failed to create inside transaction: %v", err)
		}

		if err := db.Create(transactionOutsideBefore).Error; err != nil {
			t.Fatalf("failed to create before transaction: %v", err)
		}

		if err := db.Create(transactionOutsideAfter).Error; err != nil {
			t.Fatalf("failed to create after transaction: %v", err)
		}

		transactions, err := activityRegistry.GetLedgerTransactionsForRecon(
			context.Background(),
			&ReconTimeRange{
				StartTimestamp: start,
				EndTimestamp:   end,
			},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(transactions) != 1 {
			t.Fatalf(
				"expected 1 transaction, got %d",
				len(transactions),
			)
		}

		if transactions[0].ID != transactionInside.ID {
			t.Errorf(
				"expected transaction %s, got %s",
				transactionInside.ID,
				transactions[0].ID,
			)
		}
	})

	t.Run("returns empty result when no transactions exist in range", func(t *testing.T) {
		cleanTestDB(t)

		location, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			t.Fatalf("failed to load timezone: %v", err)
		}

		start := time.Date(
			2026,
			9,
			15,
			0, 0, 0, 0,
			location,
		)

		end := start.AddDate(0, 0, 1)

		transaction := &models.LedgerTransaction{
			ID:               utils.GenerateUUID(),
			ReferenceID:      "outside-range",
			Type:             constants.LedgerTransactionTypePayment,
			Status:           constants.TransactionStatusCompleted,
			SettlementStatus: constants.LedgerSettlementSettled,
			CreatedAt:        end.Add(time.Hour),
			UpdatedAt:        end.Add(time.Hour),
		}

		if err := db.Create(transaction).Error; err != nil {
			t.Fatalf("failed to create transaction: %v", err)
		}

		transactions, err := activityRegistry.GetLedgerTransactionsForRecon(
			context.Background(),
			&ReconTimeRange{
				StartTimestamp: start,
				EndTimestamp:   end,
			},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if transactions == nil {
			t.Fatal("expected empty slice, got nil")
		}

		if len(transactions) != 0 {
			t.Errorf(
				"expected 0 transactions, got %d",
				len(transactions),
			)
		}
	})
}

func TestGetLedgerEntriesForRecon(t *testing.T) {
	activityRegistry := &Registry{
		Repo: testRepo,
	}

	t.Run("empty transaction IDs", func(t *testing.T) {
		entries, err := activityRegistry.GetLedgerEntriesForRecon(
			context.Background(),
			[]uuid.UUID{},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if entries == nil {
			t.Fatal("expected empty slice, got nil")
		}

		if len(entries) != 0 {
			t.Errorf(
				"expected 0 entries, got %d",
				len(entries),
			)
		}
	})

	t.Run("returns entries for given transaction IDs", func(t *testing.T) {
		cleanTestDB(t)

		transactionID := utils.GenerateUUID()
		otherTransactionID := utils.GenerateUUID()

		transaction := &models.LedgerTransaction{
			ID:               transactionID,
			TransactionRef:   "txn-recon-entry-test",
			ReferenceID:      "recon-entry-test",
			Type:             constants.LedgerTransactionTypePayment,
			Status:           constants.TransactionStatusCompleted,
			SettlementStatus: constants.LedgerSettlementSettled,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		otherTransaction := &models.LedgerTransaction{
			ID:               otherTransactionID,
			TransactionRef:   "txn-recon-other",
			ReferenceID:      "recon-other",
			Type:             constants.LedgerTransactionTypePayment,
			Status:           constants.TransactionStatusCompleted,
			SettlementStatus: constants.LedgerSettlementSettled,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if err := db.Create(transaction).Error; err != nil {
			t.Fatalf("failed to create transaction: %v", err)
		}

		if err := db.Create(otherTransaction).Error; err != nil {
			t.Fatalf("failed to create other transaction: %v", err)
		}

		entry1 := &models.LedgerEntry{
			ID:                 utils.GenerateUUID(),
			LedgerTransactionID: transactionID,
			AccountType:        constants.LedgerAccountTypeWallet,
			AccountID:           utils.GenerateUUID(),
			EntryType:          constants.LedgerEntryTypeDebit,
			Amount:             decimal.NewFromInt(1000),
			Currency:            "INR",
		}

		entry2 := &models.LedgerEntry{
			ID:                 utils.GenerateUUID(),
			LedgerTransactionID: transactionID,
			AccountType:        constants.LedgerAccountTypeVault,
			AccountID:           utils.GenerateUUID(),
			EntryType:          constants.LedgerEntryTypeCredit,
			Amount:             decimal.NewFromInt(1000),
			Currency:            "INR",
		}

		otherEntry := &models.LedgerEntry{
			ID:                 utils.GenerateUUID(),
			LedgerTransactionID: otherTransactionID,
			AccountType:        constants.LedgerAccountTypeWallet,
			AccountID:           utils.GenerateUUID(),
			EntryType:          constants.LedgerEntryTypeDebit,
			Amount:             decimal.NewFromInt(500),
			Currency:            "INR",
		}

		if err := db.Create(entry1).Error; err != nil {
			t.Fatalf("failed to create entry1: %v", err)
		}

		if err := db.Create(entry2).Error; err != nil {
			t.Fatalf("failed to create entry2: %v", err)
		}

		if err := db.Create(otherEntry).Error; err != nil {
			t.Fatalf("failed to create other entry: %v", err)
		}

		entries, err := activityRegistry.GetLedgerEntriesForRecon(
			context.Background(),
			[]uuid.UUID{transactionID},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(entries) != 2 {
			t.Fatalf(
				"expected 2 entries, got %d",
				len(entries),
			)
		}

		for _, entry := range entries {
			if entry == nil {
				t.Error("expected non-nil ledger entry")
				continue
			}

			if entry.LedgerTransactionID != transactionID {
				t.Errorf(
					"expected transaction ID %s, got %s",
					transactionID,
					entry.LedgerTransactionID,
				)
			}
		}
	})

	t.Run("returns entries for multiple transaction IDs", func(t *testing.T) {
		cleanTestDB(t)

		transactionID1 := utils.GenerateUUID()
		transactionID2 := utils.GenerateUUID()

		transaction1 := &models.LedgerTransaction{
			ID:               transactionID1,
			TransactionRef:   "txn-recon-multi-1",
			ReferenceID:      "recon-multi-1",
			Type:             constants.LedgerTransactionTypePayment,
			Status:           constants.TransactionStatusCompleted,
			SettlementStatus: constants.LedgerSettlementSettled,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		transaction2 := &models.LedgerTransaction{
			ID:               transactionID2,
			TransactionRef:   "txn-recon-multi-2",
			ReferenceID:      "recon-multi-2",
			Type:             constants.LedgerTransactionTypePayment,
			Status:           constants.TransactionStatusCompleted,
			SettlementStatus: constants.LedgerSettlementSettled,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if err := db.Create(transaction1).Error; err != nil {
			t.Fatalf("failed to create transaction1: %v", err)
		}

		if err := db.Create(transaction2).Error; err != nil {
			t.Fatalf("failed to create transaction2: %v", err)
		}

		entry1 := &models.LedgerEntry{
			ID:                  utils.GenerateUUID(),
			LedgerTransactionID: transactionID1,
			AccountType:         constants.LedgerAccountTypeWallet,
			AccountID:            utils.GenerateUUID(),
			EntryType:           constants.LedgerEntryTypeDebit,
			Amount:              decimal.NewFromInt(1000),
			Currency:            "INR",
		}

		entry2 := &models.LedgerEntry{
			ID:                  utils.GenerateUUID(),
			LedgerTransactionID: transactionID2,
			AccountType:         constants.LedgerAccountTypeVault,
			AccountID:            utils.GenerateUUID(),
			EntryType:            constants.LedgerEntryTypeCredit,
			Amount:              decimal.NewFromInt(2000),
			Currency:            "INR",
		}

		if err := db.Create(entry1).Error; err != nil {
			t.Fatalf("failed to create entry1: %v", err)
		}

		if err := db.Create(entry2).Error; err != nil {
			t.Fatalf("failed to create entry2: %v", err)
		}

		entries, err := activityRegistry.GetLedgerEntriesForRecon(
			context.Background(),
			[]uuid.UUID{transactionID1, transactionID2},
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(entries) != 2 {
			t.Fatalf(
				"expected 2 entries, got %d",
				len(entries),
			)
		}

		foundTransaction1 := false
		foundTransaction2 := false

		for _, entry := range entries {
			if entry == nil {
				t.Error("expected non-nil ledger entry")
				continue
			}

			if entry.LedgerTransactionID == transactionID1 {
				foundTransaction1 = true
			}

			if entry.LedgerTransactionID == transactionID2 {
				foundTransaction2 = true
			}
		}

		if !foundTransaction1 {
			t.Error("expected entry for transaction 1")
		}

		if !foundTransaction2 {
			t.Error("expected entry for transaction 2")
		}
	})
}

func TestReconcileLedgerTransactions(t *testing.T) {
	activityRegistry := &Registry{}

	t.Run("balanced transaction is matched", func(t *testing.T) {
		transactionID := utils.GenerateUUID()

		transactions := []*models.LedgerTransaction{
			{
				ID:          transactionID,
				ReferenceID: "recon-balanced",
			},
		}

		entries := []*models.LedgerEntry{
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeDebit,
				Amount:              decimal.NewFromInt(1000),
			},
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeCredit,
				Amount:              decimal.NewFromInt(1000),
			},
		}

		result, err := activityRegistry.ReconcileLedgerTransactions(
			context.Background(),
			transactions,
			entries,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if *result.TotalTransactions != 1 {
			t.Errorf(
				"expected total transactions 1, got %d",
				*result.TotalTransactions,
			)
		}

		if *result.MatchedCount != 1 {
			t.Errorf(
				"expected matched count 1, got %d",
				*result.MatchedCount,
			)
		}

		if *result.FailedCount != 0 {
			t.Errorf(
				"expected failed count 0, got %d",
				*result.FailedCount,
			)
		}

		if !result.TotalDebit.Equal(decimal.NewFromInt(1000)) {
			t.Errorf(
				"expected total debit 1000, got %s",
				result.TotalDebit.String(),
			)
		}

		if !result.TotalCredit.Equal(decimal.NewFromInt(1000)) {
			t.Errorf(
				"expected total credit 1000, got %s",
				result.TotalCredit.String(),
			)
		}

		if len(result.FailedTransactions) != 0 {
			t.Errorf(
				"expected no failed transactions, got %d",
				len(result.FailedTransactions),
			)
		}
	})

	t.Run("unbalanced transaction is marked as failed", func(t *testing.T) {
		transactionID := utils.GenerateUUID()

		transactions := []*models.LedgerTransaction{
			{
				ID:          transactionID,
				ReferenceID: "recon-failed",
			},
		}

		entries := []*models.LedgerEntry{
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeDebit,
				Amount:              decimal.NewFromInt(1000),
			},
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeCredit,
				Amount:              decimal.NewFromInt(800),
			},
		}

		result, err := activityRegistry.ReconcileLedgerTransactions(
			context.Background(),
			transactions,
			entries,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if *result.TotalTransactions != 1 {
			t.Errorf(
				"expected total transactions 1, got %d",
				*result.TotalTransactions,
			)
		}

		if *result.MatchedCount != 0 {
			t.Errorf(
				"expected matched count 0, got %d",
				*result.MatchedCount,
			)
		}

		if *result.FailedCount != 1 {
			t.Errorf(
				"expected failed count 1, got %d",
				*result.FailedCount,
			)
		}

		if !result.TotalDebit.Equal(decimal.NewFromInt(1000)) {
			t.Errorf(
				"expected total debit 1000, got %s",
				result.TotalDebit.String(),
			)
		}

		if !result.TotalCredit.Equal(decimal.NewFromInt(800)) {
			t.Errorf(
				"expected total credit 800, got %s",
				result.TotalCredit.String(),
			)
		}

		if len(result.FailedTransactions) != 1 {
			t.Fatalf(
				"expected 1 failed transaction, got %d",
				len(result.FailedTransactions),
			)
		}

		failure := result.FailedTransactions[0]

		if failure.LedgerTransactionID != transactionID {
			t.Errorf(
				"expected transaction ID %s, got %s",
				transactionID,
				failure.LedgerTransactionID,
			)
		}

		if failure.ReferenceID != "recon-failed" {
			t.Errorf(
				"expected reference ID recon-failed, got %s",
				failure.ReferenceID,
			)
		}

		if !failure.TotalDebit.Equal(decimal.NewFromInt(1000)) {
			t.Errorf(
				"expected failure debit 1000, got %s",
				failure.TotalDebit.String(),
			)
		}

		if !failure.TotalCredit.Equal(decimal.NewFromInt(800)) {
			t.Errorf(
				"expected failure credit 800, got %s",
				failure.TotalCredit.String(),
			)
		}

		if !failure.Difference.Equal(decimal.NewFromInt(200)) {
			t.Errorf(
				"expected difference 200, got %s",
				failure.Difference.String(),
			)
		}
	})

	t.Run("multiple entries are aggregated", func(t *testing.T) {
		transactionID := utils.GenerateUUID()

		transactions := []*models.LedgerTransaction{
			{
				ID:          transactionID,
				ReferenceID: "recon-aggregated",
			},
		}

		entries := []*models.LedgerEntry{
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeDebit,
				Amount:              decimal.NewFromInt(400),
			},
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeDebit,
				Amount:              decimal.NewFromInt(600),
			},
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeCredit,
				Amount:              decimal.NewFromInt(700),
			},
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeCredit,
				Amount:              decimal.NewFromInt(300),
			},
		}

		result, err := activityRegistry.ReconcileLedgerTransactions(
			context.Background(),
			transactions,
			entries,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.TotalDebit.Equal(decimal.NewFromInt(1000)) {
			t.Errorf(
				"expected total debit 1000, got %s",
				result.TotalDebit.String(),
			)
		}

		if !result.TotalCredit.Equal(decimal.NewFromInt(1000)) {
			t.Errorf(
				"expected total credit 1000, got %s",
				result.TotalCredit.String(),
			)
		}

		if *result.MatchedCount != 1 {
			t.Errorf(
				"expected matched count 1, got %d",
				*result.MatchedCount,
			)
		}

		if *result.FailedCount != 0 {
			t.Errorf(
				"expected failed count 0, got %d",
				*result.FailedCount,
			)
		}
	})

	t.Run("nil entry is ignored", func(t *testing.T) {
		transactionID := utils.GenerateUUID()

		transactions := []*models.LedgerTransaction{
			{
				ID:          transactionID,
				ReferenceID: "recon-nil-entry",
			},
		}

		entries := []*models.LedgerEntry{
			nil,
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeDebit,
				Amount:              decimal.NewFromInt(500),
			},
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeCredit,
				Amount:              decimal.NewFromInt(500),
			},
		}

		result, err := activityRegistry.ReconcileLedgerTransactions(
			context.Background(),
			transactions,
			entries,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if *result.MatchedCount != 1 {
			t.Errorf(
				"expected matched count 1, got %d",
				*result.MatchedCount,
			)
		}

		if !result.TotalDebit.Equal(decimal.NewFromInt(500)) {
			t.Errorf(
				"expected total debit 500, got %s",
				result.TotalDebit.String(),
			)
		}

		if !result.TotalCredit.Equal(decimal.NewFromInt(500)) {
			t.Errorf(
				"expected total credit 500, got %s",
				result.TotalCredit.String(),
			)
		}
	})

	t.Run("nil transaction is skipped", func(t *testing.T) {
		transactionID := utils.GenerateUUID()

		transactions := []*models.LedgerTransaction{
			nil,
			{
				ID:          transactionID,
				ReferenceID: "recon-valid",
			},
		}

		entries := []*models.LedgerEntry{
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeDebit,
				Amount:              decimal.NewFromInt(500),
			},
			{
				LedgerTransactionID: transactionID,
				EntryType:           constants.LedgerEntryTypeCredit,
				Amount:              decimal.NewFromInt(500),
			},
		}

		result, err := activityRegistry.ReconcileLedgerTransactions(
			context.Background(),
			transactions,
			entries,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// The implementation sets TotalTransactions to len(transactions),
		// so the nil transaction is included in this count.
		if *result.TotalTransactions != 2 {
			t.Errorf(
				"expected total transactions 2, got %d",
				*result.TotalTransactions,
			)
		}

		if *result.MatchedCount != 1 {
			t.Errorf(
				"expected matched count 1, got %d",
				*result.MatchedCount,
			)
		}

		if *result.FailedCount != 0 {
			t.Errorf(
				"expected failed count 0, got %d",
				*result.FailedCount,
			)
		}
	})
}

func TestGenerateReconReport(t *testing.T) {
	activityRegistry := &Registry{
		Services: testServices,
	}

	t.Run("nil recon result", func(t *testing.T) {
		report, err := activityRegistry.GenerateReconReport(
			context.Background(),
			nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "recon result cannot be nil" {
			t.Errorf(
				"expected error %q, got %q",
				"recon result cannot be nil",
				err.Error(),
			)
		}

		if report != nil {
			t.Errorf("expected nil report, got %+v", report)
		}
	})

	t.Run("generates passed report", func(t *testing.T) {
		reconDate := time.Date(
			2026,
			9,
			15,
			14, 30, 0, 0,
			time.FixedZone("IST", 5*60*60+30*60),
		)

		matchedCount := 5
		failedCount := 0
		totalTransactions := 5

		failedTransactions := make([]dto.ReconFailure, 0)

		result := &dto.ReconResult{
			ReconDate:          reconDate,
			TotalTransactions:  &totalTransactions,
			MatchedCount:       &matchedCount,
			FailedCount:        &failedCount,
			TotalDebit:         decimal.NewFromInt(5000),
			TotalCredit:        decimal.NewFromInt(5000),
			FailedTransactions: failedTransactions,
		}

		report, err := activityRegistry.GenerateReconReport(
			context.Background(),
			result,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report == nil {
			t.Fatal("expected report, got nil")
		}

		if report.ReconDate != "2026-09-15" {
			t.Errorf(
				"expected recon date 2026-09-15, got %s",
				report.ReconDate,
			)
		}

		if report.TotalTransactions != 5 {
			t.Errorf(
				"expected total transactions 5, got %d",
				report.TotalTransactions,
			)
		}

		if report.MatchedCount != 5 {
			t.Errorf(
				"expected matched count 5, got %d",
				report.MatchedCount,
			)
		}

		if report.FailedCount != 0 {
			t.Errorf(
				"expected failed count 0, got %d",
				report.FailedCount,
			)
		}

		if report.TotalDebit != "5000" {
			t.Errorf(
				"expected total debit 5000, got %s",
				report.TotalDebit,
			)
		}

		if report.TotalCredit != "5000" {
			t.Errorf(
				"expected total credit 5000, got %s",
				report.TotalCredit,
			)
		}

		if report.Difference != "0" {
			t.Errorf(
				"expected difference 0, got %s",
				report.Difference,
			)
		}

		if report.Status != "PASSED" {
			t.Errorf(
				"expected status PASSED, got %s",
				report.Status,
			)
		}

		if len(report.FailedTransactions) != 0 {
			t.Errorf(
				"expected 0 failed transactions, got %d",
				len(report.FailedTransactions),
			)
		}
	})

	t.Run("generates failed report", func(t *testing.T) {
		reconDate := time.Date(
			2026,
			9,
			15,
			14, 30, 0, 0,
			time.UTC,
		)

		matchedCount := 4
		failedCount := 1
		totalTransactions := 5

		failedTransactions := []dto.ReconFailure{
			{
				LedgerTransactionID: utils.GenerateUUID(),
				ReferenceID:         "failed-payment",
				TotalDebit:          decimal.NewFromInt(1000),
				TotalCredit:         decimal.NewFromInt(800),
				Difference:          decimal.NewFromInt(200),
			},
		}

		result := &dto.ReconResult{
			ReconDate:          reconDate,
			TotalTransactions:  &totalTransactions,
			MatchedCount:       &matchedCount,
			FailedCount:        &failedCount,
			TotalDebit:         decimal.NewFromInt(5000),
			TotalCredit:        decimal.NewFromInt(4800),
			FailedTransactions: failedTransactions,
		}

		report, err := activityRegistry.GenerateReconReport(
			context.Background(),
			result,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if report == nil {
			t.Fatal("expected report, got nil")
		}

		if report.ReconDate != "2026-09-15" {
			t.Errorf(
				"expected recon date 2026-09-15, got %s",
				report.ReconDate,
			)
		}

		if report.TotalTransactions != 5 {
			t.Errorf(
				"expected total transactions 5, got %d",
				report.TotalTransactions,
			)
		}

		if report.MatchedCount != 4 {
			t.Errorf(
				"expected matched count 4, got %d",
				report.MatchedCount,
			)
		}

		if report.FailedCount != 1 {
			t.Errorf(
				"expected failed count 1, got %d",
				report.FailedCount,
			)
		}

		if report.TotalDebit != "5000" {
			t.Errorf(
				"expected total debit 5000, got %s",
				report.TotalDebit,
			)
		}

		if report.TotalCredit != "4800" {
			t.Errorf(
				"expected total credit 4800, got %s",
				report.TotalCredit,
			)
		}

		if report.Difference != "200" {
			t.Errorf(
				"expected difference 200, got %s",
				report.Difference,
			)
		}

		if report.Status != "FAILED" {
			t.Errorf(
				"expected status FAILED, got %s",
				report.Status,
			)
		}

		if len(report.FailedTransactions) != 1 {
			t.Fatalf(
				"expected 1 failed transaction, got %d",
				len(report.FailedTransactions),
			)
		}

		if report.FailedTransactions[0].ReferenceID != "failed-payment" {
			t.Errorf(
				"expected failed transaction reference failed-payment, got %s",
				report.FailedTransactions[0].ReferenceID,
			)
		}

		if !report.FailedTransactions[0].Difference.Equal(decimal.NewFromInt(200)) {
			t.Errorf(
				"expected failed transaction difference 200, got %s",
				report.FailedTransactions[0].Difference.String(),
			)
		}
	})
}

func TestGenerateReportEmail(t *testing.T) {
	activityRegistry := &Registry{
		Services: testServices,
	}

	t.Run("generates email for passed report", func(t *testing.T) {
		report := &dto.ReconReport{
			ReconDate:         "2026-09-15",
			TotalTransactions: 5,
			MatchedCount:      5,
			FailedCount:       0,
			TotalDebit:        "5000",
			TotalCredit:       "5000",
			Difference:        "0",
			Status:            "PASSED",
			FailedTransactions: []dto.ReconFailure{},
		}

		expected := "SAMPay Reconciliation Report\n" +
			"============================\n\n" +
			"Recon Date: 2026-09-15\n" +
			"Total Transactions: 5\n" +
			"Matched: 5\n" +
			"Failed: 0\n" +
			"Total Debit: 5000\n" +
			"Total Credit: 5000\n" +
			"Difference: 0\n" +
			"Status: PASSED\n\n"

		emailBody, err := activityRegistry.GenerateReportEmail(
			context.Background(),
			report,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if emailBody == "" {
			t.Fatal("expected email body, got nil")
		}

		if emailBody != expected {
			t.Errorf(
				"email body mismatch\nexpected:\n%s\ngot:\n%s",
				expected,
				emailBody,
			)
		}
	})

	t.Run("generates email for failed report with failed transactions", func(t *testing.T) {
		transactionID := utils.GenerateUUID()

		report := &dto.ReconReport{
			ReconDate:         "2026-09-15",
			TotalTransactions: 3,
			MatchedCount:      2,
			FailedCount:       1,
			TotalDebit:        "3000",
			TotalCredit:       "2800",
			Difference:        "200",
			Status:            "FAILED",
			FailedTransactions: []dto.ReconFailure{
				{
					LedgerTransactionID: transactionID,
					ReferenceID:         "payment-ref-123",
					TotalDebit:          decimal.NewFromInt(1000),
					TotalCredit:         decimal.NewFromInt(800),
					Difference:          decimal.NewFromInt(200),
				},
			},
		}

		expected := fmt.Sprintf(
			"SAMPay Reconciliation Report\n"+
				"============================\n\n"+
				"Recon Date: 2026-09-15\n"+
				"Total Transactions: 3\n"+
				"Matched: 2\n"+
				"Failed: 1\n"+
				"Total Debit: 3000\n"+
				"Total Credit: 2800\n"+
				"Difference: 200\n"+
				"Status: FAILED\n\n"+
				"Failed Transactions\n"+
				"-------------------\n"+
				"Ledger Transaction ID: %s\n"+
				"Reference ID: payment-ref-123\n"+
				"Debit: 1000\n"+
				"Credit: 800\n"+
				"Difference: 200\n\n",
			transactionID,
		)

		emailBody, err := activityRegistry.GenerateReportEmail(
			context.Background(),
			report,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if emailBody == "" {
			t.Fatal("expected email body, got nil")
		}

		if emailBody != expected {
			t.Errorf(
				"email body mismatch\nexpected:\n%s\ngot:\n%s",
				expected,
				emailBody,
			)
		}
	})
}

func TestSendReconEmail(t *testing.T) {
	activityRegistry := &Registry{
		NotificationService: testServices.NotificationService,
	}

	t.Run("nil email body", func(t *testing.T) {
		err := activityRegistry.SendReconEmail(
			context.Background(),
			nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "recon email body is required" {
			t.Errorf(
				"expected error %q, got %q",
				"recon email body is required",
				err.Error(),
			)
		}
	})

	t.Run("returns error when email sending fails", func(t *testing.T) {
		emailBody := "SAMPay Reconciliation Report\nStatus: PASSED"

		err := activityRegistry.SendReconEmail(
			context.Background(),
			&emailBody,
		)

		if err == nil {
			t.Fatal("expected email sending error, got nil")
		}
	})
}

