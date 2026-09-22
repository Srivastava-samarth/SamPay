package activities

import (
	"context"
	"testing"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestValidatePayoutWalletToBankRequest(t *testing.T) {
	activityRegistry := &Registry{
		Services: testServices,
		Repo:     testRepo,
	}

	t.Run("nil request", func(t *testing.T) {
		err := activityRegistry.ValidatePayoutWalletToBankRequest(
			context.Background(),
			nil,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "payment request is required" {
			t.Errorf(
				"expected error %q, got %q",
				"payment request is required",
				err.Error(),
			)
		}
	})

	t.Run("nil merchant id", func(t *testing.T) {
		request := &dto.CreateWalletToBankRequest{
			Amount:   decimal.NewFromInt(100),
			Currency: "INR",
		}

		err := activityRegistry.ValidatePayoutWalletToBankRequest(
			context.Background(),
			request,
			uuid.Nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "sender merchant id is required" {
			t.Errorf(
				"expected error %q, got %q",
				"sender merchant id is required",
				err.Error(),
			)
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		request := &dto.CreateWalletToBankRequest{
			Amount:   decimal.Zero,
			Currency: "INR",
		}

		err := activityRegistry.ValidatePayoutWalletToBankRequest(
			context.Background(),
			request,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "amount must be greater than 0" {
			t.Errorf(
				"expected error %q, got %q",
				"amount must be greater than 0",
				err.Error(),
			)
		}
	})

	t.Run("negative amount", func(t *testing.T) {
		request := &dto.CreateWalletToBankRequest{
			Amount:   decimal.NewFromInt(-100),
			Currency: "INR",
		}

		err := activityRegistry.ValidatePayoutWalletToBankRequest(
			context.Background(),
			request,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "amount must be greater than 0" {
			t.Errorf(
				"expected error %q, got %q",
				"amount must be greater than 0",
				err.Error(),
			)
		}
	})

	t.Run("invalid currency", func(t *testing.T) {
		request := &dto.CreateWalletToBankRequest{
			Amount:   decimal.NewFromInt(100),
			Currency: "USD",
		}

		err := activityRegistry.ValidatePayoutWalletToBankRequest(
			context.Background(),
			request,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "currency should be INR" {
			t.Errorf(
				"expected error %q, got %q",
				"currency should be INR",
				err.Error(),
			)
		}
	})

	t.Run("sender wallet not found", func(t *testing.T) {
		merchantID := uuid.New()

		request := &dto.CreateWalletToBankRequest{
			Amount:   decimal.NewFromInt(100),
			Currency: "INR",
		}

		err := activityRegistry.ValidatePayoutWalletToBankRequest(
			context.Background(),
			request,
			merchantID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("sender wallet id is nil", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		wallet := &models.Wallets{
			ID:               uuid.New(),
			MerchantID:       merchant.ID,
			AvailableBalance: decimal.NewFromInt(1000),
			ReservedBalance:  decimal.Zero,
			Status:           "active",
		}

		if err := db.Create(wallet).Error; err != nil {
			t.Fatalf("failed to create wallet: %v", err)
		}

		bankAccount := testutils.GenerateTestBankAccount()
		if err := db.Create(bankAccount).Error; err != nil {
			t.Fatalf("error creating source bank account: %v", err)
		}

		request := &dto.CreateWalletToBankRequest{
			Amount:                   decimal.NewFromInt(100),
			Currency:                 "INR",
			SenderWalletID:           uuid.Nil,
			DestinationBankAccountID: bankAccount.ID,
		}

		err := activityRegistry.ValidatePayoutWalletToBankRequest(
			context.Background(),
			request,
			wallet.MerchantID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "sender_wallet_id issue" {
			t.Errorf(
				"expected error %q, got %q",
				"sender_wallet_id issue",
				err.Error(),
			)
		}
	})

	t.Run("sender wallet id does not match", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		wallet := &models.Wallets{
			ID:               uuid.New(),
			MerchantID:       merchant.ID,
			AvailableBalance: decimal.NewFromInt(1000),
			ReservedBalance:  decimal.Zero,
			Status:           "active",
		}

		if err := db.Create(wallet).Error; err != nil {
			t.Fatalf("failed to create wallet: %v", err)
		}

		bankAccount := testutils.GenerateTestBankAccount()
		if err := db.Create(bankAccount).Error; err != nil {
			t.Fatalf("error creating source bank account: %v", err)
		}

		request := &dto.CreateWalletToBankRequest{
			Amount:                   decimal.NewFromInt(100),
			Currency:                 "INR",
			SenderWalletID:           uuid.New(),
			DestinationBankAccountID: bankAccount.ID,
		}

		err := activityRegistry.ValidatePayoutWalletToBankRequest(
			context.Background(),
			request,
			wallet.MerchantID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "sender_wallet_id issue" {
			t.Errorf(
				"expected error %q, got %q",
				"sender_wallet_id issue",
				err.Error(),
			)
		}
	})

	t.Run("destination bank account not found", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		wallet := &models.Wallets{
			ID:               uuid.New(),
			MerchantID:       merchant.ID,
			AvailableBalance: decimal.NewFromInt(1000),
			ReservedBalance:  decimal.Zero,
			Status:           "active",
		}

		if err := db.Create(wallet).Error; err != nil {
			t.Fatalf("failed to create wallet: %v", err)
		}

		request := &dto.CreateWalletToBankRequest{
			Amount:                   decimal.NewFromInt(100),
			Currency:                 "INR",
			SenderWalletID:           wallet.ID,
			DestinationBankAccountID: uuid.New(),
		}

		err := activityRegistry.ValidatePayoutWalletToBankRequest(
			context.Background(),
			request,
			wallet.MerchantID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("valid request", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		wallet := &models.Wallets{
			ID:               uuid.New(),
			MerchantID:       merchant.ID,
			AvailableBalance: decimal.NewFromInt(1000),
			ReservedBalance:  decimal.Zero,
			Status:           "active",
		}

		if err := db.Create(wallet).Error; err != nil {
			t.Fatalf("failed to create wallet: %v", err)
		}

		bankAccount := testutils.GenerateTestBankAccount()
		if err := db.Create(bankAccount).Error; err != nil {
			t.Fatalf("error creating source bank account: %v", err)
		}

		request := &dto.CreateWalletToBankRequest{
			Amount:                   decimal.NewFromInt(100),
			Currency:                 "INR",
			SenderWalletID:           wallet.ID,
			DestinationBankAccountID: bankAccount.ID,
		}

		err := activityRegistry.ValidatePayoutWalletToBankRequest(
			context.Background(),
			request,
			wallet.MerchantID,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestValidatePayoutBankToBankRequest(t *testing.T) {
	activityRegistry := &Registry{
		Services: testServices,
		Repo:     testRepo,
	}

	t.Run("nil request", func(t *testing.T) {
		err := activityRegistry.ValidatePayoutBankToBankRequest(
			context.Background(),
			nil,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "payment request is required" {
			t.Errorf(
				"expected error %q, got %q",
				"payment request is required",
				err.Error(),
			)
		}
	})

	t.Run("nil merchant id", func(t *testing.T) {
		request := &dto.CreateBankToBankRequest{
			Amount:   decimal.NewFromInt(100),
			Currency: "INR",
		}

		err := activityRegistry.ValidatePayoutBankToBankRequest(
			context.Background(),
			request,
			uuid.Nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "sender merchant id is required" {
			t.Errorf(
				"expected error %q, got %q",
				"sender merchant id is required",
				err.Error(),
			)
		}
	})

	t.Run("invalid amount", func(t *testing.T) {
		request := &dto.CreateBankToBankRequest{
			Amount:   decimal.Zero,
			Currency: "INR",
		}

		err := activityRegistry.ValidatePayoutBankToBankRequest(
			context.Background(),
			request,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "amount must be greater than 0" {
			t.Errorf(
				"expected error %q, got %q",
				"amount must be greater than 0",
				err.Error(),
			)
		}
	})

	t.Run("invalid currency", func(t *testing.T) {
		request := &dto.CreateBankToBankRequest{
			Amount:   decimal.NewFromInt(100),
			Currency: "USD",
		}

		err := activityRegistry.ValidatePayoutBankToBankRequest(
			context.Background(),
			request,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "currency should be INR" {
			t.Errorf(
				"expected error %q, got %q",
				"currency should be INR",
				err.Error(),
			)
		}
	})

	t.Run("source bank account not found", func(t *testing.T) {
		request := &dto.CreateBankToBankRequest{
			Amount:                   decimal.NewFromInt(100),
			Currency:                 "INR",
			SourceBankAccountID:      uuid.New(),
			DestinationBankAccountID: uuid.New(),
		}

		err := activityRegistry.ValidatePayoutBankToBankRequest(
			context.Background(),
			request,
			uuid.New(),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("destination bank account not found", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}
		merchantID := merchant.ID
		sourceBankAccount := testutils.GenerateTestBankAccount()
		if err := db.Create(sourceBankAccount).Error; err != nil {
			t.Fatalf("error creating source bank account: %v", err)
		}

		linked := &models.LinkedBankAccount{
			ID:            uuid.New(),
			BankAccountID: sourceBankAccount.ID,
			MerchantID:    merchantID,
			Status:        "active",
			Type:          sourceBankAccount.AccountType,
		}

		if err := db.Create(linked).Error; err != nil {
			t.Fatalf("failed to create linked bank account: %v", err)
		}

		request := &dto.CreateBankToBankRequest{
			Amount:                   decimal.NewFromInt(100),
			Currency:                 "INR",
			SourceBankAccountID:      sourceBankAccount.ID,
			DestinationBankAccountID: uuid.New(),
		}

		err := activityRegistry.ValidatePayoutBankToBankRequest(
			context.Background(),
			request,
			merchantID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("source and destination accounts are same", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}
		merchantID := merchant.ID
		sourceBankAccount := testutils.GenerateTestBankAccount()
		if err := db.Create(sourceBankAccount).Error; err != nil {
			t.Fatalf("error creating source bank account: %v", err)
		}

		linked := &models.LinkedBankAccount{
			ID:            uuid.New(),
			BankAccountID: sourceBankAccount.ID,
			MerchantID:    merchantID,
			Status:        "active",
			Type:          sourceBankAccount.AccountType,
		}

		if err := db.Create(linked).Error; err != nil {
			t.Fatalf("failed to create linked bank account: %v", err)
		}

		request := &dto.CreateBankToBankRequest{
			Amount:                   decimal.NewFromInt(100),
			Currency:                 "INR",
			SourceBankAccountID:      sourceBankAccount.ID,
			DestinationBankAccountID: sourceBankAccount.ID,
		}

		err := activityRegistry.ValidatePayoutBankToBankRequest(
			context.Background(),
			request,
			merchantID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "accounts must be differnt" {
			t.Errorf(
				"expected error %q, got %q",
				"accounts must be differnt",
				err.Error(),
			)
		}
	})

	t.Run("valid request", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}
		merchantID := merchant.ID
		sourceBankAccount := testutils.GenerateTestBankAccount()
		if err := db.Create(sourceBankAccount).Error; err != nil {
			t.Fatalf("error creating source bank account: %v", err)
		}

		destinationBankAccount := testutils.GenerateTestBankAccount()
		if err := db.Create(destinationBankAccount).Error; err != nil {
			t.Fatalf("error creating source bank account: %v", err)
		}

		linked := &models.LinkedBankAccount{
			ID:            uuid.New(),
			BankAccountID: sourceBankAccount.ID,
			MerchantID:    merchantID,
			Status:        "active",
			Type:          sourceBankAccount.AccountType,
		}

		if err := db.Create(linked).Error; err != nil {
			t.Fatalf("failed to create linked bank account: %v", err)
		}

		request := &dto.CreateBankToBankRequest{
			Amount:                   decimal.NewFromInt(100),
			Currency:                 "INR",
			SourceBankAccountID:      sourceBankAccount.ID,
			DestinationBankAccountID: destinationBankAccount.ID,
		}

		err := activityRegistry.ValidatePayoutBankToBankRequest(
			context.Background(),
			request,
			merchantID,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCreatePayoutWalletToBank(t *testing.T) {
	activityRegistry := &Registry{
		Repo: testRepo,
	}

	t.Run("creates payout successfully", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		wallet := testutils.GenerateTestWallet(merchant.ID)
		if errW := db.Create(wallet).Error; errW != nil {
			t.Fatalf("error creating wallet: %v", errW)
		}

		destinationBankAccount := testutils.GenerateTestBankAccount()
		if errB := db.Create(destinationBankAccount).Error; errB != nil {
			t.Fatalf("error creating bank account: %v", errB)
		}

		linkedBankAccount := &models.LinkedBankAccount{
			ID:            utils.GenerateUUID(),
			MerchantID:    merchant.ID,
			BankAccountID: destinationBankAccount.ID,
			Status:        constants.MerchantStatusActive,
			Type:          destinationBankAccount.AccountType,
		}

		if errL := db.Create(linkedBankAccount).Error; errL != nil {
			t.Fatalf("error creating linked bank acccount : %v", errL)
		}

		merchantID := merchant.ID
		walletID := wallet.ID
		destinationBankAccountID := destinationBankAccount.ID

		amount := decimal.NewFromInt(1500)
		currency := "INR"
		description := "Wallet to bank payout"
		status := constants.TransactionStatusProcessing

		request := &dto.CreateWalletToBankRequest{
			Amount:                   amount,
			Currency:                 currency,
			SenderWalletID:           walletID,
			DestinationBankAccountID: destinationBankAccountID,
			Description:              &description,
		}

		beforeCreate := time.Now()

		payout, err := activityRegistry.CreatePayoutWalletToBank(
			context.Background(),
			request,
			merchantID,
			status,
		)

		afterCreate := time.Now()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if payout == nil {
			t.Fatal("expected payout, got nil")
		}

		if payout.ID == uuid.Nil {
			t.Error("expected generated payout ID, got nil UUID")
		}

		if payout.MerchantID != merchantID {
			t.Errorf(
				"expected merchant ID %v, got %v",
				merchantID,
				payout.MerchantID,
			)
		}

		if payout.SourceWalletID == nil {
			t.Fatal("expected source wallet ID, got nil")
		}

		if *payout.SourceWalletID != walletID {
			t.Errorf(
				"expected source wallet ID %v, got %v",
				walletID,
				*payout.SourceWalletID,
			)
		}

		if payout.SourceBankAccountID != nil {
			t.Errorf(
				"expected source bank account ID to be nil, got %v",
				*payout.SourceBankAccountID,
			)
		}

		if payout.DestinationBankAccountID != destinationBankAccountID {
			t.Errorf(
				"expected destination bank account ID %v, got %v",
				destinationBankAccountID,
				payout.DestinationBankAccountID,
			)
		}

		if payout.Amount.Equal(amount) == false {
			t.Errorf(
				"expected amount %s, got %s",
				amount,
				payout.Amount,
			)
		}

		if payout.Currency != currency {
			t.Errorf(
				"expected currency %q, got %q",
				currency,
				payout.Currency,
			)
		}

		if payout.Description != &description {
			t.Errorf(
				"expected description %v, got %v",
				description,
				payout.Description,
			)
		}

		if payout.PayoutReference == "" {
			t.Error("expected generated payout reference")
		}

		if payout.ExternalReference == nil || *payout.ExternalReference == "" {
			t.Error("expected generated external reference")
		}

		if payout.Status == "" {
			t.Fatal("expected status, got nil")
		}

		if payout.Status != status {
			t.Errorf(
				"expected status %q, got %q",
				status,
				payout.Status,
			)
		}

		if payout.CreatedAt.Before(beforeCreate) ||
			payout.CreatedAt.After(afterCreate) {
			t.Errorf(
				"created_at %v is outside expected range [%v, %v]",
				payout.CreatedAt,
				beforeCreate,
				afterCreate,
			)
		}

		if payout.UpdatedAt.Before(beforeCreate) ||
			payout.UpdatedAt.After(afterCreate) {
			t.Errorf(
				"updated_at %v is outside expected range [%v, %v]",
				payout.UpdatedAt,
				beforeCreate,
				afterCreate,
			)
		}

		var savedPayout models.Payout

		if err := db.
			Where("id = ?", payout.ID).
			First(&savedPayout).Error; err != nil {
			t.Fatalf("failed to fetch created payout: %v", err)
		}

		if savedPayout.MerchantID != merchantID {
			t.Errorf(
				"persisted merchant ID mismatch: expected %v, got %v",
				merchantID,
				savedPayout.MerchantID,
			)
		}

		if savedPayout.SourceWalletID == nil {
			t.Fatal("persisted source wallet ID is nil")
		}

		if *savedPayout.SourceWalletID != walletID {
			t.Errorf(
				"persisted source wallet ID mismatch: expected %v, got %v",
				walletID,
				*savedPayout.SourceWalletID,
			)
		}

		if savedPayout.DestinationBankAccountID != destinationBankAccountID {
			t.Errorf(
				"persisted destination bank account ID mismatch: expected %v, got %v",
				destinationBankAccountID,
				savedPayout.DestinationBankAccountID,
			)
		}

		if !savedPayout.Amount.Equal(amount) {
			t.Errorf(
				"persisted amount mismatch: expected %s, got %s",
				amount,
				savedPayout.Amount,
			)
		}

		if savedPayout.Currency != currency {
			t.Errorf(
				"persisted currency mismatch: expected %q, got %q",
				currency,
				savedPayout.Currency,
			)
		}

		if savedPayout.Status == "" {
			t.Fatal("persisted status is nil")
		}

		if savedPayout.Status != status {
			t.Errorf(
				"persisted status mismatch: expected %q, got %q",
				status,
				savedPayout.Status,
			)
		}
	})
}

func TestCreatePayoutBankToBank(t *testing.T) {
	activityRegistry := &Registry{
		Repo: testRepo,
	}

	t.Run("creates payout successfully", func(t *testing.T) {
		sourceMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(sourceMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		destinationMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(destinationMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		sourceBankAccount := testutils.GenerateTestBankAccount()
		if errS := db.Create(sourceBankAccount).Error; errS != nil {
			t.Fatalf("error creating bank account: %v", errS)
		}

		sourceLinkedBankAccount := &models.LinkedBankAccount{
			ID:            utils.GenerateUUID(),
			MerchantID:    destinationMerchant.ID,
			BankAccountID: sourceBankAccount.ID,
			Status:        constants.MerchantStatusActive,
			Type:          sourceBankAccount.AccountType,
		}

		if errL := db.Create(sourceLinkedBankAccount).Error; errL != nil {
			t.Fatalf("error creating linked bank acccount : %v", errL)
		}

		destinationBankAccount := testutils.GenerateTestSecondaryBankAccount()
		if errB := db.Create(destinationBankAccount).Error; errB != nil {
			t.Fatalf("error creating bank account: %v", errB)
		}

		destinationLinkedBankAccount := &models.LinkedBankAccount{
			ID:            utils.GenerateUUID(),
			MerchantID:    destinationMerchant.ID,
			BankAccountID: destinationBankAccount.ID,
			Status:        constants.MerchantStatusActive,
			Type:          destinationBankAccount.AccountType,
		}

		if errL := db.Create(destinationLinkedBankAccount).Error; errL != nil {
			t.Fatalf("error creating linked bank acccount : %v", errL)
		}

		merchantID := sourceMerchant.ID
		sourceBankAccountID := sourceBankAccount.ID
		destinationBankAccountID := destinationBankAccount.ID

		amount := decimal.NewFromInt(2500)
		currency := "INR"
		description := "Bank to bank payout"
		status := constants.TransactionStatusProcessing

		request := &dto.CreateBankToBankRequest{
			Amount:                   amount,
			Currency:                 currency,
			SourceBankAccountID:      sourceBankAccountID,
			DestinationBankAccountID: destinationBankAccountID,
			Description:              &description,
		}

		beforeCreate := time.Now()

		payout, err := activityRegistry.CreatePayoutBankToBank(
			context.Background(),
			request,
			merchantID,
			status,
		)

		afterCreate := time.Now()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if payout == nil {
			t.Fatal("expected payout, got nil")
		}

		if payout.ID == uuid.Nil {
			t.Error("expected generated payout ID, got nil UUID")
		}

		if payout.MerchantID != merchantID {
			t.Errorf(
				"expected merchant ID %v, got %v",
				merchantID,
				payout.MerchantID,
			)
		}

		if payout.SourceWalletID != nil {
			t.Errorf(
				"expected source wallet ID to be nil, got %v",
				*payout.SourceWalletID,
			)
		}

		if payout.SourceBankAccountID == nil {
			t.Fatal("expected source bank account ID, got nil")
		}

		if *payout.SourceBankAccountID != sourceBankAccountID {
			t.Errorf(
				"expected source bank account ID %v, got %v",
				sourceBankAccountID,
				*payout.SourceBankAccountID,
			)
		}

		if payout.DestinationBankAccountID != destinationBankAccountID {
			t.Errorf(
				"expected destination bank account ID %v, got %v",
				destinationBankAccountID,
				payout.DestinationBankAccountID,
			)
		}

		if !payout.Amount.Equal(amount) {
			t.Errorf(
				"expected amount %s, got %s",
				amount,
				payout.Amount,
			)
		}

		if payout.Currency != currency {
			t.Errorf(
				"expected currency %q, got %q",
				currency,
				payout.Currency,
			)
		}

		if payout.Description != &description {
			t.Errorf(
				"expected description %v, got %v",
				description,
				payout.Description,
			)
		}

		if payout.PayoutReference == "" {
			t.Error("expected generated payout reference")
		}

		if payout.ExternalReference == nil || *payout.ExternalReference == "" {
			t.Error("expected generated external reference")
		}

		if payout.Status == "" {
			t.Fatal("expected status, got nil")
		}

		if payout.Status != status {
			t.Errorf(
				"expected status %q, got %q",
				status,
				payout.Status,
			)
		}

		if payout.CreatedAt.Before(beforeCreate) ||
			payout.CreatedAt.After(afterCreate) {
			t.Errorf(
				"created_at %v is outside expected range [%v, %v]",
				payout.CreatedAt,
				beforeCreate,
				afterCreate,
			)
		}

		if payout.UpdatedAt.Before(beforeCreate) ||
			payout.UpdatedAt.After(afterCreate) {
			t.Errorf(
				"updated_at %v is outside expected range [%v, %v]",
				payout.UpdatedAt,
				beforeCreate,
				afterCreate,
			)
		}

		var savedPayout models.Payout

		if err := db.
			Where("id = ?", payout.ID).
			First(&savedPayout).Error; err != nil {
			t.Fatalf("failed to fetch created payout: %v", err)
		}

		if savedPayout.MerchantID != merchantID {
			t.Errorf(
				"persisted merchant ID mismatch: expected %v, got %v",
				merchantID,
				savedPayout.MerchantID,
			)
		}

		if savedPayout.SourceWalletID != nil {
			t.Errorf("expected persisted source wallet ID to be nil")
		}

		if savedPayout.SourceBankAccountID == nil {
			t.Fatal("persisted source bank account ID is nil")
		}

		if *savedPayout.SourceBankAccountID != sourceBankAccountID {
			t.Errorf(
				"persisted source bank account ID mismatch: expected %v, got %v",
				sourceBankAccountID,
				*savedPayout.SourceBankAccountID,
			)
		}

		if savedPayout.DestinationBankAccountID != destinationBankAccountID {
			t.Errorf(
				"persisted destination bank account ID mismatch: expected %v, got %v",
				destinationBankAccountID,
				savedPayout.DestinationBankAccountID,
			)
		}

		if !savedPayout.Amount.Equal(amount) {
			t.Errorf(
				"persisted amount mismatch: expected %s, got %s",
				amount,
				savedPayout.Amount,
			)
		}

		if savedPayout.Currency != currency {
			t.Errorf(
				"persisted currency mismatch: expected %q, got %q",
				currency,
				savedPayout.Currency,
			)
		}

		if savedPayout.Status == "" {
			t.Fatal("persisted status is nil")
		}

		if savedPayout.Status != status {
			t.Errorf(
				"persisted status mismatch: expected %q, got %q",
				status,
				savedPayout.Status,
			)
		}

		if savedPayout.PayoutReference == "" ||
			payout.PayoutReference != savedPayout.PayoutReference {
			t.Error("persisted payout reference does not match returned payout")
		}

		if savedPayout.ExternalReference == nil ||
			*payout.ExternalReference != *savedPayout.ExternalReference {
			t.Error("persisted external reference does not match returned payout")
		}
	})
}

func TestUpdatePayoutStatus(t *testing.T) {
	activityRegistry := &Registry{
		Repo: testRepo,
	}

	t.Run("payout not found", func(t *testing.T) {
		status := constants.TransactionStatusCompleted
		payoutReference := "payout-does-not-exist"

		payout, err := activityRegistry.UpdatePayoutStatus(
			context.Background(),
			status,
			payoutReference,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if payout != nil {
			t.Errorf("expected payout to be nil, got %+v", payout)
		}
	})

	t.Run("status already matches", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if errM := db.Create(merchant).Error; errM != nil {
			t.Fatalf("error creating merchant: %v", errM)
		}

		sourceWallet := testutils.GenerateTestWallet(merchant.ID)
		if errS := db.Create(sourceWallet).Error; errS != nil {
			t.Fatalf("error creating wallet: %v", errS)
		}

		bankAccount := testutils.GenerateTestBankAccount()
		if errB := db.Create(bankAccount).Error; errB != nil {
			t.Fatalf("error creating bankA account: %v", errB)
		}

		linkedBankAccount := &models.LinkedBankAccount{
			ID:            utils.GenerateUUID(),
			MerchantID:    merchant.ID,
			BankAccountID: bankAccount.ID,
			Status:        constants.MerchantStatusActive,
			Type:          bankAccount.AccountType,
		}

		if errLB := db.Create(linkedBankAccount).Error; errLB != nil {
			t.Fatalf("error creating linked bank account: %v", errLB)
		}

		sourceWalletID := sourceWallet.ID
		merchantID := merchant.ID
		status := constants.TransactionStatusProcessing
		payoutReference := "payout-status-unchanged"

		payout := &models.Payout{
			ID:                       uuid.New(),
			MerchantID:               merchantID,
			SourceWalletID:           &sourceWalletID,
			DestinationBankAccountID: bankAccount.ID,
			PayoutReference:          payoutReference,
			Amount:                   decimal.NewFromInt(1000),
			Currency:                 "INR",
			Status:                   status,
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		if err := db.Create(payout).Error; err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		result, err := activityRegistry.UpdatePayoutStatus(
			context.Background(),
			status,
			payoutReference,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected payout, got nil")
		}

		if result.ID != payout.ID {
			t.Errorf(
				"expected payout ID %v, got %v",
				payout.ID,
				result.ID,
			)
		}

		if result.Status == "" {
			t.Fatal("expected status, got nil")
		}

		if result.Status != status {
			t.Errorf(
				"expected status %q, got %q",
				status,
				result.Status,
			)
		}
	})

	t.Run("updates payout status", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if errM := db.Create(merchant).Error; errM != nil {
			t.Fatalf("error creating merchant: %v", errM)
		}

		sourceWallet := testutils.GenerateTestWallet(merchant.ID)
		if errS := db.Create(sourceWallet).Error; errS != nil {
			t.Fatalf("error creating wallet: %v", errS)
		}

		bankAccount := testutils.GenerateTestBankAccount()
		if errB := db.Create(bankAccount).Error; errB != nil {
			t.Fatalf("error creating bankA account: %v", errB)
		}

		linkedBankAccount := &models.LinkedBankAccount{
			ID:            utils.GenerateUUID(),
			MerchantID:    merchant.ID,
			BankAccountID: bankAccount.ID,
			Status:        constants.MerchantStatusActive,
			Type:          bankAccount.AccountType,
		}

		if errLB := db.Create(linkedBankAccount).Error; errLB != nil {
			t.Fatalf("error creating linked bank account: %v", errLB)
		}

		oldStatus := constants.TransactionStatusProcessing
		newStatus := constants.TransactionStatusCompleted
		sourceWalletID := sourceWallet.ID
		payoutReference := "payout-status-update"

		payout := &models.Payout{
			ID:                       uuid.New(),
			MerchantID:               merchant.ID,
			SourceWalletID:           &sourceWalletID,
			DestinationBankAccountID: bankAccount.ID,
			PayoutReference:          payoutReference,
			Amount:                   decimal.NewFromInt(2500),
			Currency:                 "INR",
			Status:                   oldStatus,
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		if err := db.Create(payout).Error; err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		originalUpdatedAt := payout.UpdatedAt

		result, err := activityRegistry.UpdatePayoutStatus(
			context.Background(),
			newStatus,
			payoutReference,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected payout, got nil")
		}

		if result.ID != payout.ID {
			t.Errorf(
				"expected payout ID %v, got %v",
				payout.ID,
				result.ID,
			)
		}

		if result.Status == "" {
			t.Fatal("expected status, got nil")
		}

		if result.Status != newStatus {
			t.Errorf(
				"expected status %q, got %q",
				newStatus,
				result.Status,
			)
		}

		if !result.UpdatedAt.After(originalUpdatedAt) {
			t.Errorf(
				"expected updated_at to be updated, old=%v new=%v",
				originalUpdatedAt,
				result.UpdatedAt,
			)
		}

		var savedPayout models.Payout

		if err := db.
			Where("id = ?", payout.ID).
			First(&savedPayout).Error; err != nil {
			t.Fatalf("failed to fetch updated payout: %v", err)
		}

		if savedPayout.Status == "" {
			t.Fatal("persisted status is nil")
		}

		if savedPayout.Status != newStatus {
			t.Errorf(
				"persisted status mismatch: expected %q, got %q",
				newStatus,
				savedPayout.Status,
			)
		}

		if !savedPayout.UpdatedAt.After(originalUpdatedAt) {
			t.Errorf(
				"expected persisted updated_at to be updated, old=%v new=%v",
				originalUpdatedAt,
				savedPayout.UpdatedAt,
			)
		}
	})
}

func TestExecutePayoutWalletToBank(t *testing.T) {
	activityRegistry := &Registry{
		DB:       db,
		Repo:     testRepo,
		Services: testServices,
	}

	t.Run("successful payout with fee", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if errM := db.Create(merchant).Error; errM != nil {
			t.Fatalf("error creating merchant: %v", errM)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if errRM := db.Create(receiverMerchant).Error; errRM != nil {
			t.Fatalf("error creating receiver merchant: %v", errRM)
		}

		merchantID := merchant.ID
		receiverMerchantID := receiverMerchant.ID

		merchantWallet := testutils.GenerateTestWallet(merchantID)
		if errW := db.Create(merchantWallet).Error; errW != nil {
			t.Fatalf("error creating wallet: %v", errW)
		}

		receivingBankAccount := testutils.GenerateTestBankAccount()
		if errRB := db.Create(receivingBankAccount).Error; errRB != nil {
			t.Fatalf("error creating bank account: %v", errRB)
		}

		linkedBankAccount := &models.LinkedBankAccount{
			ID:            utils.GenerateUUID(),
			MerchantID:    receiverMerchantID,
			BankAccountID: receivingBankAccount.ID,
			Status:        receiverMerchant.Status,
			Type:          receivingBankAccount.AccountType,
		}

		if errLBA := db.Create(linkedBankAccount).Error; errLBA != nil {
			t.Fatalf("error creating linked bank account: %v", errLBA)
		}

		payoutAmount := decimal.NewFromInt(1000)
		fee := decimal.NewFromInt(50)
		totalAmount := payoutAmount.Add(fee)

		processingStatus := constants.TransactionStatusProcessing
		payoutReference := "payout-execute-success"

		payout := &models.Payout{
			ID:                       uuid.New(),
			MerchantID:               merchantID,
			SourceWalletID:           &merchantWallet.ID,
			DestinationBankAccountID: receivingBankAccount.ID,
			PayoutReference:          payoutReference,
			Amount:                   payoutAmount,
			Currency:                 "INR",
			Status:                   processingStatus,
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		if err := db.Create(payout).Error; err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		payoutVault := &models.Vault{
			ID:      uuid.New(),
			Type:    constants.PayoutVault,
			Balance: decimal.NewFromInt(20000),
			Status:  constants.VaultStatusActive,
		}

		if err := db.Create(payoutVault).Error; err != nil {
			t.Fatalf("failed to create payout vault: %v", err)
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

		desc := "test payout"

		request := &PayoutWorkflowWalletToBanRequest{
			PayoutReference: payoutReference,
			MerchantID:      merchantID,
			Fee:             fee,
			Request: dto.CreateWalletToBankRequest{
				Amount:                   payoutAmount,
				Currency:                 "INR",
				SenderWalletID:           merchantWallet.ID,
				DestinationBankAccountID: receivingBankAccount.ID,
				Description:              &desc,
			},
		}

		result, err := activityRegistry.ExecutePayoutWalletToBank(
			context.Background(),
			request,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected payout, got nil")
		}

		if result.Status == "" {
			t.Fatal("expected payout status, got empty")
		}

		if result.Status != constants.TransactionStatusCompleted {
			t.Errorf(
				"expected payout status %q, got %q",
				constants.TransactionStatusCompleted,
				result.Status,
			)
		}

		// ---------------------------------------------------------
		// Verify sender wallet
		// ---------------------------------------------------------

		var updatedWallet models.Wallets

		if err := db.
			Where("id = ?", merchantWallet.ID).
			First(&updatedWallet).Error; err != nil {
			t.Fatalf("failed to fetch updated wallet: %v", err)
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

		// ---------------------------------------------------------
		// Verify receiver bank account
		// ---------------------------------------------------------

		var updatedReceiverBank models.BankAccount

		if err := db.
			Where("id = ?", receivingBankAccount.ID).
			First(&updatedReceiverBank).Error; err != nil {
			t.Fatalf("failed to fetch receiver bank account: %v", err)
		}

		expectedReceiverBalance := decimal.NewFromInt(11000)

		if !updatedReceiverBank.Balance.Equal(expectedReceiverBalance) {
			t.Errorf(
				"expected receiver bank balance %s, got %s",
				expectedReceiverBalance,
				updatedReceiverBank.Balance,
			)
		}

		// ---------------------------------------------------------
		// Verify payout vault
		// ---------------------------------------------------------

		var updatedPayoutVault models.Vault

		if err := db.
			Where("id = ?", payoutVault.ID).
			First(&updatedPayoutVault).Error; err != nil {
			t.Fatalf("failed to fetch payout vault: %v", err)
		}

		expectedPayoutVaultBalance := decimal.NewFromInt(20000)

		if !updatedPayoutVault.Balance.Equal(expectedPayoutVaultBalance) {
			t.Errorf(
				"expected payout vault balance %s, got %s",
				expectedPayoutVaultBalance,
				updatedPayoutVault.Balance,
			)
		}

		// ---------------------------------------------------------
		// Verify company vault
		// ---------------------------------------------------------

		var updatedCompanyVault models.Vault

		if err := db.
			Where("id = ?", companyVault.ID).
			First(&updatedCompanyVault).Error; err != nil {
			t.Fatalf("failed to fetch company vault: %v", err)
		}

		expectedCompanyBalance := decimal.NewFromInt(10050)

		if !updatedCompanyVault.Balance.Equal(expectedCompanyBalance) {
			t.Errorf(
				"expected company vault balance %s, got %s",
				expectedCompanyBalance,
				updatedCompanyVault.Balance,
			)
		}

		// ---------------------------------------------------------
		// Verify ledger transaction
		// ---------------------------------------------------------

		var ledgerTransaction models.LedgerTransaction

		if err := db.
			Where("reference_id = ?", payoutReference).
			First(&ledgerTransaction).Error; err != nil {
			t.Fatalf(
				"failed to fetch payout ledger transaction: %v",
				err,
			)
		}

		if ledgerTransaction.Status != constants.TransactionStatusCompleted {
			t.Errorf(
				"expected ledger status %q, got %q",
				constants.TransactionStatusCompleted,
				ledgerTransaction.Status,
			)
		}

		// ---------------------------------------------------------
		// Verify ledger entries
		// ---------------------------------------------------------

		entries := make([]models.LedgerEntry, 0)

		if err := db.
			Where("ledger_transaction_id = ?", ledgerTransaction.ID).
			Find(&entries).Error; err != nil {
			t.Fatalf("failed to fetch ledger entries: %v", err)
		}

		if len(entries) != 6 {
			t.Fatalf(
				"expected 6 ledger entries, got %d",
				len(entries),
			)
		}

		var walletDebit decimal.Decimal
		var payoutVaultCredit decimal.Decimal
		var payoutVaultDebit decimal.Decimal
		var bankCredit decimal.Decimal
		var companyVaultCredit decimal.Decimal

		for _, entry := range entries {
			switch {
			case entry.AccountType == constants.LedgerAccountTypeWallet &&
				entry.AccountID == merchantWallet.ID &&
				entry.EntryType == constants.LedgerEntryTypeDebit:

				walletDebit = walletDebit.Add(entry.Amount)

			case entry.AccountType == constants.LedgerAccountTypeVault &&
				entry.AccountID == payoutVault.ID &&
				entry.EntryType == constants.LedgerEntryTypeCredit:

				payoutVaultCredit = payoutVaultCredit.Add(entry.Amount)

			case entry.AccountType == constants.LedgerAccountTypeVault &&
				entry.AccountID == payoutVault.ID &&
				entry.EntryType == constants.LedgerEntryTypeDebit:

				payoutVaultDebit = payoutVaultDebit.Add(entry.Amount)

			case entry.AccountType == constants.LedgerAccountTypeBankAccount &&
				entry.AccountID == receivingBankAccount.ID &&
				entry.EntryType == constants.LedgerEntryTypeCredit:

				bankCredit = bankCredit.Add(entry.Amount)

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

		if !payoutVaultCredit.Equal(totalAmount) {
			t.Errorf(
				"expected payout vault credit %s, got %s",
				totalAmount,
				payoutVaultCredit,
			)
		}

		if !payoutVaultDebit.Equal(totalAmount) {
			t.Errorf(
				"expected payout vault debit %s, got %s",
				totalAmount,
				payoutVaultDebit,
			)
		}

		if !bankCredit.Equal(payoutAmount) {
			t.Errorf(
				"expected bank credit %s, got %s",
				payoutAmount,
				bankCredit,
			)
		}

		if !companyVaultCredit.Equal(fee) {
			t.Errorf(
				"expected company vault credit %s, got %s",
				fee,
				companyVaultCredit,
			)
		}
	})

	t.Run("insufficient balance rolls back transaction", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if errM := db.Create(merchant).Error; errM != nil {
			t.Fatalf("error creating merchant: %v", errM)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if errRM := db.Create(receiverMerchant).Error; errRM != nil {
			t.Fatalf("error creating receiver merchant: %v", errRM)
		}

		merchantID := merchant.ID
		receiverMerchantID := receiverMerchant.ID

		merchantWallet := testutils.GenerateTestWallet(merchantID)
		merchantWallet.AvailableBalance = decimal.NewFromInt(500)

		if errW := db.Create(merchantWallet).Error; errW != nil {
			t.Fatalf("error creating wallet: %v", errW)
		}

		receivingBankAccount := testutils.GenerateTestBankAccount()
		if errRB := db.Create(receivingBankAccount).Error; errRB != nil {
			t.Fatalf("error creating bank account: %v", errRB)
		}

		linkedBankAccount := &models.LinkedBankAccount{
			ID:            utils.GenerateUUID(),
			MerchantID:    receiverMerchantID,
			BankAccountID: receivingBankAccount.ID,
			Status:        receiverMerchant.Status,
			Type:          receivingBankAccount.AccountType,
		}

		if errLBA := db.Create(linkedBankAccount).Error; errLBA != nil {
			t.Fatalf("error creating linked bank account: %v", errLBA)
		}

		payoutReference := "payout-insufficient-balance"
		status := constants.TransactionStatusProcessing

		payout := &models.Payout{
			ID:                       uuid.New(),
			MerchantID:               merchantID,
			SourceWalletID:           &merchantWallet.ID,
			DestinationBankAccountID: receivingBankAccount.ID,
			PayoutReference:          payoutReference,
			Amount:                   decimal.NewFromInt(1000),
			Currency:                 "INR",
			Status:                   status,
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		if err := db.Create(payout).Error; err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		payoutVault := &models.Vault{
			ID:      uuid.New(),
			Type:    constants.PayoutVault,
			Balance: decimal.NewFromInt(5000),
			Status:  constants.VaultStatusActive,
		}

		if err := db.Create(payoutVault).Error; err != nil {
			t.Fatalf("failed to create payout vault: %v", err)
		}

		companyVault := &models.Vault{
			ID:      uuid.New(),
			Type:    constants.CompanyVault,
			Balance: decimal.NewFromInt(5000),
			Status:  constants.VaultStatusActive,
		}

		if err := db.Create(companyVault).Error; err != nil {
			t.Fatalf("failed to create company vault: %v", err)
		}

		fee := decimal.NewFromInt(50)

		request := &PayoutWorkflowWalletToBanRequest{
			PayoutReference: payoutReference,
			MerchantID:      merchantID,
			Fee:             fee,
			Request: dto.CreateWalletToBankRequest{
				Amount:                   decimal.NewFromInt(1000),
				Currency:                 "INR",
				SenderWalletID:           merchantWallet.ID,
				DestinationBankAccountID: receivingBankAccount.ID,
			},
		}

		result, err := activityRegistry.ExecutePayoutWalletToBank(
			context.Background(),
			request,
		)

		if err == nil {
			t.Fatal("expected insufficient balance error, got nil")
		}

		if result != nil {
			t.Errorf("expected nil payout, got %+v", result)
		}

		var updatedWallet models.Wallets
		if err := db.
			Where("id = ?", merchantWallet.ID).
			First(&updatedWallet).Error; err != nil {
			t.Fatalf("failed to fetch wallet: %v", err)
		}

		if !updatedWallet.AvailableBalance.Equal(decimal.NewFromInt(500)) {
			t.Errorf(
				"expected available balance 500 after rollback, got %s",
				updatedWallet.AvailableBalance,
			)
		}

		if !updatedWallet.ReservedBalance.Equal(decimal.Zero) {
			t.Errorf(
				"expected reserved balance 0 after rollback, got %s",
				updatedWallet.ReservedBalance,
			)
		}

		var updatedPayout models.Payout
		if err := db.
			Where("id = ?", payout.ID).
			First(&updatedPayout).Error; err != nil {
			t.Fatalf("failed to fetch payout: %v", err)
		}

		if updatedPayout.Status != status {
			t.Errorf(
				"expected payout status %q after rollback, got %q",
				status,
				updatedPayout.Status,
			)
		}

		var ledgerCount int64
		if err := db.
			Model(&models.LedgerTransaction{}).
			Where("reference_id = ?", payoutReference).
			Count(&ledgerCount).Error; err != nil {
			t.Fatalf("failed to count ledger transactions: %v", err)
		}

		if ledgerCount != 0 {
			t.Errorf(
				"expected no ledger transaction after rollback, got %d",
				ledgerCount,
			)
		}
	})
}

func TestExecutePayoutBankToBank(t *testing.T) {
	activityRegistry := &Registry{
		DB:       db,
		Repo:     testRepo,
		Services: testServices,
	}
	t.Run("successful bank to bank payout", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if errM := db.Create(merchant).Error; errM != nil {
			t.Fatalf("error creating merchant: %v", errM)
		}

		senderBankAccount := testutils.GenerateTestBankAccount()
		if errSBA := db.Create(senderBankAccount).Error; errSBA != nil {
			t.Fatalf("error creating sender bank account: %v", errSBA)
		}

		receiverBankAccount := testutils.GenerateTestBankAccount()
		if errRBA := db.Create(receiverBankAccount).Error; errRBA != nil {
			t.Fatalf("error creating receiver bank account: %v", errRBA)
		}

		payoutAmount := decimal.NewFromInt(1000)
		fee := decimal.NewFromInt(50)
		totalAmount := payoutAmount.Add(fee)

		payoutReference := "payout-bank-to-bank-success"
		processingStatus := constants.TransactionStatusProcessing

		payout := &models.Payout{
			ID:                       uuid.New(),
			MerchantID:               merchant.ID,
			SourceBankAccountID:      &senderBankAccount.ID,
			DestinationBankAccountID: receiverBankAccount.ID,
			PayoutReference:          payoutReference,
			Amount:                   payoutAmount,
			Currency:                 "INR",
			Status:                   processingStatus,
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		if err := db.Create(payout).Error; err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		request := &PayoutWorkflowBankToBanRequest{
			PayoutReference: payoutReference,
			MerchantID:      merchant.ID,
			Fee:             fee,
			Request: dto.CreateBankToBankRequest{
				Amount:                   payoutAmount,
				Currency:                 "INR",
				SourceBankAccountID:      senderBankAccount.ID,
				DestinationBankAccountID: receiverBankAccount.ID,
			},
		}

		result, err := activityRegistry.ExecutePayoutBankToBank(
			context.Background(),
			request,
		)

		if err != nil {
			t.Fatalf("expected successful payout, got error: %v", err)
		}

		if result == nil {
			t.Fatal("expected payout result, got nil")
		}

		if result.ID != payout.ID {
			t.Errorf(
				"expected payout ID %s, got %s",
				payout.ID,
				result.ID,
			)
		}

		if result.Status != constants.TransactionStatusCompleted {
			t.Errorf(
				"expected payout status %q, got %q",
				constants.TransactionStatusCompleted,
				result.Status,
			)
		}

		var updatedSenderBankAccount models.BankAccount
		if err := db.
			Where("id = ?", senderBankAccount.ID).
			First(&updatedSenderBankAccount).Error; err != nil {
			t.Fatalf("failed to fetch sender bank account: %v", err)
		}

		expectedSenderBalance := decimal.NewFromInt(8950)

		if !updatedSenderBankAccount.Balance.Equal(expectedSenderBalance) {
			t.Errorf(
				"expected sender bank balance %s, got %s",
				expectedSenderBalance,
				updatedSenderBankAccount.Balance,
			)
		}

		var updatedReceiverBankAccount models.BankAccount
		if err := db.
			Where("id = ?", receiverBankAccount.ID).
			First(&updatedReceiverBankAccount).Error; err != nil {
			t.Fatalf("failed to fetch receiver bank account: %v", err)
		}

		expectedReceiverBalance := decimal.NewFromInt(11050)

		if !updatedReceiverBankAccount.Balance.Equal(expectedReceiverBalance) {
			t.Errorf(
				"expected receiver bank balance %s, got %s",
				expectedReceiverBalance,
				updatedReceiverBankAccount.Balance,
			)
		}

		var updatedPayout models.Payout
		if err := db.
			Where("id = ?", payout.ID).
			First(&updatedPayout).Error; err != nil {
			t.Fatalf("failed to fetch payout: %v", err)
		}

		if updatedPayout.Status != constants.TransactionStatusCompleted {
			t.Errorf(
				"expected payout status %q, got %q",
				constants.TransactionStatusCompleted,
				updatedPayout.Status,
			)
		}

		var ledgerTransaction models.LedgerTransaction
		if err := db.
			Where("reference_id = ?", payoutReference).
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

		if ledgerTransaction.SettlementStatus != constants.TransactionStatusCompleted {
			t.Errorf(
				"expected settlement status %q, got %q",
				constants.TransactionStatusCompleted,
				ledgerTransaction.SettlementStatus,
			)
		}

		var ledgerEntries []models.LedgerEntry
		if err := db.
			Where("ledger_transaction_id = ?", ledgerTransaction.ID).
			Find(&ledgerEntries).Error; err != nil {
			t.Fatalf("failed to fetch ledger entries: %v", err)
		}

		if len(ledgerEntries) != 2 {
			t.Fatalf(
				"expected 2 ledger entries, got %d",
				len(ledgerEntries),
			)
		}

		var debitAmount decimal.Decimal
		var creditAmount decimal.Decimal

		for _, entry := range ledgerEntries {
			switch entry.EntryType {
			case constants.LedgerEntryTypeDebit:
				if entry.AccountID != senderBankAccount.ID {
					t.Errorf(
						"expected debit account %s, got %s",
						senderBankAccount.ID,
						entry.AccountID,
					)
				}

				debitAmount = debitAmount.Add(entry.Amount)

			case constants.LedgerEntryTypeCredit:
				if entry.AccountID != receiverBankAccount.ID {
					t.Errorf(
						"expected credit account %s, got %s",
						receiverBankAccount.ID,
						entry.AccountID,
					)
				}

				creditAmount = creditAmount.Add(entry.Amount)
			}
		}

		if !debitAmount.Equal(totalAmount) {
			t.Errorf(
				"expected debit amount %s, got %s",
				totalAmount,
				debitAmount,
			)
		}

		if !creditAmount.Equal(totalAmount) {
			t.Errorf(
				"expected credit amount %s, got %s",
				totalAmount,
				creditAmount,
			)
		}
	})

	t.Run("insufficient sender bank balance rolls back transaction", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if errM := db.Create(merchant).Error; errM != nil {
			t.Fatalf("error creating merchant: %v", errM)
		}

		senderBankAccount := testutils.GenerateTestBankAccount()
		senderBankAccount.Balance = decimal.NewFromInt(500)

		if errSBA := db.Create(senderBankAccount).Error; errSBA != nil {
			t.Fatalf("error creating sender bank account: %v", errSBA)
		}

		receiverBankAccount := testutils.GenerateTestBankAccount()
		if errRBA := db.Create(receiverBankAccount).Error; errRBA != nil {
			t.Fatalf("error creating receiver bank account: %v", errRBA)
		}

		payoutAmount := decimal.NewFromInt(1000)
		fee := decimal.NewFromInt(50)

		payoutReference := "payout-bank-to-bank-insufficient"
		processingStatus := constants.TransactionStatusProcessing

		payout := &models.Payout{
			ID:                       uuid.New(),
			MerchantID:               merchant.ID,
			SourceBankAccountID:      &senderBankAccount.ID,
			DestinationBankAccountID: receiverBankAccount.ID,
			PayoutReference:          payoutReference,
			Amount:                   payoutAmount,
			Currency:                 "INR",
			Status:                   processingStatus,
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}

		if err := db.Create(payout).Error; err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		request := &PayoutWorkflowBankToBanRequest{
			PayoutReference: payoutReference,
			MerchantID:      merchant.ID,
			Fee:             fee,
			Request: dto.CreateBankToBankRequest{
				Amount:                   payoutAmount,
				Currency:                 "INR",
				SourceBankAccountID:      senderBankAccount.ID,
				DestinationBankAccountID: receiverBankAccount.ID,
			},
		}

		result, err := activityRegistry.ExecutePayoutBankToBank(
			context.Background(),
			request,
		)

		if err == nil {
			t.Fatal("expected insufficient balance error, got nil")
		}

		if result != nil {
			t.Errorf("expected nil payout, got %+v", result)
		}

		var updatedSenderBankAccount models.BankAccount
		if err := db.
			Where("id = ?", senderBankAccount.ID).
			First(&updatedSenderBankAccount).Error; err != nil {
			t.Fatalf("failed to fetch sender bank account: %v", err)
		}

		if !updatedSenderBankAccount.Balance.Equal(decimal.NewFromInt(500)) {
			t.Errorf(
				"expected sender bank balance 500 after rollback, got %s",
				updatedSenderBankAccount.Balance,
			)
		}

		var updatedReceiverBankAccount models.BankAccount
		if err := db.
			Where("id = ?", receiverBankAccount.ID).
			First(&updatedReceiverBankAccount).Error; err != nil {
			t.Fatalf("failed to fetch receiver bank account: %v", err)
		}

		if !updatedReceiverBankAccount.Balance.Equal(decimal.NewFromInt(10000)) {
			t.Errorf(
				"expected receiver bank balance 10000 after rollback, got %s",
				updatedReceiverBankAccount.Balance,
			)
		}

		var updatedPayout models.Payout
		if err := db.
			Where("id = ?", payout.ID).
			First(&updatedPayout).Error; err != nil {
			t.Fatalf("failed to fetch payout: %v", err)
		}

		if updatedPayout.Status != processingStatus {
			t.Errorf(
				"expected payout status %q after rollback, got %q",
				processingStatus,
				updatedPayout.Status,
			)
		}

		var ledgerCount int64
		if err := db.
			Model(&models.LedgerTransaction{}).
			Where("reference_id = ?", payoutReference).
			Count(&ledgerCount).Error; err != nil {
			t.Fatalf("failed to count ledger transactions: %v", err)
		}

		if ledgerCount != 0 {
			t.Errorf(
				"expected no ledger transaction after rollback, got %d",
				ledgerCount,
			)
		}
	})
}
