package services

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

func TestPerformIndividualComplianceCheck(t *testing.T) {
	cs := &ComplianceService{}

	merchantID := uuid.New()
	tests := []struct {
		name       string
		request    *dto.IndividualComplianceCheckRequest
		wantErr    string
		wantStatus string
		wantReason string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "individual merchant request is required",
		},
		{
			name: "invalid request",
			request: &dto.IndividualComplianceCheckRequest{
				FirstName: "John",
			},
			wantErr: "first name, last name, country, date of birth, and email are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := cs.PerformIndividualComplianceCheck(
				tt.request,
				merchantID,
			)

			if tt.wantErr != "" {
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

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if response == nil {
				t.Fatal("expected response, got nil")
			}

			if response.ComplianceStatus != tt.wantStatus {
				t.Errorf(
					"expected status %q, got %q",
					tt.wantStatus,
					response.ComplianceStatus,
				)
			}

			if response.ComplianceReason != tt.wantReason {
				t.Errorf(
					"expected reason %q, got %q",
					tt.wantReason,
					response.ComplianceReason,
				)
			}
		})
	}
}

func TestPerformCorporateComplianceCheck(t *testing.T) {
	cs := &ComplianceService{}

	merchantID := uuid.New()

	tests := []struct {
		name    string
		request *dto.CompanyComplianceCheckRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "corporate merchant request is required",
		},
		{
			name: "missing legal name",
			request: &dto.CompanyComplianceCheckRequest{
				RegistrationNumber:   "REG123",
				TaxID:                "TAX123",
				Email:                "company@example.com",
				IncorporationCountry: "India",
			},
			wantErr: "legal name, registration number, tax ID, email, and incorporation country are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cs.PerformCorporateComplianceCheck(
				tt.request,
				merchantID,
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

func TestValidateIndividualMerchantRequest(t *testing.T) {
	cs := &ComplianceService{}

	tests := []struct {
		name    string
		request *dto.IndividualComplianceCheckRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "individual merchant request is required",
		},
		{
			name: "missing first name",
			request: &dto.IndividualComplianceCheckRequest{
				LastName:    "Doe",
				Country:     "India",
				DateOfBirth: "1995-01-01",
				Email:       "john@example.com",
			},
			wantErr: "first name, last name, country, date of birth, and email are required",
		},
		{
			name: "missing last name",
			request: &dto.IndividualComplianceCheckRequest{
				FirstName:   "John",
				Country:     "India",
				DateOfBirth: "1995-01-01",
				Email:       "john@example.com",
			},
			wantErr: "first name, last name, country, date of birth, and email are required",
		},
		{
			name: "missing country",
			request: &dto.IndividualComplianceCheckRequest{
				FirstName:   "John",
				LastName:    "Doe",
				DateOfBirth: "1995-01-01",
				Email:       "john@example.com",
			},
			wantErr: "first name, last name, country, date of birth, and email are required",
		},
		{
			name: "missing date of birth",
			request: &dto.IndividualComplianceCheckRequest{
				FirstName: "John",
				LastName:  "Doe",
				Country:   "India",
				Email:     "john@example.com",
			},
			wantErr: "first name, last name, country, date of birth, and email are required",
		},
		{
			name: "missing email",
			request: &dto.IndividualComplianceCheckRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Country:     "India",
				DateOfBirth: "1995-01-01",
			},
			wantErr: "first name, last name, country, date of birth, and email are required",
		},
		{
			name: "invalid country",
			request: &dto.IndividualComplianceCheckRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Country:     "DefinitelyNotACountry",
				DateOfBirth: "1995-01-01",
				Email:       "john@example.com",
			},
			wantErr: "invalid country",
		},
		{
			name: "invalid date format",
			request: &dto.IndividualComplianceCheckRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Country:     "India",
				DateOfBirth: "01-01-1995",
				Email:       "john@example.com",
			},
			wantErr: "date of birth must be in YYYY-MM-DD format",
		},
		{
			name: "future date of birth",
			request: &dto.IndividualComplianceCheckRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Country:     "India",
				DateOfBirth: "2099-01-01",
				Email:       "john@example.com",
			},
			wantErr: "date of birth cannot be in the future",
		},
		{
			name: "under 18",
			request: &dto.IndividualComplianceCheckRequest{
				FirstName:   "John",
				LastName:    "Doe",
				Country:     "India",
				DateOfBirth: "2015-01-01",
				Email:       "john@example.com",
			},
			wantErr: "merchant must be at least 18 years old",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cs.ValidateIndividualMerchantRequest(tt.request)

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

func TestValidateCorporateMerchantRequest(t *testing.T) {
	cs := &ComplianceService{}

	tests := []struct {
		name    string
		request *dto.CompanyComplianceCheckRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "corporate merchant request is required",
		},
		{
			name: "missing required fields",
			request: &dto.CompanyComplianceCheckRequest{
				LegalName: "Test Company",
			},
			wantErr: "legal name, registration number, tax ID, email, and incorporation country are required",
		},
		{
			name: "invalid country",
			request: &dto.CompanyComplianceCheckRequest{
				LegalName:            "Test Company",
				RegistrationNumber:   "REG123",
				TaxID:                "TAX123",
				Email:                "company@example.com",
				IncorporationCountry: "DefinitelyNotACountry",
			},
			wantErr: "invalid country",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cs.ValidateCorporateMerchantRequest(tt.request)

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

func TestValidateDateOfBirth(t *testing.T) {
	tests := []struct {
		name    string
		dob     string
		wantErr string
	}{
		{
			name:    "invalid date format",
			dob:     "01-01-2000",
			wantErr: "date of birth must be in YYYY-MM-DD format",
		},
		{
			name:    "future date",
			dob:     "2099-01-01",
			wantErr: "date of birth cannot be in the future",
		},
		{
			name:    "under 18",
			dob:     "2015-01-01",
			wantErr: "merchant must be at least 18 years old",
		},
		{
			name:    "valid adult",
			dob:     "1995-01-01",
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateOfBirth(tt.dob)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}

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
