package services

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestValidatePayoutRequestWalletToBank(t *testing.T) {

	tests := []struct {
		name           string
		request        *dto.CreateWalletToBankRequest
		senderMerchant uuid.UUID
		expectedError  string
	}{
		{
			name:           "nil request",
			request:        nil,
			senderMerchant: uuid.New(),
			expectedError:  "payment request is required",
		},
		{
			name: "nil sender merchant id",
			request: &dto.CreateWalletToBankRequest{
				Amount:   decimal.NewFromInt(1000),
				Currency: "INR",
			},
			senderMerchant: uuid.Nil,
			expectedError:  "sender merchant id is required",
		},
		{
			name: "zero amount",
			request: &dto.CreateWalletToBankRequest{
				Amount:   decimal.Zero,
				Currency: "INR",
			},
			senderMerchant: uuid.New(),
			expectedError:  "amount must be greater than 0",
		},
		{
			name: "negative amount",
			request: &dto.CreateWalletToBankRequest{
				Amount:   decimal.NewFromInt(-100),
				Currency: "INR",
			},
			senderMerchant: uuid.New(),
			expectedError:  "amount must be greater than 0",
		},
		{
			name: "invalid currency",
			request: &dto.CreateWalletToBankRequest{
				Amount:   decimal.NewFromInt(1000),
				Currency: "USD",
			},
			senderMerchant: uuid.New(),
			expectedError:  "currency should be INR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testServices.ValidatePayoutRequestWalletToBank(
				tt.request,
				tt.senderMerchant,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if (err).Error() != tt.expectedError {
				t.Fatalf(
					"expected error %q, got %q",
					tt.expectedError,
					(err).Error(),
				)
			}
		})
	}
}

func TestValidatePayoutRequestBankToBank(t *testing.T) {

	tests := []struct {
		name           string
		request        *dto.CreateBankToBankRequest
		senderMerchant uuid.UUID
		expectedError  string
	}{
		{
			name:           "nil request",
			request:        nil,
			senderMerchant: uuid.New(),
			expectedError:  "payment request is required",
		},
		{
			name: "nil sender merchant id",
			request: &dto.CreateBankToBankRequest{
				Amount:   decimal.NewFromInt(1000),
				Currency: "INR",
			},
			senderMerchant: uuid.Nil,
			expectedError:  "sender merchant id is required",
		},
		{
			name: "zero amount",
			request: &dto.CreateBankToBankRequest{
				Amount:   decimal.Zero,
				Currency: "INR",
			},
			senderMerchant: uuid.New(),
			expectedError:  "amount must be greater than 0",
		},
		{
			name: "negative amount",
			request: &dto.CreateBankToBankRequest{
				Amount:   decimal.NewFromInt(-100),
				Currency: "INR",
			},
			senderMerchant: uuid.New(),
			expectedError:  "amount must be greater than 0",
		},
		{
			name: "invalid currency",
			request: &dto.CreateBankToBankRequest{
				Amount:   decimal.NewFromInt(1000),
				Currency: "USD",
			},
			senderMerchant: uuid.New(),
			expectedError:  "currency should be INR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testServices.ValidatePayoutRequestBankToBank(
				tt.request,
				tt.senderMerchant,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if (err).Error() != tt.expectedError {
				t.Fatalf(
					"expected error %q, got %q",
					tt.expectedError,
					(err).Error(),
				)
			}
		})
	}
}

func TestGetPayoutsByMerchantID_InvalidMerchantID(t *testing.T) {

	payouts, err := testServices.GetPayoutsByMerchantID(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if payouts != nil {
		t.Fatalf("expected nil payouts, got %v", payouts)
	}

	if (err).Error() != "merchant_id is required" {
		t.Fatalf(
			"expected error %q, got %q",
			"merchant_id is required",
			(err).Error(),
		)
	}
}

func TestGetPayoutByID_InvalidPayoutID(t *testing.T) {

	payout, err := testServices.GetPayoutByID(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if payout != nil {
		t.Fatalf("expected nil payout, got %v", payout)
	}

	if (err).Error() != "payout_id is required" {
		t.Fatalf(
			"expected error %q, got %q",
			"payout_id is required",
			(err).Error(),
		)
	}
}
