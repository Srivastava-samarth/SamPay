package repositories

import (
	"testing"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCreateWalletForMerchant(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()

	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	result, err := testRepo.CreateWalletForMerchant(merchant.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected wallet, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected generated wallet ID")
	}

	if result.MerchantID != merchant.ID {
		t.Errorf("expected merchant ID %v, got %v", merchant.ID, result.MerchantID)
	}

	if !result.AvailableBalance.Equal(decimal.Zero) {
		t.Errorf("expected available balance 0, got %v", result.AvailableBalance)
	}

	if !result.ReservedBalance.Equal(decimal.Zero) {
		t.Errorf("expected reserved balance 0, got %v", result.ReservedBalance)
	}

	if result.Status == "" {
		t.Fatal("expected status, got nil")
	}

	if result.Status != "active" {
		t.Errorf("expected status active, got %v", result.Status)
	}

	if result.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if result.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestGetWalletByMerchantId(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()

	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	status := "active"

	wallet := &models.Wallets{
		ID:               uuid.New(),
		MerchantID:       merchant.ID,
		AvailableBalance: decimal.NewFromInt(5000),
		ReservedBalance:  decimal.Zero,
		Status:           status,
	}

	if err := db.Create(wallet).Error; err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	result, err := testRepo.GetWalletByMerchantId(merchant.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected wallet, got nil")
	}

	if result.ID != wallet.ID {
		t.Errorf("expected wallet ID %v, got %v", wallet.ID, result.ID)
	}

	if result.MerchantID != merchant.ID {
		t.Errorf("expected merchant ID %v, got %v", merchant.ID, result.MerchantID)
	}

	nonExistentMerchantID := uuid.New()

	result, err = testRepo.GetWalletByMerchantId(nonExistentMerchantID)

	if err != nil {
		t.Fatalf("unexpected error for non-existent merchant: %v", err)
	}

	if result != nil {
		t.Error("expected nil wallet for non-existent merchant")
	}
}

func TestUpdateWalletStatus(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()

	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	activeStatus := "active"

	wallet := &models.Wallets{
		ID:               uuid.New(),
		MerchantID:       merchant.ID,
		AvailableBalance: decimal.NewFromInt(5000),
		ReservedBalance:  decimal.Zero,
		Status:           activeStatus,
	}

	if err := db.Create(wallet).Error; err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	newStatus := "inactive"

	result, err := testRepo.UpdateWalletStatus(merchant.ID, newStatus)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected wallet, got nil")
	}

	if result.ID != wallet.ID {
		t.Errorf("expected wallet ID %v, got %v", wallet.ID, result.ID)
	}

	if result.Status == "" {
		t.Fatal("expected wallet status, got nil")
	}

	if result.Status != newStatus {
		t.Errorf("expected status %v, got %v", newStatus, result.Status)
	}
}

func TestUpdateWalletBalance(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()

	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	status := "active"

	wallet := &models.Wallets{
		ID:               uuid.New(),
		MerchantID:       merchant.ID,
		AvailableBalance: decimal.NewFromInt(5000),
		ReservedBalance:  decimal.NewFromInt(1000),
		Status:           status,
	}

	if err := db.Create(wallet).Error; err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	request := &dto.UpdateWalletBalanceRequest{
		AvailableBalance: decimal.NewFromInt(8000),
		ReservedBalance:  decimal.NewFromInt(2000),
	}

	result, err := testRepo.UpdateWalletBalance(merchant.ID, request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected wallet, got nil")
	}

	if result.ID != wallet.ID {
		t.Errorf("expected wallet ID %v, got %v", wallet.ID, result.ID)
	}

	if !result.AvailableBalance.Equal(request.AvailableBalance) {
		t.Errorf(
			"expected available balance %v, got %v",
			request.AvailableBalance,
			result.AvailableBalance,
		)
	}

	if !result.ReservedBalance.Equal(request.ReservedBalance) {
		t.Errorf(
			"expected reserved balance %v, got %v",
			request.ReservedBalance,
			result.ReservedBalance,
		)
	}

	request = &dto.UpdateWalletBalanceRequest{
		AvailableBalance: decimal.Zero,
		ReservedBalance:  decimal.Zero,
	}

	result, err = testRepo.UpdateWalletBalance(merchant.ID, request)

	if err != nil {
		t.Fatalf("unexpected error when updating with zero values: %v", err)
	}

	if result.AvailableBalance.Equal(decimal.Zero) {
		t.Error("expected available balance to remain unchanged")
	}

	if result.ReservedBalance.Equal(decimal.Zero) {
		t.Error("expected reserved balance to remain unchanged")
	}
}
