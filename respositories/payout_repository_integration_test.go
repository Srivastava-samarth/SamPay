package repositories

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCreatePayoutWalletToBank(t *testing.T) {
	repo := NewPayoutRepository(db)

	t.Run("creates wallet to bank payout", func(t *testing.T) {

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatal(err)
		}

		wallet := testutils.GenerateTestWallet(merchant.ID)

		if err := db.Create(wallet).Error; err != nil {
			t.Fatal(err)
		}

		destinationBankAccount := testutils.GenerateTestBankAccount()

		if err := db.Create(destinationBankAccount).Error; err != nil {
			t.Fatal(err)
		}

		description := "want to do test"

		request := &dto.CreateWalletToBankRequest{
			SenderWalletID:           wallet.ID,
			DestinationBankAccountID: destinationBankAccount.ID,
			Amount:                   decimal.NewFromInt(1000),
			Currency:                 "INR",
			Description:              &description,
		}
		status := constants.TransactionStatusPending

		payout, err := repo.CreatePayoutWalletToBank(
			merchant.ID,
			request,
			status,
		)

		if err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		if payout == nil {
			t.Fatal("expected payout, got nil")
		}

		if payout.ID == uuid.Nil {
			t.Error("expected generated payout ID")
		}

		if payout.MerchantID == uuid.Nil {
			t.Fatal("expected merchant ID, got nil")
		}

		if payout.MerchantID != merchant.ID {
			t.Errorf(
				"expected merchant ID %s, got %s",
				merchant.ID,
				payout.MerchantID,
			)
		}

		if payout.SourceWalletID == nil {
			t.Fatal("expected source wallet ID, got nil")
		}

		if *payout.SourceWalletID != wallet.ID {
			t.Errorf(
				"expected source wallet ID %s, got %s",
				wallet.ID,
				*payout.SourceWalletID,
			)
		}

		if payout.SourceBankAccountID != nil {
			t.Errorf(
				"expected source bank account ID to be nil, got %v",
				*payout.SourceBankAccountID,
			)
		}

		if payout.DestinationBankAccountID == uuid.Nil {
			t.Fatal("expected destination bank account ID, got nil")
		}

		if payout.DestinationBankAccountID != request.DestinationBankAccountID {
			t.Errorf(
				"expected destination bank account ID %s, got %s",
				request.DestinationBankAccountID,
				payout.DestinationBankAccountID,
			)
		}

		if payout.PayoutReference == "" {
			t.Error("expected generated payout reference")
		}

		if payout.ExternalReference == nil {
			t.Fatal("expected external reference, got nil")
		}

		if payout.Amount != request.Amount {
			t.Errorf(
				"expected amount %s, got %s",
				request.Amount,
				payout.Amount,
			)
		}

		if payout.Currency != request.Currency {
			t.Errorf(
				"expected currency %s, got %s",
				request.Currency,
				payout.Currency,
			)
		}

		if payout.Status == "" {
			t.Fatal("expected status, got nil")
		}

		if payout.Status != status {
			t.Errorf(
				"expected status %s, got %s",
				status,
				payout.Status,
			)
		}

		// Verify that the payout was actually persisted.
		var persistedPayout *models.Payout

		if err := db.
			Where("id = ?", payout.ID).
			First(&persistedPayout).Error; err != nil {
			t.Fatalf("failed to fetch persisted payout: %v", err)
		}

		if persistedPayout.ID != payout.ID {
			t.Errorf(
				"expected persisted payout ID %s, got %s",
				payout.ID,
				persistedPayout.ID,
			)
		}
	})
}

func TestCreatePayoutBankToBank(t *testing.T) {
	repo := NewPayoutRepository(db)

	t.Run("creates bank to bank payout", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatal(err)
		}

		sourceBankAccount := testutils.GenerateTestBankAccount()

		if err := db.Create(sourceBankAccount).Error; err != nil {
			t.Fatal(err)
		}

		destinationBankAccount := testutils.GenerateTestBankAccount()

		if err := db.Create(destinationBankAccount).Error; err != nil {
			t.Fatal(err)
		}

		status := constants.TransactionStatusPending
		description := "bank to bank payout test"

		request := &dto.CreateBankToBankRequest{
			SourceBankAccountID:      sourceBankAccount.ID,
			DestinationBankAccountID: destinationBankAccount.ID,
			Amount:                   decimal.NewFromInt(1000),
			Currency:                 "INR",
			Description:              &description,
		}

		payout, err := repo.CreatePayoutBankToBank(
			merchant.ID,
			request,
			status,
		)

		if err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		if payout == nil {
			t.Fatal("expected payout, got nil")
		}

		if payout.ID == uuid.Nil {
			t.Error("expected generated payout ID")
		}

		if payout.MerchantID != merchant.ID {
			t.Errorf(
				"expected merchant ID %s, got %s",
				merchant.ID,
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

		if *payout.SourceBankAccountID != sourceBankAccount.ID {
			t.Errorf(
				"expected source bank account ID %s, got %s",
				sourceBankAccount.ID,
				*payout.SourceBankAccountID,
			)
		}

		if payout.DestinationBankAccountID != destinationBankAccount.ID {
			t.Errorf(
				"expected destination bank account ID %s, got %s",
				destinationBankAccount.ID,
				payout.DestinationBankAccountID,
			)
		}

		if payout.PayoutReference == "" {
			t.Error("expected generated payout reference")
		}

		if payout.ExternalReference == nil {
			t.Fatal("expected external reference, got nil")
		}

		if payout.Amount != request.Amount {
			t.Errorf(
				"expected amount %s, got %s",
				request.Amount,
				payout.Amount,
			)
		}

		if payout.Currency != request.Currency {
			t.Errorf(
				"expected currency %s, got %s",
				request.Currency,
				payout.Currency,
			)
		}

		if payout.Status != status {
			t.Errorf(
				"expected status %s, got %s",
				status,
				payout.Status,
			)
		}

		var persistedPayout *models.Payout

		if err := db.
			Where("id = ?", payout.ID).
			First(&persistedPayout).Error; err != nil {
			t.Fatalf("failed to fetch persisted payout: %v", err)
		}

		if persistedPayout.ID != payout.ID {
			t.Errorf(
				"expected persisted payout ID %s, got %s",
				payout.ID,
				persistedPayout.ID,
			)
		}
	})
}

func TestUpdatePayoutStatus(t *testing.T) {
	repo := NewPayoutRepository(db)

	t.Run("updates payout status", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatal(err)
		}

		wallet := testutils.GenerateTestWallet(merchant.ID)

		if err := db.Create(wallet).Error; err != nil {
			t.Fatal(err)
		}

		destinationBankAccount := testutils.GenerateTestBankAccount()

		if err := db.Create(destinationBankAccount).Error; err != nil {
			t.Fatal(err)
		}

		description := "update payout status test"
		status := constants.TransactionStatusPending

		request := &dto.CreateWalletToBankRequest{
			SenderWalletID:           wallet.ID,
			DestinationBankAccountID: destinationBankAccount.ID,
			Amount:                   decimal.NewFromInt(1000),
			Currency:                 "INR",
			Description:              &description,
		}

		payout, err := repo.CreatePayoutWalletToBank(
			merchant.ID,
			request,
			status,
		)

		if err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		updatedStatus := constants.TransactionStatusCompleted

		updatedPayout, err := repo.UpdatePayoutStatus(
			payout.PayoutReference,
			updatedStatus,
		)

		if err != nil {
			t.Fatalf("failed to update payout status: %v", err)
		}

		if updatedPayout == nil {
			t.Fatal("expected updated payout, got nil")
		}

		if updatedPayout.Status == "" {
			t.Fatal("expected payout status, got nil")
		}

		if updatedPayout.Status != updatedStatus {
			t.Errorf(
				"expected status %s, got %s",
				updatedStatus,
				updatedPayout.Status,
			)
		}

		var persistedPayout *models.Payout

		if err := db.
			Where("id = ?", payout.ID).
			First(&persistedPayout).Error; err != nil {
			t.Fatalf("failed to fetch persisted payout: %v", err)
		}

		if persistedPayout.Status == "" {
			t.Fatal("expected persisted payout status, got nil")
		}

		if persistedPayout.Status != updatedStatus {
			t.Errorf(
				"expected persisted status %s, got %s",
				updatedStatus,
				persistedPayout.Status,
			)
		}
	})

	t.Run("returns payout when status is already the same", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatal(err)
		}

		wallet := testutils.GenerateTestWallet(merchant.ID)

		if err := db.Create(wallet).Error; err != nil {
			t.Fatal(err)
		}

		destinationBankAccount := testutils.GenerateTestBankAccount()

		if err := db.Create(destinationBankAccount).Error; err != nil {
			t.Fatal(err)
		}

		description := "same status test"
		status := constants.TransactionStatusPending

		request := &dto.CreateWalletToBankRequest{
			SenderWalletID:           wallet.ID,
			DestinationBankAccountID: destinationBankAccount.ID,
			Amount:                   decimal.NewFromInt(1000),
			Currency:                 "INR",
			Description:              &description,
		}

		payout, err := repo.CreatePayoutWalletToBank(
			merchant.ID,
			request,
			status,
		)

		if err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		updatedPayout, err := repo.UpdatePayoutStatus(
			payout.PayoutReference,
			status,
		)

		if err != nil {
			t.Fatalf("failed to update payout status: %v", err)
		}

		if updatedPayout == nil {
			t.Fatal("expected payout, got nil")
		}

		if updatedPayout.Status == "" {
			t.Fatal("expected payout status, got nil")
		}

		if updatedPayout.Status != status {
			t.Errorf(
				"expected status %s, got %s",
				status,
				updatedPayout.Status,
			)
		}
	})
}

func TestGetPayoutsByMerchantID(t *testing.T) {
	repo := NewPayoutRepository(db)

	t.Run("returns payouts for merchant", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatal(err)
		}

		wallet := testutils.GenerateTestWallet(merchant.ID)

		if err := db.Create(wallet).Error; err != nil {
			t.Fatal(err)
		}

		destinationBankAccount := testutils.GenerateTestBankAccount()

		if err := db.Create(destinationBankAccount).Error; err != nil {
			t.Fatal(err)
		}

		status := constants.TransactionStatusPending
		description := "get payouts test"

		request := &dto.CreateWalletToBankRequest{
			SenderWalletID:           wallet.ID,
			DestinationBankAccountID: destinationBankAccount.ID,
			Amount:                   decimal.NewFromInt(1000),
			Currency:                 "INR",
			Description:              &description,
		}

		payout, err := repo.CreatePayoutWalletToBank(
			merchant.ID,
			request,
			status,
		)

		if err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		payouts, err := repo.GetPayoutsByMerchantID(merchant.ID)

		if err != nil {
			t.Fatalf("failed to get payouts: %v", err)
		}

		if len(payouts) == 0 {
			t.Fatal("expected payouts, got none")
		}

		found := false

		for _, result := range payouts {
			if result.ID == payout.ID {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected payout %s in merchant payouts", payout.ID)
		}

		for _, result := range payouts {
			if result.MerchantID != merchant.ID {
				t.Errorf(
					"expected merchant ID %s, got %s",
					merchant.ID,
					result.MerchantID,
				)
			}
		}
	})
}

func TestGetPayoutByID(t *testing.T) {
	repo := NewPayoutRepository(db)

	t.Run("returns payout by ID", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatal(err)
		}

		wallet := testutils.GenerateTestWallet(merchant.ID)

		if err := db.Create(wallet).Error; err != nil {
			t.Fatal(err)
		}

		destinationBankAccount := testutils.GenerateTestBankAccount()

		if err := db.Create(destinationBankAccount).Error; err != nil {
			t.Fatal(err)
		}

		status := constants.TransactionStatusPending
		description := "get payout by ID test"

		request := &dto.CreateWalletToBankRequest{
			SenderWalletID:           wallet.ID,
			DestinationBankAccountID: destinationBankAccount.ID,
			Amount:                   decimal.NewFromInt(1000),
			Currency:                 "INR",
			Description:              &description,
		}

		createdPayout, err := repo.CreatePayoutWalletToBank(
			merchant.ID,
			request,
			status,
		)

		if err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		payout, err := repo.GetPayoutByID(createdPayout.ID)

		if err != nil {
			t.Fatalf("failed to get payout: %v", err)
		}

		if payout == nil {
			t.Fatal("expected payout, got nil")
		}

		if payout.ID != createdPayout.ID {
			t.Errorf(
				"expected payout ID %s, got %s",
				createdPayout.ID,
				payout.ID,
			)
		}

		if payout.MerchantID != merchant.ID {
			t.Errorf(
				"expected merchant ID %s, got %s",
				merchant.ID,
				payout.MerchantID,
			)
		}
	})

	t.Run("returns nil when payout does not exist", func(t *testing.T) {
		payoutID := uuid.New()

		payout, err := repo.GetPayoutByID(payoutID)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if payout != nil {
			t.Errorf("expected nil payout, got %v", payout)
		}
	})
}
