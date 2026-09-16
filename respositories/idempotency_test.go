package repositories

import (
	"testing"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/google/uuid"
)

func TestGetIdempotencyByKey(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()
	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	idempotencyKey := "test-idempotency-key"

	var responseBody []byte
	//nolint:staticcheck // ResponseBody requires []byte.
	for _, b := range []byte(`{
    "data": {
        "id": "7f0994a8-14d2-4b65-9b8a-b6c3a1063a60",
        "amount": "10",
        "status": "failed"
    },
    "success": true
}`) {
		responseBody = append(responseBody, b)
	}
	idempotency := &models.IdempotencyKey{
		ID:             uuid.New(),
		MerchantID:     merchant.ID,
		IdempotencyKey: idempotencyKey,
		ResponseBody:   responseBody,
	}

	if err := db.Create(idempotency).Error; err != nil {
		t.Fatalf("failed to create idempotency key: %v", err)
	}

	t.Run("existing idempotency key", func(t *testing.T) {
		result, err := testRepo.GetIdempotencyByKey(
			merchant.ID,
			idempotencyKey,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("expected idempotency key, got nil")
		}

		if result.ID != idempotency.ID {
			t.Errorf(
				"expected ID %v, got %v",
				idempotency.ID,
				result.ID,
			)
		}

		if result.MerchantID != merchant.ID {
			t.Errorf(
				"expected merchant ID %v, got %v",
				merchant.ID,
				result.MerchantID,
			)
		}

		if result.IdempotencyKey == "" ||
			result.IdempotencyKey != idempotencyKey {
			t.Errorf(
				"expected idempotency key %v, got %v",
				idempotencyKey,
				result.IdempotencyKey,
			)
		}
	})

	t.Run("non-existing idempotency key", func(t *testing.T) {
		nonExistingKey := "non-existing-idempotency-key"

		result, err := testRepo.GetIdempotencyByKey(
			merchant.ID,
			nonExistingKey,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})
}

func TestCreateIdempotencyKey(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()
	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	idempotencyKey := "test-create-idempotency-key"

	var responseBody []byte
	//nolint:staticcheck // ResponseBody requires []byte.
	for _, b := range []byte(`{
    "data": {
        "id": "7f0994a8-14d2-4b65-9b8a-b6c3a1063a60",
        "amount": "10",
        "status": "failed"
    },
    "success": true
}`) {
		responseBody = append(responseBody, b)
	}
	request := &models.IdempotencyKey{
		ID:             uuid.New(),
		MerchantID:     merchant.ID,
		IdempotencyKey: idempotencyKey,
		ResponseBody:   responseBody,
	}

	result, err := testRepo.CreateIdempotencyKey(request)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected idempotency key, got nil")
	}

	if result.ID != request.ID {
		t.Errorf("expected ID %v, got %v", request.ID, result.ID)
	}

	if result.MerchantID != merchant.ID {
		t.Errorf("expected merchant ID %v, got %v", merchant.ID, result.MerchantID)
	}

	if result.IdempotencyKey == "" || result.IdempotencyKey != idempotencyKey {
		t.Errorf("expected idempotency key %q, got %v", idempotencyKey, result.IdempotencyKey)
	}

	// Verify it was actually persisted.
	var stored models.IdempotencyKey
	if err := db.Where("id = ?", request.ID).First(&stored).Error; err != nil {
		t.Fatalf("failed to fetch created idempotency key: %v", err)
	}

	if stored.ID != request.ID {
		t.Errorf("expected persisted ID %v, got %v", request.ID, stored.ID)
	}
}
