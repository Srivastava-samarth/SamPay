package services

import (
	"testing"

	"github.com/google/uuid"
)

func TestGetBankAccountsByMerchantID(t *testing.T) {

	_, err := testServices.GetBankAccountsByMerchantID(uuid.Nil)

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

	_, err := testServices.GetBankAccount(uuid.Nil)

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
