package services

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestPaymentService_CalculateFees(t *testing.T) {
	ps := &PaymentService{}

	tests := []struct {
		name        string
		amount      decimal.Decimal
		expectedFee decimal.Decimal
		wantErr     bool
	}{
		{
			name:        "calculates 1 percent fee",
			amount:      decimal.NewFromInt(100),
			expectedFee: decimal.NewFromInt(1),
		},
		{
			name:        "calculates fee for decimal amount",
			amount:      decimal.NewFromFloat(123.45),
			expectedFee: decimal.NewFromFloat(1.2345),
		},
		{
			name:    "zero amount returns error",
			amount:  decimal.Zero,
			wantErr: true,
		},
		{
			name:    "negative amount returns error",
			amount:  decimal.NewFromInt(-100),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fee, err := ps.CalculateFees(tt.amount)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !fee.Equal(tt.expectedFee) {
				t.Fatalf(
					"expected fee %s, got %s",
					tt.expectedFee,
					fee,
				)
			}
		})
	}
}

func TestPaymentService_ValidatePaymentRequest_InvalidInput(t *testing.T) {
	ps := &PaymentService{}

	validSenderID := uuid.New()
	validReceiverID := uuid.New()

	tests := []struct {
		name        string
		request     *dto.CreatePaymentRequest
		senderID    uuid.UUID
		expectedErr string
	}{
		{
			name:        "nil request",
			request:     nil,
			senderID:    validSenderID,
			expectedErr: "payment request is required",
		},
		{
			name: "nil sender merchant",
			request: &dto.CreatePaymentRequest{
				ReceiverMerchantID: validReceiverID,
				Amount:             decimal.NewFromInt(100),
				Currency:           "INR",
			},
			senderID:    uuid.Nil,
			expectedErr: "sender merchant id is required",
		},
		{
			name: "nil receiver merchant",
			request: &dto.CreatePaymentRequest{
				ReceiverMerchantID: uuid.Nil,
				Amount:             decimal.NewFromInt(100),
				Currency:           "INR",
			},
			senderID:    validSenderID,
			expectedErr: "receiver merchant id is required",
		},
		{
			name: "sender and receiver are same",
			request: &dto.CreatePaymentRequest{
				ReceiverMerchantID: validSenderID,
				Amount:             decimal.NewFromInt(100),
				Currency:           "INR",
			},
			senderID:    validSenderID,
			expectedErr: "sender and receiver can't be same",
		},
		{
			name: "zero amount",
			request: &dto.CreatePaymentRequest{
				ReceiverMerchantID: validReceiverID,
				Amount:             decimal.Zero,
				Currency:           "INR",
			},
			senderID:    validSenderID,
			expectedErr: "amount must be greater than 0",
		},
		{
			name: "negative amount",
			request: &dto.CreatePaymentRequest{
				ReceiverMerchantID: validReceiverID,
				Amount:             decimal.NewFromInt(-100),
				Currency:           "INR",
			},
			senderID:    validSenderID,
			expectedErr: "amount must be greater than 0",
		},
		{
			name: "invalid currency",
			request: &dto.CreatePaymentRequest{
				ReceiverMerchantID: validReceiverID,
				Amount:             decimal.NewFromInt(100),
				Currency:           "USD",
			},
			senderID:    validSenderID,
			expectedErr: "currency should be INR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ps.ValidatePaymentRequest(tt.request, tt.senderID)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedErr {
				t.Fatalf(
					"expected error %q, got %q",
					tt.expectedErr,
					err.Error(),
				)
			}
		})
	}
}

func TestGetPaymentsByMerchantID_InvalidMerchantID(t *testing.T) {
	ps := &PaymentService{}

	merchantID := uuid.Nil

	payments, err := ps.GetPaymentsByMerchantID(merchantID)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if payments != nil {
		t.Fatalf("expected nil payments, got %v", payments)
	}

	if (err).Error() != "merchant_id is required" {
		t.Fatalf("expected error %q, got %q", "merchant_id is required", (err).Error())
	}
}

func TestGetPaymentByID_InvalidPaymentID(t *testing.T) {
	ps := &PaymentService{}

	paymentID := uuid.Nil

	payment, err := ps.GetPaymentByID(paymentID)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if payment != nil {
		t.Fatalf("expected nil payment, got %v", payment)
	}

	if (err).Error() != "payment_id is required" {
		t.Fatalf("expected error %q, got %q", "payment_id is required", (err).Error())
	}
}
