package repositories

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCreateVault(t *testing.T) {
	repo := NewVaultRepository(db)

	vaultType := constants.PaymentVault
	status := constants.VaultStatusActive
	balance := decimal.NewFromInt(10000)

	request := &dto.CreateVaultRequest{
		Type:    vaultType,
		Status:  status,
		Balance: balance,
	}

	result, err := repo.CreateVault(request)

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
	repo := NewVaultRepository(db)

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

	result, err := repo.GetVaults()

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
	repo := NewVaultRepository(db)

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

	result, err := repo.GetVaultByType(vaultType)

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

	result, err = repo.GetVaultByType(nonExistentType)

	if err != nil {
		t.Fatalf("unexpected error for non-existent type: %v", err)
	}

	if result != nil {
		t.Error("expected nil vault for non-existent type")
	}
}

func TestUpdateVaultStatus(t *testing.T) {
	repo := NewVaultRepository(db)

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

	result, err := repo.UpdateVaultStatus(inactiveStatus, vault.ID)

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
