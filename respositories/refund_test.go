package repositories

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCreateRefund(t *testing.T) {

	merchant1 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant1).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	merchant2 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant2).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	currency := "USD"
	paymentStatus := constants.TransactionStatusCompleted
	settlementStatus := constants.LedgerSettlementPending
	paymentReference := utils.GeneratePaymentReference()
	description := "Test payment"
	customerReference := utils.GenerateCustomerReference()

	payment := &models.Payment{
		ID:                 uuid.New(),
		SenderMerchantID:   merchant1.ID,
		ReceiverMerchantID: merchant2.ID,
		PaymentReference:   paymentReference,
		Amount:             decimal.NewFromInt(1000),
		Currency:           &currency,
		Status:             &paymentStatus,
		SettlementStatus:   &settlementStatus,
		Description:        &description,
		CustomerReference:  customerReference,
	}

	if err := db.Create(payment).Error; err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	externalReference := "EXT-REF-001"
	reason := "Customer requested refund"

	request := &dto.CreateRefundRequest{
		PaymentID:         payment.ID,
		MerchantID:        merchant1.ID,
		ExternalReference: &externalReference,
		Amount:            decimal.NewFromInt(250),
		Currency:          currency,
		Reason:            &reason,
	}

	refund, err := testRepo.CreateRefund(request)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if refund == nil {
		t.Fatal("expected refund, got nil")
	}

	if refund.ID == uuid.Nil {
		t.Error("expected refund ID to be generated")
	}

	if refund.PaymentID != request.PaymentID {
		t.Errorf("expected payment ID %v, got %v", request.PaymentID, refund.PaymentID)
	}

	if refund.MerchantID != request.MerchantID {
		t.Errorf("expected merchant ID %v, got %v", request.MerchantID, refund.MerchantID)
	}

	if refund.ExternalReference != request.ExternalReference {
		t.Errorf("expected external reference %v, got %v", request.ExternalReference, refund.ExternalReference)
	}

	if !refund.Amount.Equal(request.Amount) {
		t.Errorf("expected amount %v, got %v", request.Amount, refund.Amount)
	}

	if refund.Currency != request.Currency {
		t.Errorf("expected currency %v, got %v", request.Currency, refund.Currency)
	}

	if refund.Reason != request.Reason {
		t.Errorf("expected reason %v, got %v", request.Reason, refund.Reason)
	}

	if refund.Status == "" || refund.Status != constants.TransactionStatusProcessing {
		t.Errorf(
			"expected status %v, got %v",
			constants.TransactionStatusProcessing,
			refund.Status,
		)
	}

	var storedRefund *models.Refund
	if err := db.Where("id = ?", refund.ID).First(&storedRefund).Error; err != nil {
		t.Fatalf("failed to fetch created refund: %v", err)
	}

	if storedRefund.ID != refund.ID {
		t.Errorf("expected stored refund ID %v, got %v", refund.ID, storedRefund.ID)
	}
}

func TestGetRefundsByPaymentID(t *testing.T) {
	merchant1 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant1).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	merchant2 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant2).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	currency := "USD"
	paymentStatus := constants.TransactionStatusCompleted
	settlementStatus := constants.LedgerSettlementPending

	payment1 := &models.Payment{
		ID:                 uuid.New(),
		SenderMerchantID:   merchant1.ID,
		ReceiverMerchantID: merchant2.ID,
		PaymentReference:   utils.GeneratePaymentReference(),
		Amount:             decimal.NewFromInt(1000),
		Currency:           &currency,
		Status:             &paymentStatus,
		SettlementStatus:   &settlementStatus,
	}

	if err := db.Create(payment1).Error; err != nil {
		t.Fatalf("failed to create payment1: %v", err)
	}

	payment2 := &models.Payment{
		ID:                 uuid.New(),
		SenderMerchantID:   merchant1.ID,
		ReceiverMerchantID: merchant2.ID,
		PaymentReference:   utils.GeneratePaymentReference(),
		Amount:             decimal.NewFromInt(2000),
		Currency:           &currency,
		Status:             &paymentStatus,
		SettlementStatus:   &settlementStatus,
	}

	if err := db.Create(payment2).Error; err != nil {
		t.Fatalf("failed to create payment2: %v", err)
	}

	externalReference1 := "EXT-REF-001"
	externalReference2 := "EXT-REF-002"
	reason1 := "First refund"
	reason2 := "Second refund"

	refund1Request := &dto.CreateRefundRequest{
		PaymentID:         payment1.ID,
		MerchantID:        merchant1.ID,
		ExternalReference: &externalReference1,
		Amount:            decimal.NewFromInt(100),
		Currency:          currency,
		Reason:            &reason1,
	}

	refund1, err := testRepo.CreateRefund(refund1Request)
	if err != nil {
		t.Fatalf("failed to create refund1: %v", err)
	}

	refund2Request := &dto.CreateRefundRequest{
		PaymentID:         payment1.ID,
		MerchantID:        merchant1.ID,
		ExternalReference: &externalReference2,
		Amount:            decimal.NewFromInt(200),
		Currency:          currency,
		Reason:            &reason2,
	}

	refund2, err := testRepo.CreateRefund(refund2Request)
	if err != nil {
		t.Fatalf("failed to create refund2: %v", err)
	}

	// Create a refund for another payment.
	externalReference3 := "EXT-REF-003"

	refund3Request := &dto.CreateRefundRequest{
		PaymentID:         payment2.ID,
		MerchantID:        merchant1.ID,
		ExternalReference: &externalReference3,
		Amount:            decimal.NewFromInt(300),
		Currency:          currency,
	}

	_, err = testRepo.CreateRefund(refund3Request)
	if err != nil {
		t.Fatalf("failed to create refund3: %v", err)
	}

	refunds, err := testRepo.GetRefundsByPaymentID(payment1.ID)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(refunds) != 2 {
		t.Fatalf("expected 2 refunds, got %d", len(refunds))
	}

	if refunds[0].ID != refund2.ID {
		t.Errorf("expected newest refund first, got %v", refunds[0].ID)
	}

	if refunds[1].ID != refund1.ID {
		t.Errorf("expected oldest refund second, got %v", refunds[1].ID)
	}

	for _, refund := range refunds {
		if refund.PaymentID != payment1.ID {
			t.Errorf(
				"expected payment ID %v, got %v",
				payment1.ID,
				refund.PaymentID,
			)
		}
	}

	for _, refund := range refunds {
		if refund.ID == refund3Request.PaymentID {
			t.Error("refund belonging to another payment was returned")
		}
	}
}

func TestGetRefundByReference(t *testing.T) {

	merchant1 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant1).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	merchant2 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant2).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	currency := "USD"
	paymentStatus := constants.TransactionStatusCompleted
	settlementStatus := constants.LedgerSettlementPending

	payment := &models.Payment{
		ID:                 uuid.New(),
		SenderMerchantID:   merchant1.ID,
		ReceiverMerchantID: merchant2.ID,
		PaymentReference:   utils.GeneratePaymentReference(),
		Amount:             decimal.NewFromInt(1000),
		Currency:           &currency,
		Status:             &paymentStatus,
		SettlementStatus:   &settlementStatus,
	}

	if err := db.Create(payment).Error; err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	externalReference := "EXT-REF-001"
	reason := "Customer requested refund"

	request := &dto.CreateRefundRequest{
		PaymentID:         payment.ID,
		MerchantID:        merchant1.ID,
		ExternalReference: &externalReference,
		Amount:            decimal.NewFromInt(250),
		Currency:          currency,
		Reason:            &reason,
	}

	createdRefund, err := testRepo.CreateRefund(request)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	t.Run("existing refund", func(t *testing.T) {
		refund, err := testRepo.GetRefundByReference(createdRefund.RefundReference)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if refund == nil {
			t.Fatal("expected refund, got nil")
		}

		if refund.ID != createdRefund.ID {
			t.Errorf(
				"expected refund ID %v, got %v",
				createdRefund.ID,
				refund.ID,
			)
		}

		if refund.RefundReference != createdRefund.RefundReference {
			t.Errorf(
				"expected refund reference %v, got %v",
				createdRefund.RefundReference,
				refund.RefundReference,
			)
		}
	})

	t.Run("non-existing refund", func(t *testing.T) {
		nonExistingReference := "NON-EXISTING-REF"

		_, err := testRepo.GetRefundByReference(nonExistingReference)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})
}

func TestUpdateRefundStatus(t *testing.T) {

	merchant1 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant1).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	merchant2 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant2).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	currency := "USD"
	paymentStatus := constants.TransactionStatusCompleted
	settlementStatus := constants.LedgerSettlementPending

	payment := &models.Payment{
		ID:                 uuid.New(),
		SenderMerchantID:   merchant1.ID,
		ReceiverMerchantID: merchant2.ID,
		PaymentReference:   utils.GeneratePaymentReference(),
		Amount:             decimal.NewFromInt(1000),
		Currency:           &currency,
		Status:             &paymentStatus,
		SettlementStatus:   &settlementStatus,
	}

	if err := db.Create(payment).Error; err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	externalReference := "EXT-REF-001"
	reason := "Customer requested refund"

	request := &dto.CreateRefundRequest{
		PaymentID:         payment.ID,
		MerchantID:        merchant1.ID,
		ExternalReference: &externalReference,
		Amount:            decimal.NewFromInt(250),
		Currency:          currency,
		Reason:            &reason,
	}

	refund, err := testRepo.CreateRefund(request)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	t.Run("update refund status", func(t *testing.T) {
		newStatus := constants.TransactionStatusCompleted

		updatedRefund, err := testRepo.UpdateRefundStatus(
			refund.RefundReference,
			newStatus,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if updatedRefund == nil {
			t.Fatal("expected refund, got nil")
		}

		if updatedRefund.ID != refund.ID {
			t.Errorf(
				"expected refund ID %v, got %v",
				refund.ID,
				updatedRefund.ID,
			)
		}

		if updatedRefund.Status != newStatus {
			t.Errorf(
				"expected status %v, got %v",
				newStatus,
				updatedRefund.Status,
			)
		}
	})

	t.Run("same status", func(t *testing.T) {
		currentStatus := constants.TransactionStatusCompleted

		updatedRefund, err := testRepo.UpdateRefundStatus(
			refund.RefundReference,
			currentStatus,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if updatedRefund == nil {
			t.Fatal("expected refund, got nil")
		}

		if updatedRefund.ID != refund.ID {
			t.Errorf(
				"expected refund ID %v, got %v",
				refund.ID,
				updatedRefund.ID,
			)
		}

		if updatedRefund.Status != currentStatus {
			t.Errorf(
				"expected status %v, got %v",
				currentStatus,
				updatedRefund.Status,
			)
		}
	})
}

func TestGetRefundById(t *testing.T) {

	merchant1 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant1).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	merchant2 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant2).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	currency := "USD"
	paymentStatus := constants.TransactionStatusCompleted
	settlementStatus := constants.LedgerSettlementPending

	payment := &models.Payment{
		ID:                 uuid.New(),
		SenderMerchantID:   merchant1.ID,
		ReceiverMerchantID: merchant2.ID,
		PaymentReference:   utils.GeneratePaymentReference(),
		Amount:             decimal.NewFromInt(1000),
		Currency:           &currency,
		Status:             &paymentStatus,
		SettlementStatus:   &settlementStatus,
	}

	if err := db.Create(payment).Error; err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	externalReference := "EXT-REF-001"
	reason := "Customer requested refund"

	request := &dto.CreateRefundRequest{
		PaymentID:         payment.ID,
		MerchantID:        merchant1.ID,
		ExternalReference: &externalReference,
		Amount:            decimal.NewFromInt(250),
		Currency:          currency,
		Reason:            &reason,
	}

	createdRefund, err := testRepo.CreateRefund(request)
	if err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	t.Run("existing refund", func(t *testing.T) {
		refund, err := testRepo.GetRefundById(createdRefund.ID)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if refund == nil {
			t.Fatal("expected refund, got nil")
		}

		if refund.ID != createdRefund.ID {
			t.Errorf(
				"expected refund ID %v, got %v",
				createdRefund.ID,
				refund.ID,
			)
		}

		if refund.PaymentID != payment.ID {
			t.Errorf(
				"expected payment ID %v, got %v",
				payment.ID,
				refund.PaymentID,
			)
		}

		if refund.MerchantID != merchant1.ID {
			t.Errorf(
				"expected merchant ID %v, got %v",
				merchant1.ID,
				refund.MerchantID,
			)
		}
	})

	t.Run("non-existing refund", func(t *testing.T) {
		_, err := testRepo.GetRefundById(uuid.New())

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})
}

func TestGetRefundsByMerchantId(t *testing.T) {

	merchant1 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant1).Error; err != nil {
		t.Fatalf("failed to create merchant1: %v", err)
	}

	merchant2 := testutils.GenerateTestMerchant()
	if err := db.Create(merchant2).Error; err != nil {
		t.Fatalf("failed to create merchant2: %v", err)
	}

	currency := "USD"
	paymentStatus := constants.TransactionStatusCompleted
	settlementStatus := constants.LedgerSettlementPending

	payment1 := &models.Payment{
		ID:                 uuid.New(),
		SenderMerchantID:   merchant1.ID,
		ReceiverMerchantID: merchant2.ID,
		PaymentReference:   utils.GeneratePaymentReference(),
		Amount:             decimal.NewFromInt(1000),
		Currency:           &currency,
		Status:             &paymentStatus,
		SettlementStatus:   &settlementStatus,
	}

	if err := db.Create(payment1).Error; err != nil {
		t.Fatalf("failed to create payment1: %v", err)
	}

	payment2 := &models.Payment{
		ID:                 uuid.New(),
		SenderMerchantID:   merchant2.ID,
		ReceiverMerchantID: merchant1.ID,
		PaymentReference:   utils.GeneratePaymentReference(),
		Amount:             decimal.NewFromInt(2000),
		Currency:           &currency,
		Status:             &paymentStatus,
		SettlementStatus:   &settlementStatus,
	}

	if err := db.Create(payment2).Error; err != nil {
		t.Fatalf("failed to create payment2: %v", err)
	}

	externalReference1 := "EXT-REF-001"
	externalReference2 := "EXT-REF-002"
	externalReference3 := "EXT-REF-003"

	refund1Request := &dto.CreateRefundRequest{
		PaymentID:         payment1.ID,
		MerchantID:        merchant1.ID,
		ExternalReference: &externalReference1,
		Amount:            decimal.NewFromInt(100),
		Currency:          currency,
	}

	refund1, err := testRepo.CreateRefund(refund1Request)
	if err != nil {
		t.Fatalf("failed to create refund1: %v", err)
	}

	refund2Request := &dto.CreateRefundRequest{
		PaymentID:         payment1.ID,
		MerchantID:        merchant1.ID,
		ExternalReference: &externalReference2,
		Amount:            decimal.NewFromInt(200),
		Currency:          currency,
	}

	refund2, err := testRepo.CreateRefund(refund2Request)
	if err != nil {
		t.Fatalf("failed to create refund2: %v", err)
	}

	refund3Request := &dto.CreateRefundRequest{
		PaymentID:         payment2.ID,
		MerchantID:        merchant2.ID,
		ExternalReference: &externalReference3,
		Amount:            decimal.NewFromInt(300),
		Currency:          currency,
	}

	_, err = testRepo.CreateRefund(refund3Request)
	if err != nil {
		t.Fatalf("failed to create refund3: %v", err)
	}

	refunds, err := testRepo.GetRefundsByMerchantId(merchant1.ID)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(refunds) != 2 {
		t.Fatalf("expected 2 refunds, got %d", len(refunds))
	}

	if refunds[0].ID != refund2.ID {
		t.Errorf(
			"expected newest refund first, got %v",
			refunds[0].ID,
		)
	}

	if refunds[1].ID != refund1.ID {
		t.Errorf(
			"expected oldest refund second, got %v",
			refunds[1].ID,
		)
	}

	for _, refund := range refunds {
		if refund.MerchantID != merchant1.ID {
			t.Errorf(
				"expected merchant ID %v, got %v",
				merchant1.ID,
				refund.MerchantID,
			)
		}
	}
}
