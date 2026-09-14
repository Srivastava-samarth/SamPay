package services

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

func TestCreateMerchantUser(t *testing.T) {
	mus := &MerchantUserService{}

	validMerchantID := uuid.New()
	validUserID := uuid.New()

	tests := []struct {
		name          string
		request       *dto.CreateMerchantUserRequest
		expectedError string
	}{
		{
			name:          "nil request",
			request:       nil,
			expectedError: "request is required",
		},
		{
			name: "nil merchant id",
			request: &dto.CreateMerchantUserRequest{
				MerchantID: uuid.Nil,
				UserID:     validUserID,
				Role:       "owner",
			},
			expectedError: "merchant_id is required",
		},
		{
			name: "nil user id",
			request: &dto.CreateMerchantUserRequest{
				MerchantID: validMerchantID,
				UserID:     uuid.Nil,
				Role:       "owner",
			},
			expectedError: "user_id is required",
		},
		{
			name: "empty role",
			request: &dto.CreateMerchantUserRequest{
				MerchantID: validMerchantID,
				UserID:     validUserID,
				Role:       "",
			},
			expectedError: "role is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mus.CreateMerchantUser(tt.request)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Fatalf(
					"expected %q, got %q",
					tt.expectedError,
					err.Error(),
				)
			}
		})
	}
}
