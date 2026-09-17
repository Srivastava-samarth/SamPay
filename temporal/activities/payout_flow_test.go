package activities

import (
	"context"
	"testing"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
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
