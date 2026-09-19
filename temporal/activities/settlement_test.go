package activities

import (
	"context"
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestGetUnsettledPayment(t *testing.T) {
	t.Run("returns unsettled completed payments", func(t *testing.T) {
		cleanTestDB(t)

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("Error creating test merchant %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("Error creating test merchant %v", err)
		}

		pendingStatus := constants.LedgerSettlementPending
		settledStatus := constants.LedgerSettlementSettled
		completedStatus := constants.TransactionStatusCompleted
		failedStatus := "failed"
		currency := "INR"

		pendingPayment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Status:             &completedStatus,
			SettlementStatus:   &pendingStatus,
			Currency:           &currency,
		}

		settledPayment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(2000),
			Status:             &completedStatus,
			SettlementStatus:   &settledStatus,
			Currency:           &currency,
		}

		failedPayment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(3000),
			Status:             &failedStatus,
			SettlementStatus:   &pendingStatus,
			Currency:           &currency,
		}

		if err := db.Create(pendingPayment).Error; err != nil {
			t.Fatalf("failed to create pending payment: %v", err)
		}

		if err := db.Create(settledPayment).Error; err != nil {
			t.Fatalf("failed to create settled payment: %v", err)
		}

		if err := db.Create(failedPayment).Error; err != nil {
			t.Fatalf("failed to create failed payment: %v", err)
		}

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		payments, err := registry.GetUnsettledPayment(context.Background())

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(payments) != 1 {
			t.Fatalf("expected 1 unsettled payment, got %d", len(payments))
		}

		if payments[0].ID != pendingPayment.ID {
			t.Errorf(
				"expected payment %v, got %v",
				pendingPayment.ID,
				payments[0].ID,
			)
		}
	})

	t.Run("returns empty result when no unsettled payments exist", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		payments, err := registry.GetUnsettledPayment(context.Background())

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(payments) != 0 {
			t.Errorf("expected 0 payments, got %d", len(payments))
		}
	})
}

func TestExecuteSettlementByMerchant(t *testing.T) {

	t.Run("returns error when payment vault has insufficient balance", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		vaultType := constants.PaymentVault
		vaultStatus := constants.VaultStatusActive

		vault := &models.Vault{
			ID:      utils.GenerateUUID(),
			Type:    vaultType,
			Balance: decimal.NewFromInt(500),
			Status:  vaultStatus,
		}

		if err := db.Create(vault).Error; err != nil {
			t.Fatalf("failed to create payment vault: %v", err)
		}

		walletStatus := "active"

		wallet := &models.Wallets{
			ID:               uuid.New(),
			MerchantID:       merchant.ID,
			AvailableBalance: decimal.NewFromInt(1000),
			ReservedBalance:  decimal.Zero,
			Status:           walletStatus,
		}

		if err := db.Create(wallet).Error; err != nil {
			t.Fatalf("failed to create wallet: %v", err)
		}

		pendingSettlement := constants.LedgerSettlementPending
		completedStatus := constants.TransactionStatusCompleted
		paymentReference := "insufficient-balance-payment"
		currency := "INR"

		payment := &models.Payment{
			ID:                 uuid.New(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: merchant.ID,
			PaymentReference:   &paymentReference,
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &completedStatus,
			SettlementStatus:   &pendingSettlement,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		transactionRef := "ledger-tx-insufficient"
		ledgerType := "payment"
		ledgerStatus := "posted"

		ledgerTransaction := &models.LedgerTransaction{
			ID:               uuid.New(),
			TransactionRef:   transactionRef,
			Type:             ledgerType,
			ReferenceID:      paymentReference,
			Status:           ledgerStatus,
			SettlementStatus: pendingSettlement,
		}

		if err := db.Create(ledgerTransaction).Error; err != nil {
			t.Fatalf("failed to create ledger transaction: %v", err)
		}

		err := registry.ExecuteSettlementByMerchant(
			context.Background(),
			merchant.ID,
			payment,
		)

		if err == nil {
			t.Fatal("expected insufficient balance error, got nil")
		}

		if err.Error() != "insufficient balance of payment vault" {
			t.Errorf(
				"expected error %q, got %q",
				"insufficient balance of payment vault",
				err.Error(),
			)
		}

		var storedVault models.Vault
		if err := db.Where("id = ?", vault.ID).First(&storedVault).Error; err != nil {
			t.Fatalf("failed to fetch vault: %v", err)
		}

		if !storedVault.Balance.Equal(decimal.NewFromInt(500)) {
			t.Errorf(
				"expected vault balance 500, got %s",
				storedVault.Balance.String(),
			)
		}

		var storedWallet models.Wallets
		if err := db.Where("id = ?", wallet.ID).First(&storedWallet).Error; err != nil {
			t.Fatalf("failed to fetch wallet: %v", err)
		}

		if !storedWallet.AvailableBalance.Equal(decimal.NewFromInt(1000)) {
			t.Errorf(
				"expected wallet balance 1000, got %s",
				storedWallet.AvailableBalance.String(),
			)
		}

		var storedPayment models.Payment
		if err := db.Where("id = ?", payment.ID).First(&storedPayment).Error; err != nil {
			t.Fatalf("failed to fetch payment: %v", err)
		}

		if storedPayment.SettlementStatus == nil ||
			*storedPayment.SettlementStatus != constants.LedgerSettlementPending {
			t.Errorf("expected payment settlement status to remain pending")
		}
	})

	t.Run("settles payment successfully", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		vaultType := constants.PaymentVault
		vaultStatus := constants.VaultStatusActive

		vault := &models.Vault{
			ID:      utils.GenerateUUID(),
			Type:    vaultType,
			Balance: decimal.NewFromInt(5000),
			Status:  vaultStatus,
		}

		if err := db.Create(vault).Error; err != nil {
			t.Fatalf("failed to create payment vault: %v", err)
		}

		walletStatus := "active"

		wallet := &models.Wallets{
			ID:               uuid.New(),
			MerchantID:       merchant.ID,
			AvailableBalance: decimal.NewFromInt(1000),
			ReservedBalance:  decimal.Zero,
			Status:           walletStatus,
		}

		if err := db.Create(wallet).Error; err != nil {
			t.Fatalf("failed to create wallet: %v", err)
		}

		paymentAmount := decimal.NewFromInt(2000)
		pendingSettlement := constants.LedgerSettlementPending
		completedStatus := constants.TransactionStatusCompleted
		paymentReference := "successful-settlement-payment"
		currency := "INR"

		payment := &models.Payment{
			ID:                 uuid.New(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: merchant.ID,
			PaymentReference:   &paymentReference,
			Amount:             paymentAmount,
			Currency:           &currency,
			Status:             &completedStatus,
			SettlementStatus:   &pendingSettlement,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		transactionRef := "ledger-tx-success"
		ledgerType := "payment"
		ledgerStatus := "posted"

		ledgerTransaction := &models.LedgerTransaction{
			ID:               uuid.New(),
			TransactionRef:   transactionRef,
			Type:             ledgerType,
			ReferenceID:      paymentReference,
			Status:           ledgerStatus,
			SettlementStatus: pendingSettlement,
		}

		if err := db.Create(ledgerTransaction).Error; err != nil {
			t.Fatalf("failed to create ledger transaction: %v", err)
		}

		err := registry.ExecuteSettlementByMerchant(
			context.Background(),
			merchant.ID,
			payment,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		var updatedVault models.Vault
		if err := db.Where("id = ?", vault.ID).First(&updatedVault).Error; err != nil {
			t.Fatalf("failed to fetch updated vault: %v", err)
		}

		expectedVaultBalance := decimal.NewFromInt(3000)

		if !updatedVault.Balance.Equal(expectedVaultBalance) {
			t.Errorf(
				"expected vault balance %s, got %s",
				expectedVaultBalance.String(),
				updatedVault.Balance.String(),
			)
		}

		var updatedWallet models.Wallets
		if err := db.Where("id = ?", wallet.ID).First(&updatedWallet).Error; err != nil {
			t.Fatalf("failed to fetch updated wallet: %v", err)
		}

		expectedWalletBalance := decimal.NewFromInt(3000)

		if !updatedWallet.AvailableBalance.Equal(expectedWalletBalance) {
			t.Errorf(
				"expected wallet balance %s, got %s",
				expectedWalletBalance.String(),
				updatedWallet.AvailableBalance.String(),
			)
		}

		var updatedPayment models.Payment
		if err := db.Where("id = ?", payment.ID).First(&updatedPayment).Error; err != nil {
			t.Fatalf("failed to fetch updated payment: %v", err)
		}

		if updatedPayment.SettlementStatus == nil ||
			*updatedPayment.SettlementStatus != constants.LedgerSettlementSettled {
			t.Errorf("expected payment settlement status to be settled")
		}

		var ledgerEntries []models.LedgerEntry

		if err := db.
			Where("ledger_transaction_id = ?", ledgerTransaction.ID).
			Find(&ledgerEntries).Error; err != nil {
			t.Fatalf("failed to fetch ledger entries: %v", err)
		}

		if len(ledgerEntries) != 2 {
			t.Fatalf(
				"expected 2 settlement ledger entries, got %d",
				len(ledgerEntries),
			)
		}

		var debitFound bool
		var creditFound bool

		for _, entry := range ledgerEntries {
			if entry.AccountType != "" &&
				entry.AccountType == constants.LedgerAccountTypeVault &&
				entry.AccountID == vault.ID &&
				entry.EntryType != "" &&
				entry.EntryType == constants.LedgerEntryTypeDebit &&
				entry.Amount.Equal(paymentAmount) {
				debitFound = true
			}

			if entry.AccountType != "" &&
				entry.AccountType == constants.LedgerAccountTypeWallet &&
				entry.AccountID == updatedWallet.ID &&
				entry.EntryType != "" &&
				entry.EntryType == constants.LedgerEntryTypeCredit &&
				entry.Amount.Equal(paymentAmount) {
				creditFound = true
			}
		}

		if !debitFound {
			t.Error("expected payment vault debit ledger entry")
		}

		if !creditFound {
			t.Error("expected merchant wallet credit ledger entry")
		}
	})
}
