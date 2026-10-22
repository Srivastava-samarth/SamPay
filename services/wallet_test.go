package services

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateWalletForMerchant(t *testing.T) {

	_, err := testServices.CreateWalletForMerchant(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "merchant id is required" {
		t.Fatalf(
			"expected error %q, got %q",
			"merchant id is required",
			err.Error(),
		)
	}
}

func TestGetWalletByMerchantID(t *testing.T) {

	_, err := testServices.GetWalletByMerchantID(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "merchant id is required" {
		t.Fatalf(
			"expected error %q, got %q",
			"merchant id is required",
			err.Error(),
		)
	}
}

func TestUpdateWalletStatus(t *testing.T) {

	_, err := testServices.UpdateWalletStatus(uuid.Nil, "")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "merchant id is required" {
		t.Fatalf(
			"expected error %q, got %q",
			"merchant id is required",
			err.Error(),
		)
	}
}