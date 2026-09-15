package services

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCreateVault(t *testing.T) {
	vs := &VaultService{}

	tests := []struct {
		name          string
		request       *dto.CreateVaultRequest
		expectedError string
	}{
		{
			name:          "nil request",
			request:       nil,
			expectedError: "request payload can't be empty",
		},
		{
			name: "invalid status",
			request: &dto.CreateVaultRequest{
				Status: "invalid",
			},
			expectedError: "status is not a valid status",
		},
		{
			name: "invalid vault type",
			request: &dto.CreateVaultRequest{
				Status: constants.VaultStatusActive,
				Type:   "invalid",
			},
			expectedError: "vault type is not a valid type",
		},
		{
			name: "zero balance",
			request: &dto.CreateVaultRequest{
				Status:  constants.VaultStatusActive,
				Type:    constants.PaymentVault,
				Balance: decimal.Zero,
			},
			expectedError: "balance must be greater than 0",
		},
		{
			name: "negative balance",
			request: &dto.CreateVaultRequest{
				Status:  constants.VaultStatusActive,
				Type:    constants.PaymentVault,
				Balance: decimal.NewFromInt(-100),
			},
			expectedError: "balance must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vs.CreateVault(tt.request)

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

func TestGetVaultByID(t *testing.T) {
	vs := &VaultService{}

	_, err := vs.GetVaultByID(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "vault_id is required" {
		t.Fatalf(
			"expected error %q, got %q",
			"vault_id is required",
			err.Error(),
		)
	}
}

func TestUpdateVaultBalance(t *testing.T) {
	vs := &VaultService{}

	tests := []struct {
		name          string
		balance       decimal.Decimal
		expectedError string
	}{
		{
			name:          "zero balance",
			balance:       decimal.Zero,
			expectedError: "balance must be greater than zero",
		},
		{
			name:          "negative balance",
			balance:       decimal.NewFromInt(-100),
			expectedError: "balance must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vs.UpdateVaultBalance(
				uuid.New(),
				tt.balance,
				constants.PaymentVault,
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
