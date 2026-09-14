package services

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

func TestCreateBankAccount(t *testing.T) {
	bs := &BankService{}

	tests := []struct {
		name    string
		request *dto.CreateBankAccountRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "request is required",
		},
		{
			name: "empty account name",
			request: &dto.CreateBankAccountRequest{
				AccountName: "",
				MerchantID:  uuid.New(),
			},
			wantErr: "account_name is required",
		},
		{
			name: "nil merchant id",
			request: &dto.CreateBankAccountRequest{
				AccountName: "Primary Account",
				MerchantID:  uuid.Nil,
			},
			wantErr: "merchant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bs.CreateBankAccount(tt.request)

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

func TestUpdateBankAccountAndlink(t *testing.T) {
	bs := &BankService{}

	validRequest := &dto.UpdateBankAccountRequest{
		ID:          uuid.New(),
		AccountName: "Updated Account",
		AccountType: "primary",
		Status:      "active",
	}

	tests := []struct {
		name    string
		request *dto.UpdateBankAccountRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "request is required",
		},
		{
			name: "nil id",
			request: func() *dto.UpdateBankAccountRequest {
				req := *validRequest
				req.ID = uuid.Nil
				return &req
			}(),
			wantErr: "id is required",
		},
		{
			name: "empty account name",
			request: func() *dto.UpdateBankAccountRequest {
				req := *validRequest
				req.AccountName = ""
				return &req
			}(),
			wantErr: "account name is required",
		},
		{
			name: "empty account type",
			request: func() *dto.UpdateBankAccountRequest {
				req := *validRequest
				req.AccountType = ""
				return &req
			}(),
			wantErr: "account_type is required",
		},
		{
			name: "empty status",
			request: func() *dto.UpdateBankAccountRequest {
				req := *validRequest
				req.Status = ""
				return &req
			}(),
			wantErr: "status is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := bs.UpdateBankAccountAndlink(tt.request)

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

func TestGetBankAccountsByMerchantID(t *testing.T) {
	bs := &BankService{}

	_, err := bs.GetBankAccountsByMerchantID(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "merchant_id is required" {
		t.Errorf(
			"expected error %q, got %q",
			"merchant_id is required",
			err.Error(),
		)
	}
}

func TestGetBankAccount(t *testing.T) {
	bs := &BankService{}

	_, err := bs.GetBankAccount(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "bank_account_id is empty" {
		t.Errorf(
			"expected error %q, got %q",
			"bank_account_id is empty",
			err.Error(),
		)
	}
}
