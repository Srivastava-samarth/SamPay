package repositories

import (
	"errors"
	"testing"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func TestCreateLedgerEntry(t *testing.T) {

	ledgerTransaction := &models.LedgerTransaction{
		ID:               uuid.New(),
		TransactionRef:   utils.GenerateLedgerReference(),
		Type:             "payment",
		ReferenceID:      uuid.NewString(),
		Status:           constants.TransactionStatusCompleted,
		SettlementStatus: constants.LedgerSettlementPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create(ledgerTransaction).Error; err != nil {
		t.Fatalf("failed to create ledger transaction: %v", err)
	}

	accountID := uuid.New()
	accountType := "wallet"
	entryType := "credit"
	currency := "USDT"
	amount := decimal.NewFromInt(100)

	requestEntry := &models.LedgerEntry{
		LedgerTransactionID: ledgerTransaction.ID,
		AccountType:         accountType,
		EntryType:           entryType,
		AccountID:           accountID,
		Amount:              amount,
		Currency:            currency,
	}

	result, err := testRepo.CreateLedgerEntry(requestEntry)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected ledger entry, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected generated ledger entry ID, got nil UUID")
	}

	if result.LedgerTransactionID != ledgerTransaction.ID {
		t.Errorf(
			"expected ledger transaction ID %v, got %v",
			ledgerTransaction.ID,
			result.LedgerTransactionID,
		)
	}

	if result.AccountType != accountType {
		t.Errorf("expected account type %q, got %q", accountType, result.AccountType)
	}

	if result.EntryType != entryType {
		t.Errorf("expected entry type %q, got %q", entryType, result.EntryType)
	}

	if result.AccountID != accountID {
		t.Errorf("expected account ID %v, got %v", accountID, result.AccountID)
	}

	if !result.Amount.Equal(amount) {
		t.Errorf("expected amount %v, got %v", amount, result.Amount)
	}

	if result.Currency == "" || result.Currency != currency {
		t.Errorf("expected currency %q, got %v", currency, result.Currency)
	}

	if result.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestCreateLedgerTransaction(t *testing.T) {

	referenceID := uuid.NewString()

	requestTransaction := &models.LedgerTransaction{
		Type:             "payment",
		ReferenceID:      referenceID,
		Status:           constants.TransactionStatusCompleted,
		SettlementStatus: constants.LedgerSettlementPending,
	}

	result, err := testRepo.CreateLedgerTransaction(requestTransaction)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected ledger transaction, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected generated ID, got nil UUID")
	}

	if result.TransactionRef == "" {
		t.Error("expected generated transaction reference, got empty")
	}

	if result.Type != requestTransaction.Type {
		t.Errorf(
			"expected type %q, got %q",
			requestTransaction.Type,
			result.Type,
		)
	}

	if result.ReferenceID != requestTransaction.ReferenceID {
		t.Errorf(
			"expected reference ID %q, got %q",
			requestTransaction.ReferenceID,
			result.ReferenceID,
		)
	}

	if result.Status != requestTransaction.Status {
		t.Errorf(
			"expected status %q, got %q",
			requestTransaction.Status,
			result.Status,
		)
	}

	if result.SettlementStatus != requestTransaction.SettlementStatus {
		t.Errorf(
			"expected settlement status %q, got %q",
			requestTransaction.SettlementStatus,
			result.SettlementStatus,
		)
	}

	if result.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if result.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}

	// Verify that the record was actually persisted.
	var stored models.LedgerTransaction
	if err := db.Where("id = ?", result.ID).First(&stored).Error; err != nil {
		t.Fatalf("failed to fetch created ledger transaction: %v", err)
	}

	if stored.ID != result.ID {
		t.Errorf(
			"expected persisted ID %v, got %v",
			result.ID,
			stored.ID,
		)
	}
}

func TestExistingLedgerTransactionByReferenceID(t *testing.T) {

	referenceID := uuid.NewString()

	ledgerTransaction := &models.LedgerTransaction{
		ID:               uuid.New(),
		TransactionRef:   utils.GenerateLedgerReference(),
		Type:             "payment",
		ReferenceID:      referenceID,
		Status:           constants.TransactionStatusCompleted,
		SettlementStatus: constants.LedgerSettlementPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create(ledgerTransaction).Error; err != nil {
		t.Fatalf("failed to create ledger transaction: %v", err)
	}

	t.Run("existing reference ID", func(t *testing.T) {
		result, err := testRepo.ExistingLedgerTransactionByReferenceID(referenceID)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if !result {
			t.Error("expected true for existing reference ID, got false")
		}
	})

	t.Run("non-existing reference ID", func(t *testing.T) {
		nonExistingReferenceID := uuid.NewString()

		result, err := testRepo.ExistingLedgerTransactionByReferenceID(
			nonExistingReferenceID,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result {
			t.Error("expected false for non-existing reference ID, got true")
		}
	})
}

func TestGetLedgerTransactionByID(t *testing.T) {

	ledgerTransaction := &models.LedgerTransaction{
		ID:               uuid.New(),
		TransactionRef:   utils.GenerateLedgerReference(),
		Type:             "payment",
		ReferenceID:      uuid.NewString(),
		Status:           constants.TransactionStatusCompleted,
		SettlementStatus: constants.LedgerSettlementPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create(ledgerTransaction).Error; err != nil {
		t.Fatalf("failed to create ledger transaction: %v", err)
	}

	t.Run("existing ledger transaction", func(t *testing.T) {
		result, err := testRepo.GetLedgerTransactionByID(ledgerTransaction.ID)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("expected ledger transaction, got nil")
		}

		if result.ID != ledgerTransaction.ID {
			t.Errorf(
				"expected ID %v, got %v",
				ledgerTransaction.ID,
				result.ID,
			)
		}

		if result.TransactionRef != ledgerTransaction.TransactionRef {
			t.Errorf(
				"expected transaction reference %v, got %v",
				ledgerTransaction.TransactionRef,
				result.TransactionRef,
			)
		}

		if result.ReferenceID != ledgerTransaction.ReferenceID {
			t.Errorf(
				"expected reference ID %v, got %v",
				ledgerTransaction.ReferenceID,
				result.ReferenceID,
			)
		}
	})

	t.Run("non-existing ledger transaction", func(t *testing.T) {
		nonExistingID := uuid.New()

		result, err := testRepo.GetLedgerTransactionByID(nonExistingID)

		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf(
				"expected gorm.ErrRecordNotFound, got %v",
				err,
			)
		}
	})
}

func TestGetTransactionsByCreatedAtRange(t *testing.T) {

	baseTime := time.Now().Add(-10 * time.Minute)

	insideTransaction1 := &models.LedgerTransaction{
		ID:               uuid.New(),
		TransactionRef:   utils.GenerateLedgerReference(),
		Type:             "payment",
		ReferenceID:      uuid.NewString(),
		Status:           constants.TransactionStatusCompleted,
		SettlementStatus: constants.LedgerSettlementPending,
		CreatedAt:        baseTime.Add(2 * time.Second),
		UpdatedAt:        baseTime.Add(2 * time.Second),
	}

	insideTransaction2 := &models.LedgerTransaction{
		ID:               uuid.New(),
		TransactionRef:   utils.GenerateLedgerReference(),
		Type:             "payment",
		ReferenceID:      uuid.NewString(),
		Status:           constants.TransactionStatusCompleted,
		SettlementStatus: constants.LedgerSettlementPending,
		CreatedAt:        baseTime.Add(4 * time.Second),
		UpdatedAt:        baseTime.Add(4 * time.Second),
	}

	outsideTransaction := &models.LedgerTransaction{
		ID:               uuid.New(),
		TransactionRef:   utils.GenerateLedgerReference(),
		Type:             "payment",
		ReferenceID:      uuid.NewString(),
		Status:           constants.TransactionStatusCompleted,
		SettlementStatus: constants.LedgerSettlementPending,
		CreatedAt:        baseTime.Add(-1 * time.Second),
		UpdatedAt:        baseTime.Add(-1 * time.Second),
	}

	transactions := []*models.LedgerTransaction{
		insideTransaction1,
		insideTransaction2,
		outsideTransaction,
	}

	if err := db.Create(&transactions).Error; err != nil {
		t.Fatalf("failed to create ledger transactions: %v", err)
	}

	startTimestamp := baseTime
	endTimestamp := baseTime.Add(5 * time.Second)

	result, err := testRepo.GetTransactionsByCreatedAtRange(
		startTimestamp,
		endTimestamp,
	)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	found := make(map[uuid.UUID]bool)

	for _, transaction := range result {
		found[transaction.ID] = true
	}

	if !found[insideTransaction1.ID] {
		t.Errorf("expected transaction %v to be returned", insideTransaction1.ID)
	}

	if !found[insideTransaction2.ID] {
		t.Errorf("expected transaction %v to be returned", insideTransaction2.ID)
	}

	if found[outsideTransaction.ID] {
		t.Errorf("did not expect transaction %v to be returned", outsideTransaction.ID)
	}

	// Verify ordering among our transactions.
	var ourTransactions []*models.LedgerTransaction

	for _, transaction := range result {
		if transaction.ID == insideTransaction1.ID ||
			transaction.ID == insideTransaction2.ID {
			ourTransactions = append(ourTransactions, transaction)
		}
	}

	if len(ourTransactions) != 2 {
		t.Fatalf(
			"expected both test transactions to be returned, got %d",
			len(ourTransactions),
		)
	}

	if !ourTransactions[0].CreatedAt.Before(ourTransactions[1].CreatedAt) {
		t.Errorf(
			"expected ascending created_at order, got %v and %v",
			ourTransactions[0].CreatedAt,
			ourTransactions[1].CreatedAt,
		)
	}
}

func TestGetEntriesByTransactionIDs(t *testing.T) {

	transaction1 := &models.LedgerTransaction{
		ID:               uuid.New(),
		TransactionRef:   utils.GenerateLedgerReference(),
		Type:             "payment",
		ReferenceID:      uuid.NewString(),
		Status:           constants.TransactionStatusCompleted,
		SettlementStatus: constants.LedgerSettlementPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	transaction2 := &models.LedgerTransaction{
		ID:               uuid.New(),
		TransactionRef:   utils.GenerateLedgerReference(),
		Type:             "payment",
		ReferenceID:      uuid.NewString(),
		Status:           constants.TransactionStatusCompleted,
		SettlementStatus: constants.LedgerSettlementPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create([]*models.LedgerTransaction{
		transaction1,
		transaction2,
	}).Error; err != nil {
		t.Fatalf("failed to create ledger transactions: %v", err)
	}

	accountID1 := uuid.New()
	accountID2 := uuid.New()
	accountType := "wallet"
	entryType := "credit"
	currency := "INR"

	entry1 := &models.LedgerEntry{
		ID:                  uuid.New(),
		LedgerTransactionID: transaction1.ID,
		AccountType:         accountType,
		AccountID:           accountID1,
		EntryType:           entryType,
		Amount:              decimal.NewFromInt(100),
		Currency:            currency,
		CreatedAt:           time.Now().Add(-2 * time.Second),
	}

	entry2 := &models.LedgerEntry{
		ID:                  uuid.New(),
		LedgerTransactionID: transaction1.ID,
		AccountType:         accountType,
		AccountID:           accountID2,
		EntryType:           entryType,
		Amount:              decimal.NewFromInt(200),
		Currency:            currency,
		CreatedAt:           time.Now().Add(-1 * time.Second),
	}

	entryFromOtherTransaction := &models.LedgerEntry{
		ID:                  uuid.New(),
		LedgerTransactionID: transaction2.ID,
		AccountType:         accountType,
		AccountID:           uuid.New(),
		EntryType:           entryType,
		Amount:              decimal.NewFromInt(300),
		Currency:            currency,
		CreatedAt:           time.Now(),
	}

	if err := db.Create([]*models.LedgerEntry{
		entry1,
		entry2,
		entryFromOtherTransaction,
	}).Error; err != nil {
		t.Fatalf("failed to create ledger entries: %v", err)
	}

	result, err := testRepo.GetEntriesByTransactionIDs(
		[]uuid.UUID{transaction1.ID},
	)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	found := make(map[uuid.UUID]*models.LedgerEntry)

	for _, entry := range result {
		found[entry.ID] = entry
	}

	if _, ok := found[entry1.ID]; !ok {
		t.Errorf("expected entry %v to be returned", entry1.ID)
	}

	if _, ok := found[entry2.ID]; !ok {
		t.Errorf("expected entry %v to be returned", entry2.ID)
	}

	if _, ok := found[entryFromOtherTransaction.ID]; ok {
		t.Errorf(
			"did not expect entry %v from another transaction",
			entryFromOtherTransaction.ID,
		)
	}

	var ourEntries []*models.LedgerEntry

	for _, entry := range result {
		if entry.ID == entry1.ID || entry.ID == entry2.ID {
			ourEntries = append(ourEntries, entry)
		}
	}

	if len(ourEntries) != 2 {
		t.Fatalf(
			"expected both test entries to be returned, got %d",
			len(ourEntries),
		)
	}

	if !ourEntries[0].CreatedAt.Before(ourEntries[1].CreatedAt) {
		t.Errorf(
			"expected ascending created_at order, got %v and %v",
			ourEntries[0].CreatedAt,
			ourEntries[1].CreatedAt,
		)
	}
}

func TestGetLedgerTransactionByReferenceID(t *testing.T) {

	referenceID := uuid.NewString()

	ledgerTransaction := &models.LedgerTransaction{
		ID:               uuid.New(),
		TransactionRef:   utils.GenerateLedgerReference(),
		Type:             "payment",
		ReferenceID:      referenceID,
		Status:           constants.TransactionStatusCompleted,
		SettlementStatus: constants.LedgerSettlementPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create(ledgerTransaction).Error; err != nil {
		t.Fatalf("failed to create ledger transaction: %v", err)
	}

	t.Run("existing reference ID", func(t *testing.T) {
		result, err := testRepo.GetLedgerTransactionByReferenceID(referenceID)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("expected ledger transaction, got nil")
		}

		if result.ID != ledgerTransaction.ID {
			t.Errorf(
				"expected ID %v, got %v",
				ledgerTransaction.ID,
				result.ID,
			)
		}

		if result.ReferenceID != ledgerTransaction.ReferenceID {
			t.Errorf(
				"expected reference ID %v, got %v",
				ledgerTransaction.ReferenceID,
				result.ReferenceID,
			)
		}

		if result.TransactionRef != ledgerTransaction.TransactionRef {
			t.Errorf(
				"expected transaction reference %v, got %v",
				ledgerTransaction.TransactionRef,
				result.TransactionRef,
			)
		}
	})

	t.Run("non-existing reference ID", func(t *testing.T) {
		nonExistingReferenceID := uuid.NewString()

		result, err := testRepo.GetLedgerTransactionByReferenceID(
			nonExistingReferenceID,
		)

		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf(
				"expected gorm.ErrRecordNotFound, got %v",
				err,
			)
		}
	})
}

func TestUpdateLedgerStatus(t *testing.T) {

	ledgerTransaction := &models.LedgerTransaction{
		ID:               uuid.New(),
		TransactionRef:   utils.GenerateLedgerReference(),
		Type:             "payment",
		ReferenceID:      uuid.NewString(),
		Status:           constants.TransactionStatusProcessing,
		SettlementStatus: constants.LedgerSettlementPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create(ledgerTransaction).Error; err != nil {
		t.Fatalf("failed to create ledger transaction: %v", err)
	}

	t.Run("update status", func(t *testing.T) {
		newStatus := constants.TransactionStatusCompleted
		settlementStatus := ledgerTransaction.SettlementStatus

		result, err := testRepo.UpdateLedgerStatus(
			ledgerTransaction.ID,
			newStatus,
			settlementStatus,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("expected ledger transaction, got nil")
		}

		if result.Status != newStatus {
			t.Errorf(
				"expected status %v, got %v",
				newStatus,
				result.Status,
			)
		}
	})

	t.Run("update settlement status", func(t *testing.T) {
		status := constants.TransactionStatusCompleted
		newSettlementStatus := constants.LedgerSettlementSettled

		result, err := testRepo.UpdateLedgerStatus(
			ledgerTransaction.ID,
			status,
			newSettlementStatus,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("expected ledger transaction, got nil")
		}

		if result.SettlementStatus != newSettlementStatus {
			t.Errorf(
				"expected settlement status %v, got %v",
				newSettlementStatus,
				result.SettlementStatus,
			)
		}
	})

	t.Run("no status change", func(t *testing.T) {
		currentStatus := constants.TransactionStatusCompleted
		currentSettlementStatus := constants.LedgerSettlementSettled

		result, err := testRepo.UpdateLedgerStatus(
			ledgerTransaction.ID,
			currentStatus,
			currentSettlementStatus,
		)

		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetWalletTransactions(t *testing.T) {

	accountType := "wallet"
	accountID := uuid.New()

	baseTime := time.Now().Add(-1 * time.Hour)

	var transactionIDs []uuid.UUID

	// Create 12 transactions for the same wallet.
	for i := 0; i < 12; i++ {
		transaction := &models.LedgerTransaction{
			ID:               uuid.New(),
			TransactionRef:   utils.GenerateLedgerReference(),
			Type:             "payment",
			ReferenceID:      uuid.NewString(),
			Status:           "completed",
			SettlementStatus: "pending",
			CreatedAt:        baseTime.Add(time.Duration(i) * time.Minute),
			UpdatedAt:        baseTime.Add(time.Duration(i) * time.Minute),
		}

		if err := db.Create(transaction).Error; err != nil {
			t.Fatalf("failed to create ledger transaction: %v", err)
		}

		transactionIDs = append(transactionIDs, transaction.ID)

		entry := &models.LedgerEntry{
			ID:                  uuid.New(),
			LedgerTransactionID: transaction.ID,
			AccountType:         accountType,
			AccountID:           accountID,
			Amount:              decimal.NewFromInt(int64(i + 1)),
			Currency:            "INR",
			EntryType:           "credit",
			CreatedAt:           transaction.CreatedAt,
		}

		if err := db.Create(entry).Error; err != nil {
			t.Fatalf("failed to create ledger entry: %v", err)
		}
	}

	t.Run("without cursor", func(t *testing.T) {
		result, err := testRepo.GetWalletTransactions(
			&accountType,
			accountID,
			nil,
			"",
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 11 {
			t.Fatalf("expected 11 transactions, got %d", len(result))
		}

		// Newest transaction should be first.
		if result[0].LedgerTransactionID != transactionIDs[11] {
			t.Errorf(
				"expected first transaction %v, got %v",
				transactionIDs[11],
				result[0].LedgerTransactionID,
			)
		}

		// 11th transaction should be transaction index 1.
		if result[10].LedgerTransactionID != transactionIDs[1] {
			t.Errorf(
				"expected last transaction %v, got %v",
				transactionIDs[1],
				result[10].LedgerTransactionID,
			)
		}
	})

	t.Run("next cursor", func(t *testing.T) {
		cursor := transactionIDs[7]

		result, err := testRepo.GetWalletTransactions(
			&accountType,
			accountID,
			&cursor,
			"next",
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) == 0 {
			t.Fatal("expected transactions, got none")
		}

		// Cursor itself is included because the query uses <=.
		if result[0].LedgerTransactionID != transactionIDs[7] {
			t.Errorf(
				"expected first transaction %v, got %v",
				transactionIDs[7],
				result[0].LedgerTransactionID,
			)
		}

		if len(result) > 1 &&
			result[1].LedgerTransactionID != transactionIDs[6] {
			t.Errorf(
				"expected second transaction %v, got %v",
				transactionIDs[6],
				result[1].LedgerTransactionID,
			)
		}
	})

	t.Run("previous cursor", func(t *testing.T) {
		cursor := transactionIDs[4]

		result, err := testRepo.GetWalletTransactions(
			&accountType,
			accountID,
			&cursor,
			"previous",
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) == 0 {
			t.Fatal("expected transactions, got none")
		}

		// Previous transactions are returned in ASC order.
		if result[0].LedgerTransactionID != transactionIDs[5] {
			t.Errorf(
				"expected first transaction %v, got %v",
				transactionIDs[5],
				result[0].LedgerTransactionID,
			)
		}

		if len(result) > 1 &&
			result[1].LedgerTransactionID != transactionIDs[6] {
			t.Errorf(
				"expected second transaction %v, got %v",
				transactionIDs[6],
				result[1].LedgerTransactionID,
			)
		}
	})
}

func TestGetWalletTransactionById(t *testing.T) {

	accountID := uuid.New()
	otherAccountID := uuid.New()
	accountType := "wallet"

	transaction := &models.LedgerTransaction{
		ID:               uuid.New(),
		TransactionRef:   utils.GenerateLedgerReference(),
		Type:             "payment",
		ReferenceID:      uuid.NewString(),
		Status:           "completed",
		SettlementStatus: "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create(transaction).Error; err != nil {
		t.Fatalf("failed to create ledger transaction: %v", err)
	}

	entry := &models.LedgerEntry{
		ID:                  uuid.New(),
		LedgerTransactionID: transaction.ID,
		AccountType:         accountType,
		AccountID:           accountID,
		Amount:              decimal.NewFromInt(500),
		Currency:            "INR",
		EntryType:           "credit",
		CreatedAt:           time.Now(),
	}

	if err := db.Create(entry).Error; err != nil {
		t.Fatalf("failed to create ledger entry: %v", err)
	}

	t.Run("existing transaction", func(t *testing.T) {
		result, err := testRepo.GetWalletTransactionById(
			accountID,
			transaction.ID,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected transaction, got nil")
		}

		if result.LedgerTransactionID != transaction.ID {
			t.Errorf(
				"expected transaction ID %v, got %v",
				transaction.ID,
				result.LedgerTransactionID,
			)
		}

		if result.TransactionRef != transaction.TransactionRef {
			t.Errorf(
				"expected transaction reference %v, got %v",
				transaction.TransactionRef,
				result.TransactionRef,
			)
		}

		if result.Type != transaction.Type {
			t.Errorf(
				"expected type %v, got %v",
				transaction.Type,
				result.Type,
			)
		}

		if result.ReferenceID != transaction.ReferenceID {
			t.Errorf(
				"expected reference ID %v, got %v",
				transaction.ReferenceID,
				result.ReferenceID,
			)
		}

		if result.Status != transaction.Status {
			t.Errorf(
				"expected status %v, got %v",
				transaction.Status,
				result.Status,
			)
		}

		if result.SettlementStatus != transaction.SettlementStatus {
			t.Errorf(
				"expected settlement status %v, got %v",
				transaction.SettlementStatus,
				result.SettlementStatus,
			)
		}

		if result.Amount.Cmp(entry.Amount) != 0 {
			t.Errorf(
				"expected amount %v, got %v",
				entry.Amount,
				result.Amount,
			)
		}

		if result.Currency != entry.Currency {
			t.Errorf(
				"expected currency %v, got %v",
				entry.Currency,
				result.Currency,
			)
		}

		if result.EntryType != entry.EntryType {
			t.Errorf(
				"expected entry type %v, got %v",
				entry.EntryType,
				result.EntryType,
			)
		}
	})

	t.Run("wrong account", func(t *testing.T) {
		result, err := testRepo.GetWalletTransactionById(
			otherAccountID,
			transaction.ID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if result != nil {
			t.Errorf("expected nil transaction, got %v", result)
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf("expected record not found error, got %v", err)
		}
	})

	t.Run("non existing transaction", func(t *testing.T) {
		result, err := testRepo.GetWalletTransactionById(
			accountID,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if result != nil {
			t.Errorf("expected nil transaction, got %v", result)
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf("expected record not found error, got %v", err)
		}
	})
}
