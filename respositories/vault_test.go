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

func cleanTestDB(t *testing.T) {
	t.Helper()

	if err := testutils.CleanupTestDB(db); err != nil {
		t.Fatalf("failed to clean test DB: %v", err)
	}
}

func TestCreateVault(t *testing.T) {
	cleanTestDB(t)

	vaultType := constants.PaymentVault
	status := constants.VaultStatusActive
	balance := decimal.NewFromInt(10000)

	request := &dto.CreateVaultRequest{
		Type:    vaultType,
		Status:  status,
		Balance: balance,
	}

	result, err := testRepo.CreateVault(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected vault, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected generated ID")
	}

	if result.Type != request.Type {
		t.Error("expected vault type to match")
	}

	if result.Status != request.Status {
		t.Error("expected vault status to match")
	}

	if !result.Balance.Equal(balance) {
		t.Errorf("expected balance %v, got %v", balance, result.Balance)
	}

	if result.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if result.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestGetVaults(t *testing.T) {
	cleanTestDB(t)

	vaultType1 := constants.PaymentVault
	vaultType2 := constants.CompanyVault
	status := constants.VaultStatusActive

	vault1 := &models.Vault{
		ID:      uuid.New(),
		Type:    vaultType1,
		Status:  status,
		Balance: decimal.NewFromInt(10000),
	}

	vault2 := &models.Vault{
		ID:      uuid.New(),
		Type:    vaultType2,
		Status:  status,
		Balance: decimal.NewFromInt(20000),
	}

	if err := db.Create(vault1).Error; err != nil {
		t.Fatalf("failed to create vault1: %v", err)
	}

	if err := db.Create(vault2).Error; err != nil {
		t.Fatalf("failed to create vault2: %v", err)
	}

	result, err := testRepo.GetVaults()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected vaults, got nil")
	}

	foundVault1 := false
	foundVault2 := false

	for _, vault := range result {
		if vault.ID == vault1.ID {
			foundVault1 = true
		}

		if vault.ID == vault2.ID {
			foundVault2 = true
		}
	}

	if !foundVault1 {
		t.Error("expected vault1 to be returned")
	}

	if !foundVault2 {
		t.Error("expected vault2 to be returned")
	}
}

func TestGetVaultByType(t *testing.T) {
	cleanTestDB(t)

	vaultType := constants.PaymentVault
	status := constants.VaultStatusActive

	vault := &models.Vault{
		ID:      uuid.New(),
		Type:    vaultType,
		Status:  status,
		Balance: decimal.NewFromInt(10000),
	}

	if err := db.Create(vault).Error; err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	result, err := testRepo.GetVaultByType(vaultType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected vault, got nil")
	}

	if result.ID != vault.ID {
		t.Errorf("expected vault ID %v, got %v", vault.ID, result.ID)
	}

	nonExistentType := "non_existent"

	result, err = testRepo.GetVaultByType(nonExistentType)
	if err != nil {
		t.Fatalf("unexpected error for non-existent type: %v", err)
	}

	if result != nil {
		t.Error("expected nil vault for non-existent type")
	}
}

func TestUpdateVaultStatus(t *testing.T) {
	cleanTestDB(t)

	vaultType := constants.PaymentVault
	activeStatus := constants.VaultStatusActive
	inactiveStatus := constants.VaultStatusInactive

	vault := &models.Vault{
		ID:      uuid.New(),
		Type:    vaultType,
		Status:  activeStatus,
		Balance: decimal.NewFromInt(10000),
	}

	if err := db.Create(vault).Error; err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	result, err := testRepo.UpdateVaultStatus(inactiveStatus, vault.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected vault, got nil")
	}

	if result.ID != vault.ID {
		t.Errorf("expected vault ID %v, got %v", vault.ID, result.ID)
	}

	if result.Status == "" {
		t.Fatal("expected vault status, got nil")
	}

	if result.Status != inactiveStatus {
		t.Errorf("expected status %v, got %v", inactiveStatus, result.Status)
	}
}

func TestGetVault(t *testing.T) {
	cleanTestDB(t)

	vault := &models.Vault{
		ID:      uuid.New(),
		Type:    constants.PaymentVault,
		Status:  constants.VaultStatusActive,
		Balance: decimal.NewFromInt(10000),
	}

	if err := db.Create(vault).Error; err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	result, err := testRepo.GetVault(vault.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected vault, got nil")
	}

	if result.ID != vault.ID {
		t.Errorf("expected vault ID %v, got %v", vault.ID, result.ID)
	}

	if result.Type != vault.Type {
		t.Errorf("expected vault type %v, got %v", vault.Type, result.Type)
	}

	if result.Status != vault.Status {
		t.Errorf("expected vault status %v, got %v", vault.Status, result.Status)
	}

	if !result.Balance.Equal(vault.Balance) {
		t.Errorf("expected balance %v, got %v", vault.Balance, result.Balance)
	}

	nonExistentID := uuid.New()

	result, err = testRepo.GetVault(nonExistentID)
	if err == nil {
		t.Fatal("expected error for non-existent vault")
	}

	if result != nil {
		t.Error("expected nil vault for non-existent ID")
	}
}

func TestUpdateVaultBalance(t *testing.T) {
	cleanTestDB(t)

	vaultType := constants.PaymentVault
	status := constants.VaultStatusActive
	initialBalance := decimal.NewFromInt(10000)
	updatedBalance := decimal.NewFromInt(25000)

	vault := &models.Vault{
		ID:      uuid.New(),
		Type:    vaultType,
		Status:  status,
		Balance: initialBalance,
	}

	if err := db.Create(vault).Error; err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	result, err := testRepo.UpdateVaultBalance(updatedBalance, vaultType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected vault, got nil")
	}

	if result.ID != vault.ID {
		t.Errorf("expected vault ID %v, got %v", vault.ID, result.ID)
	}

	if !result.Balance.Equal(updatedBalance) {
		t.Errorf(
			"expected balance %v, got %v",
			updatedBalance,
			result.Balance,
		)
	}

	// Verify the persisted value in the database.
	var persistedVault models.Vault
	if err := db.Where("id = ?", vault.ID).First(&persistedVault).Error; err != nil {
		t.Fatalf("failed to fetch persisted vault: %v", err)
	}

	if !persistedVault.Balance.Equal(updatedBalance) {
		t.Errorf(
			"expected persisted balance %v, got %v",
			updatedBalance,
			persistedVault.Balance,
		)
	}

	// Invalid balance should return the exact validation error.
	invalidBalance := decimal.Zero

	result, err = testRepo.UpdateVaultBalance(invalidBalance, vaultType)
	if err == nil {
		t.Fatal("expected error for zero balance")
	}

	if err.Error() != "balance should be greater than zero" {
		t.Errorf(
			"expected error %q, got %q",
			"balance should be greater than zero",
			err.Error(),
		)
	}

	if result != nil {
		t.Error("expected nil vault for invalid balance")
	}
}
