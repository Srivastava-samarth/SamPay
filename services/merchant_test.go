package services

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

func TestGetMerchantByID(t *testing.T) {
	ms := &MerchantService{}

	_, err := ms.GetMerchantByID(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "merchant_id is required" {
		t.Fatalf("expected %q, got %q", "merchant_id is required", err.Error())
	}
}

func TestCreateInitialMerchant(t *testing.T) {
	ms := &MerchantService{}

	validRequest := &dto.CreateMerchantOnboardingRequest{
		MerchantName: "Test Merchant",
		Email:        "test@example.com",
		MerchantType: "individual",
		PhoneNumber:  "9876543210",
	}

	tests := []struct {
		name          string
		request       *dto.CreateMerchantOnboardingRequest
		expectedError string
	}{
		{
			name:          "nil request",
			request:       nil,
			expectedError: "request can't be empty",
		},
		{
			name: "missing merchant name",
			request: func() *dto.CreateMerchantOnboardingRequest {
				r := *validRequest
				r.MerchantName = ""
				return &r
			}(),
			expectedError: "merchant name is required",
		},
		{
			name: "missing email",
			request: func() *dto.CreateMerchantOnboardingRequest {
				r := *validRequest
				r.Email = ""
				return &r
			}(),
			expectedError: "email is required",
		},
		{
			name: "missing merchant type",
			request: func() *dto.CreateMerchantOnboardingRequest {
				r := *validRequest
				r.MerchantType = ""
				return &r
			}(),
			expectedError: "merchant type is required",
		},
		{
			name: "missing phone number",
			request: func() *dto.CreateMerchantOnboardingRequest {
				r := *validRequest
				r.PhoneNumber = ""
				return &r
			}(),
			expectedError: "phone number is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ms.CreateInitialMerchant(tt.request)

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

func TestUpdateMerchantCompliance(t *testing.T) {
	ms := &MerchantService{}

	tests := []struct {
		name          string
		merchantID    uuid.UUID
		response      *dto.ComplianceCheckResponse
		expectedError string
	}{
		{
			name:          "nil merchant id",
			merchantID:    uuid.Nil,
			response:      &dto.ComplianceCheckResponse{},
			expectedError: "merchant_id is required",
		},
		{
			name:          "nil compliance response",
			merchantID:    uuid.New(),
			response:      nil,
			expectedError: "compliance response is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ms.UpdateMerchantCompliance(
				tt.merchantID,
				tt.response,
			)

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

func TestValidateMerchantOnboardingRequest(t *testing.T) {
	ms := &MerchantService{}

	tests := []struct {
		name          string
		request       *dto.CreateMerchantOnboardingRequest
		expectedError string
	}{
		{
			name: "individual without individual details",
			request: &dto.CreateMerchantOnboardingRequest{
				MerchantType: "individual",
			},
			expectedError: "individual details are required",
		},
		{
			name: "company without company details",
			request: &dto.CreateMerchantOnboardingRequest{
				MerchantType: "company",
			},
			expectedError: "company details are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ms.ValidateMerchantOnboardingRequest(tt.request)

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

func TestUpdateMerchant(t *testing.T) {
	ms := &MerchantService{}

	validRequest := &dto.UpdateMerchantRequest{}

	tests := []struct {
		name          string
		request       *dto.UpdateMerchantRequest
		merchantID    uuid.UUID
		expectedError string
	}{
		{
			name:          "nil request",
			request:       nil,
			merchantID:    uuid.New(),
			expectedError: "request can't be empty",
		},
		{
			name:          "nil merchant id",
			request:       validRequest,
			merchantID:    uuid.Nil,
			expectedError: "merchant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ms.UpdateMerchant(
				tt.request,
				tt.merchantID,
			)

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

func TestUpdateMerchantStatus(t *testing.T) {
	ms := &MerchantService{}

	validStatus := "active"
	invalidStatus := "invalid_status"

	tests := []struct {
		name          string
		status        *string
		merchantID    uuid.UUID
		expectedError string
	}{
		{
			name:          "nil status",
			status:        nil,
			merchantID:    uuid.New(),
			expectedError: "merchant status is required",
		},
		{
			name:          "invalid status",
			status:        &invalidStatus,
			merchantID:    uuid.New(),
			expectedError: "invalid merchant status",
		},
		{
			name:          "nil merchant id",
			status:        &validStatus,
			merchantID:    uuid.Nil,
			expectedError: "merchant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ms.UpdateMerchantStatus(
				tt.status,
				tt.merchantID,
			)

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

func TestUpdateMerchantKycInfo(t *testing.T) {
	ms := &MerchantService{}

	validRequest := &dto.UpdateMerchantKycRequest{}

	tests := []struct {
		name          string
		request       *dto.UpdateMerchantKycRequest
		merchantID    uuid.UUID
		expectedError string
	}{
		{
			name:          "nil merchant id",
			request:       validRequest,
			merchantID:    uuid.Nil,
			expectedError: "merchant_id is required",
		},
		{
			name:          "nil request",
			request:       nil,
			merchantID:    uuid.New(),
			expectedError: "request can't be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ms.UpdateMerchantKycInfo(
				tt.request,
				tt.merchantID,
			)

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

func TestGetMerchantByEmail(t *testing.T) {
	ms := &MerchantService{}

	tests := []struct {
		name          string
		email         string
		expectedError string
	}{
		{
			name:          "empty email",
			email:         "",
			expectedError: "email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ms.GetMerchantByEmail(tt.email)

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
