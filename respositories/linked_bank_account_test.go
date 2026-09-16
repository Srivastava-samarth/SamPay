package repositories

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/google/uuid"
)

func TestCreateLinkedBankAccount(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()

	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	bankAccount := testutils.GenerateTestBankAccount()

	if err := db.Create(bankAccount).Error; err != nil {
		t.Fatalf("failed to create bank account: %v", err)
	}

	accountType := constants.BankAccountTypePrimary
	status := constants.MerchantStatusActive

	linkedBankAccount := &models.LinkedBankAccount{
		MerchantID:    merchant.ID,
		BankAccountID: bankAccount.ID,
		Type:          accountType,
		Status:        status,
	}

	result, err := testRepo.CreateLinkedBankAccount(linkedBankAccount)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected linked bank account, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected generated ID, got nil UUID")
	}

	if result.MerchantID != merchant.ID {
		t.Errorf(
			"expected merchant ID %v, got %v",
			merchant.ID,
			result.MerchantID,
		)
	}

	if result.BankAccountID != bankAccount.ID {
		t.Errorf(
			"expected bank account ID %v, got %v",
			bankAccount.ID,
			result.BankAccountID,
		)
	}

	if result.Type != accountType {
		t.Errorf(
			"expected type %v, got %v",
			accountType,
			result.Type,
		)
	}

	if result.Status != status {
		t.Errorf(
			"expected status %v, got %v",
			status,
			result.Status,
		)
	}

	if result.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if result.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}

	var persisted models.LinkedBankAccount

	errDB := db.
		Where("id = ?", result.ID).
		First(&persisted).Error

	if errDB != nil {
		t.Fatalf("failed to fetch persisted linked bank account: %v", errDB)
	}

	if persisted.MerchantID != merchant.ID {
		t.Errorf(
			"expected persisted merchant ID %v, got %v",
			merchant.ID,
			persisted.MerchantID,
		)
	}

	if persisted.BankAccountID != bankAccount.ID {
		t.Errorf(
			"expected persisted bank account ID %v, got %v",
			bankAccount.ID,
			persisted.BankAccountID,
		)
	}
}

func TestGetAllBankAccountLinkedByMerchantID(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()
	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	otherMerchant := testutils.GenerateTestMerchant()
	if err := db.Create(otherMerchant).Error; err != nil {
		t.Fatalf("failed to create other merchant: %v", err)
	}

	bankAccount1 := testutils.GenerateTestBankAccount()
	if err := db.Create(bankAccount1).Error; err != nil {
		t.Fatalf("failed to create bank account 1: %v", err)
	}

	bankAccount2 := testutils.GenerateTestBankAccount()
	if err := db.Create(bankAccount2).Error; err != nil {
		t.Fatalf("failed to create bank account 2: %v", err)
	}

	bankAccount3 := testutils.GenerateTestBankAccount()
	if err := db.Create(bankAccount3).Error; err != nil {
		t.Fatalf("failed to create bank account 3: %v", err)
	}

	primaryType := constants.BankAccountTypePrimary
	secondaryType := "secondary"
	activeStatus := constants.MerchantStatusActive
	inactiveStatus := "inactive"

	activeLinkedAccount1 := &models.LinkedBankAccount{
		ID:            uuid.New(),
		MerchantID:    merchant.ID,
		BankAccountID: bankAccount1.ID,
		Type:          primaryType,
		Status:        activeStatus,
	}

	activeLinkedAccount2 := &models.LinkedBankAccount{
		ID:            uuid.New(),
		MerchantID:    merchant.ID,
		BankAccountID: bankAccount2.ID,
		Type:          secondaryType,
		Status:        activeStatus,
	}

	inactiveLinkedAccount := &models.LinkedBankAccount{
		ID:            uuid.New(),
		MerchantID:    merchant.ID,
		BankAccountID: bankAccount3.ID,
		Type:          secondaryType,
		Status:        inactiveStatus,
	}

	otherMerchantLinkedAccount := &models.LinkedBankAccount{
		ID:            uuid.New(),
		MerchantID:    otherMerchant.ID,
		BankAccountID: bankAccount3.ID,
		Type:          primaryType,
		Status:        activeStatus,
	}

	accounts := []*models.LinkedBankAccount{
		activeLinkedAccount1,
		activeLinkedAccount2,
		inactiveLinkedAccount,
		otherMerchantLinkedAccount,
	}

	for _, account := range accounts {
		if err := db.Create(account).Error; err != nil {
			t.Fatalf("failed to create linked bank account: %v", err)
		}
	}

	result, err := testRepo.GetAllBankAccountLinkedByMerchantID(merchant.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 active linked bank accounts, got %d", len(result))
	}

	foundActive1 := false
	foundActive2 := false

	for _, account := range result {
		if account.ID == activeLinkedAccount1.ID {
			foundActive1 = true
		}

		if account.ID == activeLinkedAccount2.ID {
			foundActive2 = true
		}

		if account.ID == inactiveLinkedAccount.ID {
			t.Errorf("inactive linked bank account was returned")
		}

		if account.ID == otherMerchantLinkedAccount.ID {
			t.Errorf("linked bank account of another merchant was returned")
		}

		if account.Status != activeStatus {
			t.Errorf(
				"expected active status, got %v",
				account.Status,
			)
		}

		if account.MerchantID != merchant.ID {
			t.Errorf(
				"expected merchant ID %v, got %v",
				merchant.ID,
				account.MerchantID,
			)
		}
	}

	if !foundActive1 {
		t.Errorf("expected linked bank account %v in result", activeLinkedAccount1.ID)
	}

	if !foundActive2 {
		t.Errorf("expected linked bank account %v in result", activeLinkedAccount2.ID)
	}
}

func TestUpdateLinkedBankAccount(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()
	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	bankAccount := testutils.GenerateTestBankAccount()
	if err := db.Create(bankAccount).Error; err != nil {
		t.Fatalf("failed to create bank account: %v", err)
	}

	originalType := constants.BankAccountTypePrimary
	originalStatus := constants.MerchantStatusActive

	linkedBankAccount := &models.LinkedBankAccount{
		ID:            uuid.New(),
		MerchantID:    merchant.ID,
		BankAccountID: bankAccount.ID,
		Type:          originalType,
		Status:        originalStatus,
	}

	if err := db.Create(linkedBankAccount).Error; err != nil {
		t.Fatalf("failed to create linked bank account: %v", err)
	}

	newType := "secondary"
	newStatus := "inactive"

	request := &dto.UpdateLinkedBankAccountRequest{
		Type:   newType,
		Status: newStatus,
	}

	result, err := testRepo.UpdateLinkedBankAccount(
		bankAccount.ID,
		request,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected updated linked bank account, got nil")
	}

	if result.ID != linkedBankAccount.ID {
		t.Errorf(
			"expected ID %v, got %v",
			linkedBankAccount.ID,
			result.ID,
		)
	}

	if result.MerchantID != merchant.ID {
		t.Errorf(
			"expected merchant ID %v, got %v",
			merchant.ID,
			result.MerchantID,
		)
	}

	if result.BankAccountID != bankAccount.ID {
		t.Errorf(
			"expected bank account ID %v, got %v",
			bankAccount.ID,
			result.BankAccountID,
		)
	}

	if result.Type != newType {
		t.Errorf(
			"expected type %v, got %v",
			newType,
			result.Type,
		)
	}

	if result.Status != newStatus {
		t.Errorf(
			"expected status %v, got %v",
			newStatus,
			result.Status,
		)
	}

	var persisted models.LinkedBankAccount

	if err := db.
		Where("id = ?", linkedBankAccount.ID).
		First(&persisted).Error; err != nil {
		t.Fatalf("failed to fetch persisted linked bank account: %v", err)
	}

	if persisted.Type != newType {
		t.Errorf(
			"expected persisted type %v, got %v",
			newType,
			persisted.Type,
		)
	}

	if persisted.Status != newStatus {
		t.Errorf(
			"expected persisted status %v, got %v",
			newStatus,
			persisted.Status,
		)
	}
}

func TestGetPrimaryBankAccountLinkedByMerchantID(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()
	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	otherMerchant := testutils.GenerateTestMerchant()
	if err := db.Create(otherMerchant).Error; err != nil {
		t.Fatalf("failed to create other merchant: %v", err)
	}

	bankAccount1 := testutils.GenerateTestBankAccount()
	if err := db.Create(bankAccount1).Error; err != nil {
		t.Fatalf("failed to create bank account 1: %v", err)
	}

	bankAccount2 := testutils.GenerateTestBankAccount()
	if err := db.Create(bankAccount2).Error; err != nil {
		t.Fatalf("failed to create bank account 2: %v", err)
	}

	bankAccount3 := testutils.GenerateTestBankAccount()
	if err := db.Create(bankAccount3).Error; err != nil {
		t.Fatalf("failed to create bank account 3: %v", err)
	}

	activeStatus := constants.MerchantStatusActive
	primaryType := constants.BankAccountTypePrimary
	secondaryType := "secondary"

	primaryLinkedAccount := &models.LinkedBankAccount{
		ID:            uuid.New(),
		MerchantID:    merchant.ID,
		BankAccountID: bankAccount1.ID,
		Type:          primaryType,
		Status:        activeStatus,
	}

	secondaryLinkedAccount := &models.LinkedBankAccount{
		ID:            uuid.New(),
		MerchantID:    merchant.ID,
		BankAccountID: bankAccount2.ID,
		Type:          secondaryType,
		Status:        activeStatus,
	}

	otherMerchantPrimary := &models.LinkedBankAccount{
		ID:            uuid.New(),
		MerchantID:    otherMerchant.ID,
		BankAccountID: bankAccount3.ID,
		Type:          primaryType,
		Status:        activeStatus,
	}

	accounts := []*models.LinkedBankAccount{
		primaryLinkedAccount,
		secondaryLinkedAccount,
		otherMerchantPrimary,
	}

	for _, account := range accounts {
		if err := db.Create(account).Error; err != nil {
			t.Fatalf("failed to create linked bank account: %v", err)
		}
	}

	result, err := testRepo.GetPrimaryBankAccountLinkedByMerchantID(merchant.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected primary linked bank account, got nil")
	}

	if result.ID != primaryLinkedAccount.ID {
		t.Errorf(
			"expected linked bank account ID %v, got %v",
			primaryLinkedAccount.ID,
			result.ID,
		)
	}

	if result.MerchantID != merchant.ID {
		t.Errorf(
			"expected merchant ID %v, got %v",
			merchant.ID,
			result.MerchantID,
		)
	}

	if result.BankAccountID != bankAccount1.ID {
		t.Errorf(
			"expected bank account ID %v, got %v",
			bankAccount1.ID,
			result.BankAccountID,
		)
	}

	if result.Type != primaryType {
		t.Errorf(
			"expected type %v, got %v",
			primaryType,
			result.Type,
		)
	}

	if result.Status != activeStatus {
		t.Errorf(
			"expected status %v, got %v",
			activeStatus,
			result.Status,
		)
	}
}

func TestGetBankAccountLinkedByID(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()
	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	activeBankAccount := testutils.GenerateTestBankAccount()
	if err := db.Create(activeBankAccount).Error; err != nil {
		t.Fatalf("failed to create active bank account: %v", err)
	}

	inactiveBankAccount := testutils.GenerateTestBankAccount()
	if err := db.Create(inactiveBankAccount).Error; err != nil {
		t.Fatalf("failed to create inactive bank account: %v", err)
	}

	activeStatus := constants.MerchantStatusActive
	inactiveStatus := "inactive"
	primaryType := constants.BankAccountTypePrimary
	secondaryType := "secondary"

	activeLinkedAccount := &models.LinkedBankAccount{
		ID:            uuid.New(),
		MerchantID:    merchant.ID,
		BankAccountID: activeBankAccount.ID,
		Type:          primaryType,
		Status:        activeStatus,
	}

	inactiveLinkedAccount := &models.LinkedBankAccount{
		ID:            uuid.New(),
		MerchantID:    merchant.ID,
		BankAccountID: inactiveBankAccount.ID,
		Type:          secondaryType,
		Status:        inactiveStatus,
	}

	if err := db.Create(activeLinkedAccount).Error; err != nil {
		t.Fatalf("failed to create active linked bank account: %v", err)
	}

	if err := db.Create(inactiveLinkedAccount).Error; err != nil {
		t.Fatalf("failed to create inactive linked bank account: %v", err)
	}

	t.Run("active linked bank account", func(t *testing.T) {
		result, err := testRepo.GetBankAccountLinkedByID(activeBankAccount.ID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected linked bank account, got nil")
		}

		if result.ID != activeLinkedAccount.ID {
			t.Errorf(
				"expected linked bank account ID %v, got %v",
				activeLinkedAccount.ID,
				result.ID,
			)
		}

		if result.BankAccountID != activeBankAccount.ID {
			t.Errorf(
				"expected bank account ID %v, got %v",
				activeBankAccount.ID,
				result.BankAccountID,
			)
		}

		if result.MerchantID != merchant.ID {
			t.Errorf(
				"expected merchant ID %v, got %v",
				merchant.ID,
				result.MerchantID,
			)
		}

		if result.Status != activeStatus {
			t.Errorf(
				"expected status %v, got %v",
				activeStatus,
				result.Status,
			)
		}
	})

	t.Run("inactive linked bank account", func(t *testing.T) {
		result, err := testRepo.GetBankAccountLinkedByID(inactiveBankAccount.ID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != nil && result.ID == inactiveLinkedAccount.ID {
			t.Errorf("inactive linked bank account should not be returned")
		}
	})

	t.Run("non existing bank account", func(t *testing.T) {
		nonExistingID := uuid.New()

		_, err := testRepo.GetBankAccountLinkedByID(nonExistingID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
