package activities

import (
	"context"
	"testing"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestExecutePayment(t *testing.T) {
	activityRegistry := &Registry{
		DB:       db,
		Services: testServices,
		Repo:     testRepo,
	}
	t.Run("successful payment with sufficient wallet balance", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if errM := db.Create(merchant).Error; errM != nil {
			t.Fatalf("error creating merchant: %v", errM)
		}

		merchantID := merchant.ID

		merchantWallet := testutils.GenerateTestWallet(merchantID)
		if errW := db.Create(merchantWallet).Error; errW != nil {
			t.Fatalf("error creating wallet: %v", errW)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if errRM := db.Create(receiverMerchant).Error; errRM != nil {
			t.Fatalf("error creating merchant: %v", errRM)
		}

		receiverMerchantWallet := testutils.GenerateTestWallet(receiverMerchant.ID)
		if errRW := db.Create(receiverMerchantWallet).Error; errRW != nil {
			t.Fatalf("error creating wallet: %v", errRW)
		}

		paymentAmount := decimal.NewFromInt(1000)
		fee := decimal.NewFromInt(50)
		totalAmount := paymentAmount.Add(fee)

		paymentReference := "payment-execute-success"
		processingStatus := constants.TransactionStatusProcessing
		settledStatus := constants.LedgerSettlementSettled
		currency := "INR"

		payment := &models.Payment{
			ID:                 uuid.New(),
			SenderMerchantID:   merchantID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   &paymentReference,
			Amount:             paymentAmount,
			Currency:           &currency,
			Status:             &processingStatus,
			SettlementStatus:   &settledStatus,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		paymentVault := &models.Vault{
			ID:      uuid.New(),
			Type:    constants.PaymentVault,
			Balance: decimal.NewFromInt(20000),
			Status:  constants.VaultStatusActive,
		}

		if err := db.Create(paymentVault).Error; err != nil {
			t.Fatalf("failed to create payment vault: %v", err)
		}

		companyVault := &models.Vault{
			ID:      uuid.New(),
			Type:    constants.CompanyVault,
			Balance: decimal.NewFromInt(10000),
			Status:  constants.VaultStatusActive,
		}

		if err := db.Create(companyVault).Error; err != nil {
			t.Fatalf("failed to create company vault: %v", err)
		}

		request := &PaymentWorkflowRequest{
			PaymentReference: paymentReference,
			MerchantID:       merchantID,
			Fee:              fee,
			Request: dto.CreatePaymentRequest{
				Amount:   paymentAmount,
				Currency: "INR",
			},
		}

		result, err := activityRegistry.ExecutePayment(
			context.Background(),
			request,
		)

		if err != nil {
			t.Fatalf("expected successful payment, got error: %v", err)
		}

		if result == nil {
			t.Fatal("expected payment result, got nil")
		}

		if result.ID != payment.ID {
			t.Errorf(
				"expected payment ID %s, got %s",
				payment.ID,
				result.ID,
			)
		}

		if *result.Status != constants.TransactionStatusCompleted {
			t.Errorf(
				"expected payment status %v, got %v",
				constants.TransactionStatusCompleted,
				result.Status,
			)
		}

		var updatedWallet models.Wallets
		if err := db.
			Where("id = ?", merchantWallet.ID).
			First(&updatedWallet).Error; err != nil {
			t.Fatalf("failed to fetch wallet: %v", err)
		}

		expectedAvailableBalance := decimal.NewFromInt(3950)

		if !updatedWallet.AvailableBalance.Equal(expectedAvailableBalance) {
			t.Errorf(
				"expected wallet available balance %s, got %s",
				expectedAvailableBalance,
				updatedWallet.AvailableBalance,
			)
		}

		if !updatedWallet.ReservedBalance.Equal(decimal.Zero) {
			t.Errorf(
				"expected wallet reserved balance 0, got %s",
				updatedWallet.ReservedBalance,
			)
		}

		var updatedPaymentVault models.Vault
		if err := db.
			Where("id = ?", paymentVault.ID).
			First(&updatedPaymentVault).Error; err != nil {
			t.Fatalf("failed to fetch payment vault: %v", err)
		}

		expectedPaymentVaultBalance := decimal.NewFromInt(21000)

		if !updatedPaymentVault.Balance.Equal(expectedPaymentVaultBalance) {
			t.Errorf(
				"expected payment vault balance %s, got %s",
				expectedPaymentVaultBalance,
				updatedPaymentVault.Balance,
			)
		}

		var updatedCompanyVault models.Vault
		if err := db.
			Where("id = ?", companyVault.ID).
			First(&updatedCompanyVault).Error; err != nil {
			t.Fatalf("failed to fetch company vault: %v", err)
		}

		expectedCompanyVaultBalance := decimal.NewFromInt(10050)

		if !updatedCompanyVault.Balance.Equal(expectedCompanyVaultBalance) {
			t.Errorf(
				"expected company vault balance %s, got %s",
				expectedCompanyVaultBalance,
				updatedCompanyVault.Balance,
			)
		}

		var updatedPayment models.Payment
		if err := db.
			Where("id = ?", payment.ID).
			First(&updatedPayment).Error; err != nil {
			t.Fatalf("failed to fetch payment: %v", err)
		}

		if *updatedPayment.Status != constants.TransactionStatusCompleted {
			t.Errorf(
				"expected payment status %v, got %v",
				constants.TransactionStatusCompleted,
				updatedPayment.Status,
			)
		}

		var ledgerTransaction models.LedgerTransaction
		if err := db.
			Where("reference_id = ?", paymentReference).
			First(&ledgerTransaction).Error; err != nil {
			t.Fatalf("failed to fetch ledger transaction: %v", err)
		}

		if ledgerTransaction.Status != constants.TransactionStatusCompleted {
			t.Errorf(
				"expected ledger status %q, got %q",
				constants.TransactionStatusCompleted,
				ledgerTransaction.Status,
			)
		}

		if ledgerTransaction.SettlementStatus != constants.LedgerSettlementPending {
			t.Errorf(
				"expected settlement status %q, got %q",
				constants.LedgerSettlementPending,
				ledgerTransaction.SettlementStatus,
			)
		}

		var ledgerEntries []models.LedgerEntry
		if err := db.
			Where("ledger_transaction_id = ?", ledgerTransaction.ID).
			Find(&ledgerEntries).Error; err != nil {
			t.Fatalf("failed to fetch ledger entries: %v", err)
		}

		if len(ledgerEntries) != 4 {
			t.Fatalf(
				"expected 4 ledger entries, got %d",
				len(ledgerEntries),
			)
		}

		var walletDebit decimal.Decimal
		var paymentVaultCredit decimal.Decimal
		var paymentVaultDebit decimal.Decimal
		var companyVaultCredit decimal.Decimal

		for _, entry := range ledgerEntries {
			switch {
			case entry.AccountType == constants.LedgerAccountTypeWallet &&
				entry.AccountID == merchantWallet.ID &&
				entry.EntryType == constants.LedgerEntryTypeDebit:

				walletDebit = walletDebit.Add(entry.Amount)

			case entry.AccountType == constants.LedgerAccountTypeVault &&
				entry.AccountID == paymentVault.ID &&
				entry.EntryType == constants.LedgerEntryTypeCredit:

				paymentVaultCredit = paymentVaultCredit.Add(entry.Amount)

			case entry.AccountType == constants.LedgerAccountTypeVault &&
				entry.AccountID == paymentVault.ID &&
				entry.EntryType == constants.LedgerEntryTypeDebit:

				paymentVaultDebit = paymentVaultDebit.Add(entry.Amount)

			case entry.AccountType == constants.LedgerAccountTypeVault &&
				entry.AccountID == companyVault.ID &&
				entry.EntryType == constants.LedgerEntryTypeCredit:

				companyVaultCredit = companyVaultCredit.Add(entry.Amount)
			}
		}

		if !walletDebit.Equal(totalAmount) {
			t.Errorf(
				"expected wallet debit %s, got %s",
				totalAmount,
				walletDebit,
			)
		}

		if !paymentVaultCredit.Equal(totalAmount) {
			t.Errorf(
				"expected payment vault credit %s, got %s",
				totalAmount,
				paymentVaultCredit,
			)
		}

		if !paymentVaultDebit.Equal(fee) {
			t.Errorf(
				"expected payment vault fee debit %s, got %s",
				fee,
				paymentVaultDebit,
			)
		}

		if !companyVaultCredit.Equal(fee) {
			t.Errorf(
				"expected company vault fee credit %s, got %s",
				fee,
				companyVaultCredit,
			)
		}
	})
}

func TestValidatePaymentRequest(t *testing.T) {
	activityRegistry := &Registry{
		DB:       db,
		Services: testServices,
		Repo:     testRepo,
	}
	t.Run("request is nil", func(t *testing.T) {
		err := testServices.ValidatePaymentRequest(
			nil,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "payment request is required" {
			t.Errorf("expected error %q, got %q",
				"payment request is required",
				(err).Error(),
			)
		}
	})

	t.Run("sender merchant id is nil", func(t *testing.T) {
		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: uuid.New(),
			Amount:             decimal.NewFromInt(100),
			Currency:           "INR",
		}

		err := testServices.ValidatePaymentRequest(
			request,
			uuid.Nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "sender merchant id is required" {
			t.Errorf("expected error %q, got %q",
				"sender merchant id is required",
				(err).Error(),
			)
		}
	})

	t.Run("receiver merchant id is nil", func(t *testing.T) {
		request := &dto.CreatePaymentRequest{
			Amount:   decimal.NewFromInt(100),
			Currency: "INR",
		}

		err := testServices.ValidatePaymentRequest(
			request,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "receiver merchant id is required" {
			t.Errorf("expected error %q, got %q",
				"receiver merchant id is required",
				(err).Error(),
			)
		}
	})

	t.Run("sender and receiver are same", func(t *testing.T) {
		merchantID := uuid.New()

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: merchantID,
			Amount:             decimal.NewFromInt(100),
			Currency:           "INR",
		}

		err := testServices.ValidatePaymentRequest(
			request,
			merchantID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "sender and receiver can't be same" {
			t.Errorf("expected error %q, got %q",
				"sender and receiver can't be same",
				(err).Error(),
			)
		}
	})

	t.Run("amount is zero", func(t *testing.T) {
		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: uuid.New(),
			Amount:             decimal.Zero,
			Currency:           "INR",
		}

		err := testServices.ValidatePaymentRequest(
			request,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "amount must be greater than 0" {
			t.Errorf("expected error %q, got %q",
				"amount must be greater than 0",
				(err).Error(),
			)
		}
	})

	t.Run("amount is negative", func(t *testing.T) {
		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: uuid.New(),
			Amount:             decimal.NewFromInt(-100),
			Currency:           "INR",
		}

		err := testServices.ValidatePaymentRequest(
			request,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "amount must be greater than 0" {
			t.Errorf("expected error %q, got %q",
				"amount must be greater than 0",
				(err).Error(),
			)
		}
	})

	t.Run("currency is not INR", func(t *testing.T) {
		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: uuid.New(),
			Amount:             decimal.NewFromInt(100),
			Currency:           "USD",
		}

		err := testServices.ValidatePaymentRequest(
			request,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "currency should be INR" {
			t.Errorf("expected error %q, got %q",
				"currency should be INR",
				(err).Error(),
			)
		}
	})

	t.Run("sender merchant does not exist", func(t *testing.T) {
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create receiver merchant: %v", err)
		}

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(100),
			Currency:           "INR",
		}

		err := testServices.ValidatePaymentRequest(
			request,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "sender does not exist" {
			t.Errorf("expected error %q, got %q",
				"sender does not exist",
				(err).Error(),
			)
		}
	})

	t.Run("sender wallet does not exist", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("failed to create sender merchant: %v", err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create receiver merchant: %v", err)
		}

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(100),
			Currency:           "INR",
		}

		err := testServices.ValidatePaymentRequest(
			request,
			senderMerchant.ID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "sender wallet not found" {
			t.Errorf("expected error %q, got %q",
				"sender wallet not found",
				(err).Error(),
			)
		}
	})

	t.Run("receiver merchant does not exist", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		senderWallet := testutils.GenerateTestWallet(senderMerchant.ID)

		senderWallet.MerchantID = senderMerchant.ID

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("failed to create sender merchant: %v", err)
		}

		if err := db.Create(senderWallet).Error; err != nil {
			t.Fatalf("failed to create sender wallet: %v", err)
		}

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: uuid.New(),
			Amount:             decimal.NewFromInt(100),
			Currency:           "INR",
		}

		err := testServices.ValidatePaymentRequest(
			request,
			senderMerchant.ID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "receiver does not exist" {
			t.Errorf("expected error %q, got %q",
				"receiver does not exist",
				(err).Error(),
			)
		}
	})

	t.Run("receiver wallet does not exist", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		senderWallet := testutils.GenerateTestWallet(senderMerchant.ID)
		senderWallet.MerchantID = senderMerchant.ID

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("failed to create sender merchant: %v", err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create receiver merchant: %v", err)
		}

		if err := db.Create(senderWallet).Error; err != nil {
			t.Fatalf("failed to create sender wallet: %v", err)
		}

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(100),
			Currency:           "INR",
		}

		err := testServices.ValidatePaymentRequest(
			request,
			senderMerchant.ID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "receiver wallet does not exist" {
			t.Errorf("expected error %q, got %q",
				"receiver wallet does not exist",
				(err).Error(),
			)
		}
	})

	t.Run("valid payment request", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		senderWallet := testutils.GenerateTestWallet(senderMerchant.ID)
		senderWallet.MerchantID = senderMerchant.ID

		receiverWallet := testutils.GenerateTestWallet(receiverMerchant.ID)
		receiverWallet.MerchantID = receiverMerchant.ID

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("failed to create sender merchant: %v", err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create receiver merchant: %v", err)
		}

		if err := db.Create(senderWallet).Error; err != nil {
			t.Fatalf("failed to create sender wallet: %v", err)
		}

		if err := db.Create(receiverWallet).Error; err != nil {
			t.Fatalf("failed to create receiver wallet: %v", err)
		}

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(100),
			Currency:           "INR",
		}

		err := activityRegistry.ValidatePaymentRequest(
			context.Background(),
			request,
			senderMerchant.ID,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestCalculateFees(t *testing.T) {
	activityRegistry := &Registry{
		DB:       db,
		Services: testServices,
		Repo:     testRepo,
	}
	t.Run("zero amount", func(t *testing.T) {
		amount := decimal.Zero

		fee, err := testServices.CalculateFees(amount)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "amount must be greater than 0" {
			t.Errorf(
				"expected error %q, got %q",
				"amount must be greater than 0",
				(err).Error(),
			)
		}

		if !fee.Equal(decimal.Zero) {
			t.Errorf("expected fee 0, got %s", fee.String())
		}
	})

	t.Run("negative amount", func(t *testing.T) {
		amount := decimal.NewFromInt(-100)

		fee, err := testServices.CalculateFees(amount)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "amount must be greater than 0" {
			t.Errorf(
				"expected error %q, got %q",
				"amount must be greater than 0",
				(err).Error(),
			)
		}

		if !fee.Equal(decimal.Zero) {
			t.Errorf("expected fee 0, got %s", fee.String())
		}
	})

	t.Run("calculates 1 percent fee", func(t *testing.T) {
		amount := decimal.NewFromInt(1000)

		fee, err := testServices.CalculateFees(amount)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedFee := decimal.NewFromInt(10)

		if !fee.Equal(expectedFee) {
			t.Errorf(
				"expected fee %s, got %s",
				expectedFee.String(),
				fee.String(),
			)
		}
	})

	t.Run("calculates fee for decimal amount", func(t *testing.T) {
		amount := decimal.NewFromFloat(1234.56)

		fee, err := activityRegistry.CalculateFees(amount)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedFee := decimal.NewFromFloat(12.3456)

		if !fee.Equal(expectedFee) {
			t.Errorf(
				"expected fee %s, got %s",
				expectedFee.String(),
				fee.String(),
			)
		}
	})
}

func TestCreatePayment(t *testing.T) {
	activityRegistry := &Registry{
		DB:       db,
		Services: testServices,
		Repo:     testRepo,
	}
	t.Run("creates payment successfully", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("failed to create sender merchant: %v", err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create receiver merchant: %v", err)
		}

		status := constants.LedgerSettlementPending

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "test payment",
		}

		payment, err := activityRegistry.CreatePayment(
			context.Background(),
			request,
			senderMerchant.ID,
			&status,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if payment == nil {
			t.Fatal("expected payment, got nil")
		}

		if payment.ID == uuid.Nil {
			t.Error("expected payment ID to be generated")
		}

		if payment.PaymentReference == nil {
			t.Fatal("expected payment reference to be generated")
		}

		if *payment.PaymentReference == "" {
			t.Error("expected payment reference to be non-empty")
		}

		if payment.CustomerReference == nil {
			t.Fatal("expected customer reference to be generated")
		}

		if *payment.CustomerReference == "" {
			t.Error("expected customer reference to be non-empty")
		}

		if payment.SenderMerchantID == uuid.Nil {
			t.Fatal("expected sender merchant ID to be set")
		}

		if payment.SenderMerchantID != senderMerchant.ID {
			t.Errorf(
				"expected sender merchant ID %v, got %v",
				senderMerchant.ID,
				payment.SenderMerchantID,
			)
		}

		if payment.ReceiverMerchantID != receiverMerchant.ID {
			t.Errorf(
				"expected receiver merchant ID %v, got %v",
				receiverMerchant.ID,
				payment.ReceiverMerchantID,
			)
		}

		if !payment.Amount.Equal(request.Amount) {
			t.Errorf(
				"expected amount %s, got %s",
				request.Amount.String(),
				payment.Amount.String(),
			)
		}

		if payment.Currency == nil {
			t.Fatal("expected currency to be set")
		}

		if *payment.Currency != request.Currency {
			t.Errorf(
				"expected currency %q, got %q",
				request.Currency,
				*payment.Currency,
			)
		}

		if payment.Description == nil {
			t.Fatal("expected description to be set")
		}

		if *payment.Description != request.Description {
			t.Errorf(
				"expected description %q, got %q",
				request.Description,
				*payment.Description,
			)
		}

		if payment.Status == nil {
			t.Fatal("expected status to be set")
		}

		if *payment.Status != status {
			t.Errorf(
				"expected status %q, got %q",
				status,
				*payment.Status,
			)
		}

		if payment.SettlementStatus == nil {
			t.Fatal("expected settlement status to be set")
		}

		if *payment.SettlementStatus != constants.LedgerSettlementPending {
			t.Errorf(
				"expected settlement status %q, got %q",
				constants.LedgerSettlementPending,
				*payment.SettlementStatus,
			)
		}

		if payment.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}

		if payment.UpdatedAt.IsZero() {
			t.Error("expected UpdatedAt to be set")
		}

		var storedPayment models.Payment

		if err := db.Where("id = ?", payment.ID).First(&storedPayment).Error; err != nil {
			t.Fatalf("failed to fetch created payment: %v", err)
		}

		if storedPayment.ID != payment.ID {
			t.Errorf(
				"expected stored payment ID %v, got %v",
				payment.ID,
				storedPayment.ID,
			)
		}

		if storedPayment.ReceiverMerchantID != receiverMerchant.ID {
			t.Errorf(
				"expected stored receiver merchant ID %v, got %v",
				receiverMerchant.ID,
				storedPayment.ReceiverMerchantID,
			)
		}

		if !storedPayment.Amount.Equal(request.Amount) {
			t.Errorf(
				"expected stored amount %s, got %s",
				request.Amount.String(),
				storedPayment.Amount.String(),
			)
		}
	})

	t.Run("fails when request is nil", func(t *testing.T) {
		merchantID := uuid.New()
		status := constants.TransactionStatusPending

		// The repository dereferences request directly, so this currently
		// panics instead of returning an error.
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when request is nil")
			}
		}()

		_, _ = activityRegistry.CreatePayment(
			context.Background(),
			nil,
			merchantID,
			&status,
		)
	})

	t.Run("fails when database rejects payment", func(t *testing.T) {
		// The repository requires a valid merchant relationship.
		// Using a non-existent sender merchant should make the DB reject
		// the insert when the foreign-key constraint is enforced.
		merchantID := uuid.New()
		receiverMerchantID := uuid.New()

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchantID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "invalid payment",
		}

		status := constants.TransactionStatusPending

		payment, err := activityRegistry.CreatePayment(
			context.Background(),
			request,
			merchantID,
			&status,
		)

		if err == nil {
			// Some test DB configurations may not enforce foreign keys.
			// In that case, don't claim this repository-level branch is
			// covered by this test.
			if payment == nil {
				t.Fatal("expected payment or database error")
			}

			t.Skip("test database does not enforce the expected foreign-key constraint")
		}

		if payment != nil {
			t.Error("expected nil payment when database insert fails")
		}
	})
}

func TestUpdatePaymentStatus(t *testing.T) {
	activityRegistry := &Registry{
		DB:       db,
		Services: testServices,
		Repo:     testRepo,
	}
	t.Run("updates payment status successfully", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("failed to create sender merchant: %v", err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create receiver merchant: %v", err)
		}

		initialStatus := constants.TransactionStatusPending

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "test payment",
		}

		payment, err := activityRegistry.CreatePayment(
			context.Background(),
			request,
			senderMerchant.ID,
			&initialStatus,
		)

		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		if payment == nil {
			t.Fatal("expected payment, got nil")
		}

		newStatus := constants.TransactionStatusProcessing

		updatedPayment, err := testRepo.UpdatePaymentStatus(
			*payment.PaymentReference,
			&newStatus,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if updatedPayment == nil {
			t.Fatal("expected updated payment, got nil")
		}

		if updatedPayment.ID != payment.ID {
			t.Errorf(
				"expected payment ID %v, got %v",
				payment.ID,
				updatedPayment.ID,
			)
		}

		if updatedPayment.Status == nil {
			t.Fatal("expected payment status to be set")
		}

		if *updatedPayment.Status != newStatus {
			t.Errorf(
				"expected status %q, got %q",
				newStatus,
				*updatedPayment.Status,
			)
		}

		var storedPayment models.Payment

		if err := db.
			Where("id = ?", payment.ID).
			First(&storedPayment).Error; err != nil {
			t.Fatalf("failed to fetch updated payment: %v", err)
		}

		if storedPayment.Status == nil {
			t.Fatal("expected stored payment status to be set")
		}

		if *storedPayment.Status != newStatus {
			t.Errorf(
				"expected stored status %q, got %q",
				newStatus,
				*storedPayment.Status,
			)
		}
	})

	t.Run("returns payment without updating when status is already same", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("failed to create sender merchant: %v", err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create receiver merchant: %v", err)
		}

		status := constants.TransactionStatusProcessing

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(500),
			Currency:           "INR",
			Description:        "same status test",
		}

		payment, err := activityRegistry.CreatePayment(
			context.Background(),
			request,
			senderMerchant.ID,
			&status,
		)

		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		updatedPayment, err := testRepo.UpdatePaymentStatus(
			*payment.PaymentReference,
			&status,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if updatedPayment == nil {
			t.Fatal("expected payment, got nil")
		}

		if updatedPayment.ID != payment.ID {
			t.Errorf(
				"expected payment ID %v, got %v",
				payment.ID,
				updatedPayment.ID,
			)
		}

		if updatedPayment.Status == nil {
			t.Fatal("expected payment status to be set")
		}

		if *updatedPayment.Status != status {
			t.Errorf(
				"expected status %q, got %q",
				status,
				*updatedPayment.Status,
			)
		}
	})

	t.Run("payment does not exist", func(t *testing.T) {
		paymentReference := "payment-reference-does-not-exist"
		status := constants.TransactionStatusProcessing

		payment, err := activityRegistry.UpdatePaymentStatus(
			context.Background(),
			&status,
			paymentReference,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if payment != nil {
			t.Error("expected nil payment when payment does not exist")
		}
	})

	t.Run("nil payment reference", func(t *testing.T) {
		status := constants.TransactionStatusProcessing

		payment, err := activityRegistry.UpdatePaymentStatus(
			context.Background(),
			&status,
			"",
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if payment != nil {
			t.Error("expected nil payment, got non-nil")
		}
	})
}
