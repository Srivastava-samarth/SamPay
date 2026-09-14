package services

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

func TestCreateLinkedBankAccount(t *testing.T) {
	lbs := &LinkedBankAccountService{}

	validRequest := &dto.CreateLinkedBankAccountRequest{
		MerchantID:    uuid.New(),
		BankAccountID: uuid.New(),
		Type:          "primary",
		Status:        "active",
	}

	tests := []struct {
		name    string
		request *dto.CreateLinkedBankAccountRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "request is required",
		},
		{
			name: "nil merchant id",
			request: func() *dto.CreateLinkedBankAccountRequest {
				req := *validRequest
				req.MerchantID = uuid.Nil
				return &req
			}(),
			wantErr: "merchant_id is required",
		},
		{
			name: "nil bank account id",
			request: func() *dto.CreateLinkedBankAccountRequest {
				req := *validRequest
				req.BankAccountID = uuid.Nil
				return &req
			}(),
			wantErr: "bank_account_id is required",
		},
		{
			name: "empty type",
			request: func() *dto.CreateLinkedBankAccountRequest {
				req := *validRequest
				req.Type = ""
				return &req
			}(),
			wantErr: "type is required",
		},
		{
			name: "empty status",
			request: func() *dto.CreateLinkedBankAccountRequest {
				req := *validRequest
				req.Status = ""
				return &req
			}(),
			wantErr: "status is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := lbs.CreateLinkedBankAccount(tt.request)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.wantErr {
				t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}
