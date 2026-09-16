package repositories

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCreatePayment(t *testing.T) {
	senderMerchant := testutils.GenerateTestMerchant()
	receiverMerchant := testutils.GenerateTestMerchant()

	if err := db.Create(senderMerchant).Error; err != nil {
		t.Fatalf("failed to create sender merchant: %v", err)
	}

	if err := db.Create(receiverMerchant).Error; err != nil {
		t.Fatalf("failed to create receiver merchant: %v", err)
	}

	status := constants.TransactionStatusCompleted

	request := &dto.CreatePaymentRequest{
		ReceiverMerchantID: receiverMerchant.ID,
		Amount:             decimal.NewFromInt(1000),
		Currency:           "INR",
		Description:        "integration test payment",
	}

	payment, err := testRepo.CreatePayment(
		senderMerchant.ID,
		request,
		&status,
	)

	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}

	if payment == nil {
		t.Fatal("expected payment, got nil")
	}

	if payment.ID == uuid.Nil {
		t.Fatal("expected payment ID to be generated")
	}

	if payment.PaymentReference == nil || *payment.PaymentReference == "" {
		t.Fatal("expected payment reference to be generated")
	}

	if payment.CustomerReference == nil || *payment.CustomerReference == "" {
		t.Fatal("expected customer reference to be generated")
	}

	if payment.SenderMerchantID != senderMerchant.ID {
		t.Fatalf(
			"expected sender merchant ID %v, got %v",
			senderMerchant.ID,
			payment.SenderMerchantID,
		)
	}

	if payment.ReceiverMerchantID != receiverMerchant.ID {
		t.Fatalf(
			"expected receiver merchant ID %v, got %v",
			receiverMerchant.ID,
			payment.ReceiverMerchantID,
		)
	}

	if !payment.Amount.Equal(decimal.NewFromInt(1000)) {
		t.Fatalf("expected amount 1000, got %s", payment.Amount)
	}

	if payment.Currency == nil || *payment.Currency != "INR" {
		t.Fatalf("expected currency INR, got %v", payment.Currency)
	}

	if payment.Status == nil || *payment.Status != status {
		t.Fatalf("expected status %s, got %v", status, payment.Status)
	}

	expectedSettlementStatus := constants.LedgerSettlementPending

	if payment.SettlementStatus == nil ||
		*payment.SettlementStatus != expectedSettlementStatus {
		t.Fatalf(
			"expected settlement status %s, got %v",
			expectedSettlementStatus,
			payment.SettlementStatus,
		)
	}
}

func TestUpdatePaymentStatus(t *testing.T) {

	t.Run("updates payment status", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatal(err)
		}

		initialStatus := constants.TransactionStatusPending

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "update status test",
		}

		payment, err := testRepo.CreatePayment(
			senderMerchant.ID,
			request,
			&initialStatus,
		)
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		newStatus := constants.TransactionStatusCompleted

		updatedPayment, err := testRepo.UpdatePaymentStatus(
			*payment.PaymentReference,
			&newStatus,
		)
		if err != nil {
			t.Fatalf("failed to update payment status: %v", err)
		}

		if updatedPayment == nil {
			t.Fatal("expected updated payment, got nil")
		}

		if updatedPayment.Status == nil {
			t.Fatal("expected payment status, got nil")
		}

		if *updatedPayment.Status != newStatus {
			t.Errorf(
				"expected status %q, got %q",
				newStatus,
				*updatedPayment.Status,
			)
		}
	})

	t.Run("returns payment when status is already same", func(t *testing.T) {
		status := constants.TransactionStatusCompleted

		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatal(err)
		}

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "same status test",
		}

		payment, err := testRepo.CreatePayment(
			senderMerchant.ID,
			request,
			&status,
		)
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		updatedPayment, err := testRepo.UpdatePaymentStatus(
			*payment.PaymentReference,
			&status,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if updatedPayment == nil {
			t.Fatal("expected payment, got nil")
		}

		if updatedPayment.ID != payment.ID {
			t.Errorf(
				"expected payment ID %s, got %s",
				payment.ID,
				updatedPayment.ID,
			)
		}

		if updatedPayment.Status == nil {
			t.Fatal("expected payment status, got nil")
		}

		if *updatedPayment.Status != status {
			t.Errorf(
				"expected status %q, got %q",
				status,
				*updatedPayment.Status,
			)
		}
	})

	t.Run("returns error when payment does not exist", func(t *testing.T) {
		status := constants.TransactionStatusCompleted
		paymentReference := "nonexistent-payment-reference"

		updatedPayment, err := testRepo.UpdatePaymentStatus(
			paymentReference,
			&status,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if updatedPayment != nil {
			t.Errorf("expected nil payment, got %v", updatedPayment)
		}
	})
}

func TestUpdateSettlementStatusByID(t *testing.T) {

	t.Run("updates settlement status", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatal(err)
		}

		status := constants.TransactionStatusCompleted

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "settlement status test",
		}

		payment, err := testRepo.CreatePayment(
			senderMerchant.ID,
			request,
			&status,
		)
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		newSettlementStatus := constants.LedgerSettlementSettled

		updatedPayment, err := testRepo.UpdateSettlementStatusByID(
			payment.ID,
			&newSettlementStatus,
		)
		if err != nil {
			t.Fatalf("failed to update settlement status: %v", err)
		}

		if updatedPayment == nil {
			t.Fatal("expected updated payment, got nil")
		}

		if updatedPayment.ID != payment.ID {
			t.Errorf(
				"expected payment ID %s, got %s",
				payment.ID,
				updatedPayment.ID,
			)
		}

		if updatedPayment.SettlementStatus == nil {
			t.Fatal("expected settlement status, got nil")
		}

		if *updatedPayment.SettlementStatus != newSettlementStatus {
			t.Errorf(
				"expected settlement status %q, got %q",
				newSettlementStatus,
				*updatedPayment.SettlementStatus,
			)
		}
	})

	t.Run("returns error when payment does not exist", func(t *testing.T) {
		paymentID := uuid.New()
		settlementStatus := constants.LedgerSettlementSettled

		updatedPayment, err := testRepo.UpdateSettlementStatusByID(
			paymentID,
			&settlementStatus,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if updatedPayment != nil {
			t.Errorf("expected nil payment, got %v", updatedPayment)
		}
	})
}

func TestGetPaymentByID(t *testing.T) {

	t.Run("returns payment by ID", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatal(err)
		}

		status := constants.TransactionStatusCompleted

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "get payment test",
		}

		createdPayment, err := testRepo.CreatePayment(
			senderMerchant.ID,
			request,
			&status,
		)
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		payment, err := testRepo.GetPaymentByID(createdPayment.ID)
		if err != nil {
			t.Fatalf("failed to get payment: %v", err)
		}

		if payment == nil {
			t.Fatal("expected payment, got nil")
		}

		if payment.ID != createdPayment.ID {
			t.Errorf(
				"expected payment ID %s, got %s",
				createdPayment.ID,
				payment.ID,
			)
		}

		if payment.PaymentReference == nil {
			t.Fatal("expected payment reference, got nil")
		}

		if createdPayment.PaymentReference == nil {
			t.Fatal("created payment reference is nil")
		}

		if *payment.PaymentReference != *createdPayment.PaymentReference {
			t.Errorf(
				"expected payment reference %s, got %s",
				*createdPayment.PaymentReference,
				*payment.PaymentReference,
			)
		}
	})

	t.Run("returns nil when payment does not exist", func(t *testing.T) {
		paymentID := uuid.New()

		payment, err := testRepo.GetPaymentByID(paymentID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if payment != nil {
			t.Errorf("expected nil payment, got %v", payment)
		}
	})
}

func TestGetPaymentsBySettlementStatus(t *testing.T) {

	t.Run("returns completed payments with matching settlement status", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatal(err)
		}

		status := constants.TransactionStatusCompleted

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "settlement status filter test",
		}

		payment, err := testRepo.CreatePayment(
			senderMerchant.ID,
			request,
			&status,
		)
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		settlementStatus := constants.LedgerSettlementSettled

		_, err = testRepo.UpdateSettlementStatusByID(
			payment.ID,
			&settlementStatus,
		)
		if err != nil {
			t.Fatalf("failed to update settlement status: %v", err)
		}

		payments, err := testRepo.GetPaymentsBySettlementStatus(
			settlementStatus,
		)
		if err != nil {
			t.Fatalf("failed to get payments: %v", err)
		}

		found := false

		for _, result := range payments {
			if result.ID == payment.ID {
				found = true
				break
			}
		}

		if !found {
			t.Errorf(
				"expected payment %s to be returned",
				payment.ID,
			)
		}
	})

	t.Run("does not return payment with different settlement status", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatal(err)
		}

		status := constants.TransactionStatusCompleted

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "different settlement status test",
		}

		payment, err := testRepo.CreatePayment(
			senderMerchant.ID,
			request,
			&status,
		)
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		differentSettlementStatus := constants.LedgerSettlementSettled

		_, err = testRepo.UpdateSettlementStatusByID(
			payment.ID,
			&differentSettlementStatus,
		)
		if err != nil {
			t.Fatalf("failed to update settlement status: %v", err)
		}

		requestedSettlementStatus := constants.LedgerSettlementPending

		payments, err := testRepo.GetPaymentsBySettlementStatus(
			requestedSettlementStatus,
		)
		if err != nil {
			t.Fatalf("failed to get payments: %v", err)
		}

		for _, result := range payments {
			if result.ID == payment.ID {
				t.Errorf(
					"payment %s should not have been returned",
					payment.ID,
				)
			}
		}
	})

	t.Run("does not return non-completed payment", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatal(err)
		}

		status := constants.TransactionStatusPending

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "non completed payment test",
		}

		payment, err := testRepo.CreatePayment(
			senderMerchant.ID,
			request,
			&status,
		)
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		settlementStatus := constants.LedgerSettlementPending

		payments, err := testRepo.GetPaymentsBySettlementStatus(
			settlementStatus,
		)
		if err != nil {
			t.Fatalf("failed to get payments: %v", err)
		}

		for _, result := range payments {
			if result.ID == payment.ID {
				t.Errorf(
					"non-completed payment %s should not have been returned",
					payment.ID,
				)
			}
		}
	})

	t.Run("returns empty result when no payments match", func(t *testing.T) {
		settlementStatus := "non_existent_settlement_status"

		payments, err := testRepo.GetPaymentsBySettlementStatus(
			settlementStatus,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(payments) != 0 {
			t.Errorf(
				"expected 0 payments, got %d",
				len(payments),
			)
		}
	})
}

func TestGetPaymentsByMerchantID(t *testing.T) {

	t.Run("returns payments for merchant", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatal(err)
		}

		status := constants.TransactionStatusCompleted

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "merchant payment test",
		}

		payment, err := testRepo.CreatePayment(
			senderMerchant.ID,
			request,
			&status,
		)
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		payments, err := testRepo.GetPaymentsByMerchantID(senderMerchant.ID)
		if err != nil {
			t.Fatalf("failed to get payments: %v", err)
		}

		found := false

		for _, result := range payments {
			if result.ID == payment.ID {
				found = true
				break
			}
		}

		if !found {
			t.Errorf(
				"expected payment %s to be returned for merchant %s",
				payment.ID,
				senderMerchant.ID,
			)
		}
	})

	t.Run("does not return payments belonging to another merchant", func(t *testing.T) {
		senderMerchant := testutils.GenerateTestMerchant()
		receiverMerchant := testutils.GenerateTestMerchant()
		otherMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatal(err)
		}

		if err := db.Create(otherMerchant).Error; err != nil {
			t.Fatal(err)
		}

		status := constants.TransactionStatusCompleted

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverMerchant.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "merchant filter test",
		}

		payment, err := testRepo.CreatePayment(
			senderMerchant.ID,
			request,
			&status,
		)
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		payments, err := testRepo.GetPaymentsByMerchantID(otherMerchant.ID)
		if err != nil {
			t.Fatalf("failed to get payments: %v", err)
		}

		for _, result := range payments {
			if result.ID == payment.ID {
				t.Errorf(
					"payment %s should not be returned for merchant %s",
					payment.ID,
					otherMerchant.ID,
				)
			}
		}
	})

	t.Run("returns empty result when merchant has no payments", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatal(err)
		}

		payments, err := testRepo.GetPaymentsByMerchantID(merchant.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(payments) != 0 {
			t.Errorf(
				"expected 0 payments, got %d",
				len(payments),
			)
		}
	})
}
