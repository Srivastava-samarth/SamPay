package repositories

import (
	"errors"
	"testing"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func TestCreateBankAccount(t *testing.T) {
	repo := NewBankRepository(db)

	accountName := "Test Bank Account"
	accountType := constants.BankAccountTypePrimary

	request := &models.BankAccount{
		AccountName: accountName,
		AccountType: accountType,
	}

	bankAccount, err := repo.CreateBankAccount(request)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if bankAccount == nil {
		t.Fatal("expected bank account, got nil")
	}

	if bankAccount.ID == uuid.Nil {
		t.Error("expected bank account ID to be generated")
	}

	if bankAccount.AccountNumber == "" {
		t.Error("expected account number to be generated")
	}

	if bankAccount.AccountName == "" || bankAccount.AccountName != accountName {
		t.Errorf(
			"expected account name %v, got %v",
			accountName,
			bankAccount.AccountName,
		)
	}

	if bankAccount.BankName != "SamPay Bank" {
		t.Errorf(
			"expected bank name %v, got %v",
			"SamPay Bank",
			bankAccount.BankName,
		)
	}

	if bankAccount.IFSCCode != "SAMP5917AY" {
		t.Errorf(
			"expected IFSC code %v, got %v",
			"SAMP5917AY",
			bankAccount.IFSCCode,
		)
	}

	if bankAccount.AccountType == "" || bankAccount.AccountType != accountType {
		t.Errorf(
			"expected account type %v, got %v",
			accountType,
			bankAccount.AccountType,
		)
	}

	if !bankAccount.Balance.Equal(decimal.Zero) {
		t.Errorf(
			"expected balance %v, got %v",
			decimal.Zero,
			bankAccount.Balance,
		)
	}

	if bankAccount.Status != "active" {
		t.Errorf(
			"expected status %v, got %v",
			"active",
			bankAccount.Status,
		)
	}

	var storedBankAccount models.BankAccount
	if err := db.Where("id = ?", bankAccount.ID).First(&storedBankAccount).Error; err != nil {
		t.Fatalf("failed to fetch created bank account: %v", err)
	}

	if storedBankAccount.ID != bankAccount.ID {
		t.Errorf(
			"expected stored ID %v, got %v",
			bankAccount.ID,
			storedBankAccount.ID,
		)
	}
}

func TestGetBankAccountByID(t *testing.T) {
	repo := NewBankRepository(db)

	accountName := "Test Bank Account"
	accountType := "savings"

	request := &models.BankAccount{
		AccountName: accountName,
		AccountType: accountType,
	}

	createdBankAccount, err := repo.CreateBankAccount(request)
	if err != nil {
		t.Fatalf("failed to create bank account: %v", err)
	}

	t.Run("existing bank account", func(t *testing.T) {
		bankAccount, err := repo.GetBankAccountByID(createdBankAccount.ID)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if bankAccount == nil {
			t.Fatal("expected bank account, got nil")
		}

		if bankAccount.ID != createdBankAccount.ID {
			t.Errorf(
				"expected bank account ID %v, got %v",
				createdBankAccount.ID,
				bankAccount.ID,
			)
		}

		if bankAccount.AccountName != createdBankAccount.AccountName {
			t.Errorf(
				"expected account name %v, got %v",
				createdBankAccount.AccountName,
				bankAccount.AccountName,
			)
		}

		if bankAccount.AccountNumber != createdBankAccount.AccountNumber {
			t.Errorf(
				"expected account number %v, got %v",
				createdBankAccount.AccountNumber,
				bankAccount.AccountNumber,
			)
		}
	})

	t.Run("non-existing bank account", func(t *testing.T) {
		nonExistingID := uuid.New()

		bankAccount, err := repo.GetBankAccountByID(nonExistingID)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf(
				"expected gorm.ErrRecordNotFound, got %v",
				err,
			)
		}

		if bankAccount != nil {
			t.Errorf("expected nil bank account, got %v", bankAccount)
		}
	})
}

func TestUpdateBankAccount(t *testing.T) {
	repo := NewBankRepository(db)

	accountName := "Original Account"
	accountType := constants.BankAccountTypePrimary

	request := &models.BankAccount{
		AccountName: accountName,
		AccountType: accountType,
	}

	bankAccount, err := repo.CreateBankAccount(request)
	if err != nil {
		t.Fatalf("failed to create bank account: %v", err)
	}

	t.Run("update bank account", func(t *testing.T) {
		newAccountName := "Updated Account"
		newAccountType := "current"
		newStatus := "inactive"
		newBalance := decimal.NewFromInt(5000)

		updateRequest := &dto.UpdateBankAccountRequest{
			Balance:     newBalance,
			AccountName: newAccountName,
			AccountType: newAccountType,
			Status:      newStatus,
		}

		updatedBankAccount, err := repo.UpdateBankAccount(
			bankAccount.ID,
			updateRequest,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if updatedBankAccount == nil {
			t.Fatal("expected bank account, got nil")
		}

		if updatedBankAccount.ID != bankAccount.ID {
			t.Errorf(
				"expected ID %v, got %v",
				bankAccount.ID,
				updatedBankAccount.ID,
			)
		}

		if !updatedBankAccount.Balance.Equal(newBalance) {
			t.Errorf(
				"expected balance %v, got %v",
				newBalance,
				updatedBankAccount.Balance,
			)
		}

		if updatedBankAccount.AccountName == "" ||
			updatedBankAccount.AccountName != newAccountName {
			t.Errorf(
				"expected account name %v, got %v",
				newAccountName,
				updatedBankAccount.AccountName,
			)
		}

		if updatedBankAccount.AccountType == "" ||
			updatedBankAccount.AccountType != newAccountType {
			t.Errorf(
				"expected account type %v, got %v",
				newAccountType,
				updatedBankAccount.AccountType,
			)
		}

		if updatedBankAccount.Status == "" ||
			updatedBankAccount.Status != newStatus {
			t.Errorf(
				"expected status %v, got %v",
				newStatus,
				updatedBankAccount.Status,
			)
		}
	})

	t.Run("no updates", func(t *testing.T) {
		// Create a fresh active account because the previous case
		// changed the original account to inactive.
		accountName := "No Update Account"
		accountType := "savings"

		createRequest := &models.BankAccount{
			AccountName: accountName,
			AccountType: accountType,
		}

		account, err := repo.CreateBankAccount(createRequest)
		if err != nil {
			t.Fatalf("failed to create bank account: %v", err)
		}

		updateRequest := &dto.UpdateBankAccountRequest{
			Balance:     decimal.Zero,
			AccountName: account.AccountName,
			AccountType: account.AccountType,
			Status:      account.Status,
		}

		updatedBankAccount, err := repo.UpdateBankAccount(
			account.ID,
			updateRequest,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "no updates to be done" {
			t.Errorf(
				"expected error %q, got %q",
				"no updates to be done",
				err.Error(),
			)
		}

		if updatedBankAccount != nil {
			t.Errorf(
				"expected nil bank account, got %v",
				updatedBankAccount,
			)
		}
	})

	t.Run("inactive bank account", func(t *testing.T) {
		inactiveStatus := "inactive"
		accountName := "Inactive Account"
		accountType := "savings"

		inactiveAccount := &models.BankAccount{
			ID:            uuid.New(),
			AccountNumber: utils.GenerateBankAccountNumber(),
			AccountName:   accountName,
			BankName:      "SamPay Bank",
			IFSCCode:      "SAMP5917AY",
			AccountType:   accountType,
			Balance:       decimal.Zero,
			Status:        inactiveStatus,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := db.Create(inactiveAccount).Error; err != nil {
			t.Fatalf("failed to create inactive bank account: %v", err)
		}

		newName := "Should Not Update"

		updateRequest := &dto.UpdateBankAccountRequest{
			AccountName: newName,
		}

		updatedBankAccount, err := repo.UpdateBankAccount(
			inactiveAccount.ID,
			updateRequest,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf(
				"expected gorm.ErrRecordNotFound, got %v",
				err,
			)
		}

		if updatedBankAccount != nil {
			t.Errorf(
				"expected nil bank account, got %v",
				updatedBankAccount,
			)
		}
	})
}
