package services

import (
	"testing"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCreateWalletForMerchant(t *testing.T) {

	_, err := testServices.CreateWalletForMerchant(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "merchant id is required" {
		t.Fatalf(
			"expected error %q, got %q",
			"merchant id is required",
			err.Error(),
		)
	}
}

func TestGetWalletByMerchantID(t *testing.T) {

	_, err := testServices.GetWalletByMerchantID(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "merchant id is required" {
		t.Fatalf(
			"expected error %q, got %q",
			"merchant id is required",
			err.Error(),
		)
	}
}

func TestUpdateWalletStatus(t *testing.T) {

	_, err := testServices.UpdateWalletStatus(uuid.Nil, "")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "merchant id is required" {
		t.Fatalf(
			"expected error %q, got %q",
			"merchant id is required",
			err.Error(),
		)
	}
}

func TestTopUpWalletFromPrimaryBank(t *testing.T) {

	tests := []struct {
		name          string
		wallet        *models.Wallets
		amount        decimal.Decimal
		expectedError string
	}{
		{
			name:          "zero amount",
			wallet:        &models.Wallets{},
			amount:        decimal.Zero,
			expectedError: "amount should be greater than 0",
		},
		{
			name:          "negative amount",
			wallet:        &models.Wallets{},
			amount:        decimal.NewFromInt(-100),
			expectedError: "amount should be greater than 0",
		},
		{
			name:          "nil wallet",
			wallet:        nil,
			amount:        decimal.NewFromInt(100),
			expectedError: "wallet not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testServices.TopUpWalletFromPrimaryBank(
				nil,
				tt.wallet,
				tt.amount,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Fatalf(
					"expected error %q, got %q",
					tt.expectedError,
					err.Error(),
				)
			}
		})
	}
}
