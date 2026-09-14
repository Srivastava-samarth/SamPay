package services

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestValidateCreateRefundRequest(t *testing.T) {
	rs := &RefundService{}

	tests := []struct {
		name          string
		request       *dto.CreateRefundRequest
		expectedError string
	}{
		{
			name:          "nil request",
			request:       nil,
			expectedError: "request is required",
		},
		{
			name: "nil payment id",
			request: &dto.CreateRefundRequest{
				PaymentID: uuid.Nil,
			},
			expectedError: "payment id not provided",
		},
		{
			name: "nil merchant id",
			request: &dto.CreateRefundRequest{
				PaymentID:  uuid.New(),
				MerchantID: uuid.Nil,
			},
			expectedError: "merchant id not provided",
		},
		{
			name: "zero amount",
			request: &dto.CreateRefundRequest{
				PaymentID:  uuid.New(),
				MerchantID: uuid.New(),
				Amount:     decimal.Zero,
			},
			expectedError: "amount must be greater than 0",
		},
		{
			name: "negative amount",
			request: &dto.CreateRefundRequest{
				PaymentID:  uuid.New(),
				MerchantID: uuid.New(),
				Amount:     decimal.NewFromInt(-100),
			},
			expectedError: "amount must be greater than 0",
		},
		{
			name: "nil reason",
			request: &dto.CreateRefundRequest{
				PaymentID:  uuid.New(),
				MerchantID: uuid.New(),
				Amount:     decimal.NewFromInt(100),
				Reason:     nil,
			},
			expectedError: "reason is necessary",
		},
		{
			name: "empty reason",
			request: func() *dto.CreateRefundRequest {
				reason := ""
				return &dto.CreateRefundRequest{
					PaymentID:  uuid.New(),
					MerchantID: uuid.New(),
					Amount:     decimal.NewFromInt(100),
					Reason:     &reason,
				}
			}(),
			expectedError: "reason is necessary",
		},
		{
			name: "whitespace reason",
			request: func() *dto.CreateRefundRequest {
				reason := "   "
				return &dto.CreateRefundRequest{
					PaymentID:  uuid.New(),
					MerchantID: uuid.New(),
					Amount:     decimal.NewFromInt(100),
					Reason:     &reason,
				}
			}(),
			expectedError: "reason is necessary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rs.ValidateCreateRefundRequest(tt.request)

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

func TestGetRefundByReference(t *testing.T) {
	rs := &RefundService{}

	refundRef := ""

	_, err := rs.GetRefundByReference(refundRef)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "refund reference is required" {
		t.Fatalf("expected error %q, got %q",
			"refund reference is required",
			err.Error(),
		)
	}
}

func TestGetRefundsByMerchantId(t *testing.T) {
	rs := &RefundService{}

	_, err := rs.GetRefundsByMerchantId(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "merchant id is required" {
		t.Fatalf("expected error %q, got %q",
			"merchant id is required",
			err.Error(),
		)
	}
}

func TestGetRefundById(t *testing.T) {
	rs := &RefundService{}

	_, err := rs.GetRefundById(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "refund id is required" {
		t.Fatalf("expected error %q, got %q",
			"refund id is required",
			err.Error(),
		)
	}
}
