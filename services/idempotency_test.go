package services

import (
	"testing"

	"github.com/google/uuid"
)

func TestGetIdempotencyByKeyAndMerchantId(t *testing.T) {
	is := &IdempotencyService{}

	validMerchantID := uuid.New()
	validKey := "payment-123"

	tests := []struct {
		name       string
		merchantID uuid.UUID
		key        string
		wantErr    string
	}{
		{
			name:       "nil merchant id",
			merchantID: uuid.Nil,
			key:        validKey,
			wantErr:    "merchantID is required",
		},
		{
			name:       "empty key",
			merchantID: validMerchantID,
			key:        "",
			wantErr:    "key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := is.GetIdempotencyByKeyAndMerchantId(
				tt.merchantID,
				tt.key,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.wantErr {
				t.Errorf(
					"expected error %q, got %q",
					tt.wantErr,
					err.Error(),
				)
			}
		})
	}
}

func TestCreateIdempotencyKey(t *testing.T) {
	is := &IdempotencyService{}

	validMerchantID := uuid.New()
	validKey := "payment-123"

	tests := []struct {
		name         string
		merchantID   uuid.UUID
		key          string
		responseBody []byte
		wantErr      string
	}{
		{
			name:       "nil merchant id",
			merchantID: uuid.Nil,
			key:        validKey,
			wantErr:    "merchantID is required",
		},
		{
			name:       "empty key",
			merchantID: validMerchantID,
			key:        "",
			wantErr:    "key is required",
		},
		{
			name:       "whitespace key",
			merchantID: validMerchantID,
			key:        " ",
			wantErr:    "key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := is.CreateIdempotencyKey(
				tt.merchantID,
				tt.key,
				tt.responseBody,
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.wantErr {
				t.Errorf(
					"expected error %q, got %q",
					tt.wantErr,
					err.Error(),
				)
			}
		})
	}
}
