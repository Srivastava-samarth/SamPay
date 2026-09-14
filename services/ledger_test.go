package services

import (
	"testing"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestGetTransactions(t *testing.T) {
	ls := &LedgerService{}

	validAccountType := "wallet"
	validAccountID := uuid.New()

	tests := []struct {
		name          string
		accountType   *string
		accountID     uuid.UUID
		cursor        *uuid.UUID
		direction     string
		expectedError string
	}{
		{
			name:          "nil account type",
			accountType:   nil,
			accountID:     validAccountID,
			direction:     "next",
			expectedError: "account type is empty",
		},
		{
			name: "empty account type",
			accountType: func() *string {
				value := ""
				return &value
			}(),
			accountID:     validAccountID,
			direction:     "next",
			expectedError: "account type is empty",
		},
		{
			name:          "nil account id",
			accountType:   &validAccountType,
			accountID:     uuid.Nil,
			direction:     "next",
			expectedError: "account ID is empty",
		},
		{
			name:        "nil cursor uuid",
			accountType: &validAccountType,
			accountID:   validAccountID,
			cursor: func() *uuid.UUID {
				value := uuid.Nil
				return &value
			}(),
			direction:     "next",
			expectedError: "invalid cursor",
		},
		{
			name:          "invalid pagination direction",
			accountType:   &validAccountType,
			accountID:     validAccountID,
			direction:     "invalid",
			expectedError: "invalid pagination direction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ls.GetTransactions(
				tt.accountType,
				tt.accountID,
				tt.cursor,
				tt.direction,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Fatalf(
					"expected %q, got %q",
					tt.expectedError,
					err.Error(),
				)
			}
		})
	}
}

func TestGetTransaction(t *testing.T) {
	ls := &LedgerService{}

	validWalletID := uuid.New()
	validTransactionID := uuid.New()

	tests := []struct {
		name          string
		walletID      uuid.UUID
		transactionID uuid.UUID
		expectedError string
	}{
		{
			name:          "nil wallet id",
			walletID:      uuid.Nil,
			transactionID: validTransactionID,
			expectedError: "WalletId missing",
		},
		{
			name:          "nil transaction id",
			walletID:      validWalletID,
			transactionID: uuid.Nil,
			expectedError: "TransactionId missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ls.GetTransaction(
				tt.walletID,
				tt.transactionID,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Fatalf(
					"expected %q, got %q",
					tt.expectedError,
					err.Error(),
				)
			}
		})
	}
}

func TestCreateLedgerEntries(t *testing.T) {
	ls := &LedgerService{}

	validTransactionID := uuid.New()

	tests := []struct {
		name          string
		tx            *gorm.DB
		entries       []*models.LedgerEntry
		expectedError string
	}{
		{
			name:          "empty entries",
			tx:            &gorm.DB{},
			entries:       []*models.LedgerEntry{},
			expectedError: "no entries were provided",
		},
		{
			name: "nil transaction",
			entries: []*models.LedgerEntry{
				{
					LedgerTransactionID: validTransactionID,
				},
			},
			expectedError: "transaction cannot be nil",
		},
		{
			name: "nil ledger entry",
			tx:   &gorm.DB{},
			entries: []*models.LedgerEntry{
				nil,
			},
			expectedError: "ledger entry cannot be nil",
		},
		{
			name: "missing ledger transaction id",
			tx:   &gorm.DB{},
			entries: []*models.LedgerEntry{
				{},
			},
			expectedError: "ledger transaction ID is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ls.CreateLedgerEntries(
				tt.tx,
				tt.entries,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Fatalf(
					"expected %q, got %q",
					tt.expectedError,
					err.Error(),
				)
			}
		})
	}
}

func TestCreateLedgerTransaction(t *testing.T) {
	ls := &LedgerService{}

	_, err := ls.CreateLedgerTransaction(nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "transaction not found" {
		t.Fatalf(
			"expected %q, got %q",
			"transaction not found",
			err.Error(),
		)
	}
}

func TestPostTransaction(t *testing.T) {
	ls := &LedgerService{}

	validTx := &gorm.DB{}

	validRequest := &dto.PostLedgerTransactionRequest{
		ReferenceID: "payment-ref-123",
		Currency:    "INR",
	}

	validEntry := &models.LedgerEntry{}

	tests := []struct {
		name          string
		tx            *gorm.DB
		request       *dto.PostLedgerTransactionRequest
		entries       []*models.LedgerEntry
		expectedError string
	}{
		{
			name:          "nil transaction",
			tx:            nil,
			request:       validRequest,
			entries:       []*models.LedgerEntry{validEntry},
			expectedError: "transaction cannot be nil",
		},
		{
			name:          "nil request",
			tx:            validTx,
			request:       nil,
			entries:       []*models.LedgerEntry{validEntry},
			expectedError: "transaction request cannot be nil",
		},
		{
			name: "empty reference id",
			tx:   validTx,
			request: &dto.PostLedgerTransactionRequest{
				ReferenceID: "",
				Currency:    "INR",
			},
			entries:       []*models.LedgerEntry{validEntry},
			expectedError: "reference ID is required",
		},
		{
			name: "invalid currency",
			tx:   validTx,
			request: &dto.PostLedgerTransactionRequest{
				ReferenceID: "payment-ref-123",
				Currency:    "USD",
			},
			entries:       []*models.LedgerEntry{validEntry},
			expectedError: "currency must be INR",
		},
		{
			name:          "empty entries",
			tx:            validTx,
			request:       validRequest,
			entries:       []*models.LedgerEntry{},
			expectedError: "no ledger entries were provided",
		},
		{
			name:    "nil ledger entry",
			tx:      validTx,
			request: validRequest,
			entries: []*models.LedgerEntry{
				nil,
			},
			expectedError: "ledger entry cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ls.PostTransaction(
				tt.tx,
				tt.request,
				tt.entries,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Fatalf(
					"expected %q, got %q",
					tt.expectedError,
					err.Error(),
				)
			}
		})
	}
}

func TestGetLedgerTransactionByReferenceID(t *testing.T) {
	ls := &LedgerService{}

	emptyReferenceID := ""

	tests := []struct {
		name          string
		referenceID   string
		expectedError string
	}{
		{
			name:          "empty reference id",
			referenceID:   emptyReferenceID,
			expectedError: "reference id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ls.GetLedgerTransactionByReferenceID(
				tt.referenceID,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Fatalf(
					"expected %q, got %q",
					"reference id is required",
					err.Error(),
				)
			}
		})
	}
}

func TestReverseTransactions(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()

	transactions := []dto.LedgerTransactionRow{
		{
			LedgerTransactionID: id1,
		},
		{
			LedgerTransactionID: id2,
		},
		{
			LedgerTransactionID: id3,
		},
	}

	reverseTransactions(transactions)

	expected := []uuid.UUID{
		id3,
		id2,
		id1,
	}

	for i, transaction := range transactions {
		if transaction.LedgerTransactionID != expected[i] {
			t.Fatalf(
				"index %d: expected %v, got %v",
				i,
				expected[i],
				transaction.LedgerTransactionID,
			)
		}
	}
}
