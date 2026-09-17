package activities

import (
	"context"
	"testing"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestValidateRefundRequest(t *testing.T) {
	t.Run("request is nil", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		err := registry.ValidateRefundRequest(
			context.Background(),
			nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "request is required" {
			t.Errorf(
				"expected error %q, got %q",
				"request is required",
				err.Error(),
			)
		}
	})

	t.Run("payment ID is missing", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant %v", err)
		}
		reason := "customer requested refund"

		request := &dto.CreateRefundRequest{
			MerchantID: merchant.ID,
			Amount:     decimal.NewFromInt(100),
			Reason:     &reason,
			Currency:   "INR",
		}

		err := registry.ValidateRefundRequest(
			context.Background(),
			request,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "payment id not provided" {
			t.Errorf(
				"expected error %q, got %q",
				"payment id not provided",
				err.Error(),
			)
		}
	})

	t.Run("merchant ID is missing", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		reason := "customer requested refund"

		request := &dto.CreateRefundRequest{
			PaymentID: uuid.New(),
			Amount:    decimal.NewFromInt(100),
			Reason:    &reason,
			Currency:  "INR",
		}

		err := registry.ValidateRefundRequest(
			context.Background(),
			request,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "merchant id not provided" {
			t.Errorf(
				"expected error %q, got %q",
				"merchant id not provided",
				err.Error(),
			)
		}
	})

	t.Run("amount must be greater than zero", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		request := &dto.CreateRefundRequest{
			PaymentID:  uuid.New(),
			MerchantID: uuid.New(),
			Amount:     decimal.Zero,
		}

		err := registry.ValidateRefundRequest(
			context.Background(),
			request,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "amount must be greater than 0" {
			t.Errorf(
				"expected error %q, got %q",
				"amount must be greater than 0",
				err.Error(),
			)
		}
	})

	t.Run("reason is required", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		request := &dto.CreateRefundRequest{
			PaymentID:  uuid.New(),
			MerchantID: uuid.New(),
			Amount:     decimal.NewFromInt(100),
		}

		err := registry.ValidateRefundRequest(
			context.Background(),
			request,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "reason is necessary" {
			t.Errorf(
				"expected error %q, got %q",
				"reason is necessary",
				err.Error(),
			)
		}
	})

	t.Run("payment is not found", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		reason := "customer requested refund"

		request := &dto.CreateRefundRequest{
			PaymentID:  uuid.New(),
			MerchantID: uuid.New(),
			Amount:     decimal.NewFromInt(100),
			Reason:     &reason,
			Currency:   "INR",
		}

		err := registry.ValidateRefundRequest(
			context.Background(),
			request,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("payment does not belong to merchant", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating merchant %v", err)
		}

		senderMerchantID := merchant.ID
		requestMerchantID := uuid.New()

		currency := "INR"
		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled
		paymentReference := "refund-validation-payment"

		payment := &models.Payment{
			ID:                 uuid.New(),
			SenderMerchantID:   senderMerchantID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   &paymentReference,
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		reason := "customer requested refund"

		request := &dto.CreateRefundRequest{
			PaymentID:  payment.ID,
			MerchantID: requestMerchantID,
			Amount:     decimal.NewFromInt(100),
			Reason:     &reason,
			Currency:   "INR",
		}

		err := registry.ValidateRefundRequest(
			context.Background(),
			request,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "payment doesn't belong to the merchant" {
			t.Errorf(
				"expected error %q, got %q",
				"payment doesn't belong to the merchant",
				err.Error(),
			)
		}
	})

	t.Run("currency is required", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating merchant %v", err)
		}

		merchantID := merchant.ID
		currency := "INR"
		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled
		paymentReference := "currency-required-payment"

		payment := &models.Payment{
			ID:                 uuid.New(),
			SenderMerchantID:   merchantID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   &paymentReference,
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		reason := "customer requested refund"

		request := &dto.CreateRefundRequest{
			PaymentID:  payment.ID,
			MerchantID: merchantID,
			Amount:     decimal.NewFromInt(100),
			Reason:     &reason,
		}

		err := registry.ValidateRefundRequest(
			context.Background(),
			request,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "currency is required" {
			t.Errorf(
				"expected error %q, got %q",
				"currency is required",
				err.Error(),
			)
		}
	})

	t.Run("currency must match payment currency", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating merchant %v", err)
		}

		merchantID := merchant.ID
		currency := "INR"
		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled
		paymentReference := "currency-mismatch-payment"

		payment := &models.Payment{
			ID:                 uuid.New(),
			SenderMerchantID:   merchantID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   &paymentReference,
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		reason := "customer requested refund"

		request := &dto.CreateRefundRequest{
			PaymentID:  payment.ID,
			MerchantID: merchantID,
			Amount:     decimal.NewFromInt(100),
			Reason:     &reason,
			Currency:   "USD",
		}

		err := registry.ValidateRefundRequest(
			context.Background(),
			request,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "refund currency must match payment currency" {
			t.Errorf(
				"expected error %q, got %q",
				"refund currency must match payment currency",
				err.Error(),
			)
		}
	})

	t.Run("refund amount exceeds refundable amount", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating merchant %v", err)
		}

		merchantID := merchant.ID
		currency := "INR"
		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled
		paymentReference := utils.GeneratePaymentReference()

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   merchantID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   paymentReference,
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		reason := "customer requested refund"

		request := &dto.CreateRefundRequest{
			PaymentID:  payment.ID,
			MerchantID: merchantID,
			Amount:     decimal.NewFromInt(1200),
			Reason:     &reason,
			Currency:   "INR",
		}

		err := registry.ValidateRefundRequest(
			context.Background(),
			request,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "refund amount exceeds refundable amount" {
			t.Errorf(
				"expected error %q, got %q",
				"refund amount exceeds refundable amount",
				err.Error(),
			)
		}

		reason = "customer requested refund"

		request = &dto.CreateRefundRequest{
			PaymentID:  payment.ID,
			MerchantID: merchantID,
			Amount:     decimal.NewFromInt(500),
			Reason:     &reason,
			Currency:   "INR",
		}

		errV := registry.ValidateRefundRequest(
			context.Background(),
			request,
		)

		if errV != nil {
			t.Fatalf("expected no error, got %v", errV)
		}
	})
}

func TestUpdateRefundStatus(t *testing.T) {
	t.Run("refund not found", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		refundReference := utils.GenerateRefundReference()
		status := constants.TransactionStatusRefunded

		refund, err := registry.UpdateRefundStatus(
			context.Background(),
			*refundReference,
			status,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if refund != nil {
			t.Errorf("expected refund to be nil, got %+v", refund)
		}
	})

	t.Run("status is already the same", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		paymentSettlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &paymentSettlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("error creating payment: %v", err)
		}

		refundReference := utils.GenerateRefundReference()
		status := constants.TransactionStatusRefunded
		reason := "customer requested refund"

		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      merchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(500),
			Currency:        currency,
			Status:          status,
			Reason:          &reason,
		}

		if err := db.Create(refund).Error; err != nil {
			t.Fatalf("error creating refund: %v", err)
		}

		updatedRefund, err := registry.UpdateRefundStatus(
			context.Background(),
			*refundReference,
			status,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
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

		if updatedRefund.Status == "" {
			t.Fatal("expected refund status, got nil")
		}

		if updatedRefund.Status != status {
			t.Errorf(
				"expected status %q, got %q",
				status,
				updatedRefund.Status,
			)
		}
	})

	t.Run("status is updated successfully", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		paymentSettlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &paymentSettlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("error creating payment: %v", err)
		}

		refundReference := utils.GenerateRefundReference()
		initialStatus := constants.TransactionStatusPending
		newStatus := constants.TransactionStatusRefunded
		reason := "customer requested refund"

		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      merchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(500),
			Currency:        currency,
			Status:          initialStatus,
			Reason:          &reason,
		}

		if err := db.Create(refund).Error; err != nil {
			t.Fatalf("error creating refund: %v", err)
		}

		updatedRefund, err := registry.UpdateRefundStatus(
			context.Background(),
			*refundReference,
			newStatus,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
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

		if updatedRefund.Status == "" {
			t.Fatal("expected refund status, got nil")
		}

		if updatedRefund.Status != newStatus {
			t.Errorf(
				"expected status %q, got %q",
				newStatus,
				updatedRefund.Status,
			)
		}

		var persistedRefund models.Refund
		if err := db.Where("id = ?", refund.ID).First(&persistedRefund).Error; err != nil {
			t.Fatalf("failed to fetch persisted refund: %v", err)
		}

		if persistedRefund.Status == "" {
			t.Fatal("expected persisted refund status, got nil")
		}

		if persistedRefund.Status != newStatus {
			t.Errorf(
				"expected persisted status %q, got %q",
				newStatus,
				persistedRefund.Status,
			)
		}
	})
}

func TestGetRefundByPaymentID(t *testing.T) {
	t.Run("returns refunds for payment", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		paymentSettlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &paymentSettlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("error creating payment: %v", err)
		}

		status := constants.TransactionStatusPending

		refundReference1 := utils.GenerateRefundReference()
		reason1 := "first refund"

		refund1 := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      merchant.ID,
			RefundReference: *refundReference1,
			Amount:          decimal.NewFromInt(200),
			Currency:        currency,
			Status:          status,
			Reason:          &reason1,
			CreatedAt:       time.Now().Add(-2 * time.Hour),
		}

		refundReference2 := utils.GenerateRefundReference()
		reason2 := "second refund"

		refund2 := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      merchant.ID,
			RefundReference: *refundReference2,
			Amount:          decimal.NewFromInt(300),
			Currency:        currency,
			Status:          status,
			Reason:          &reason2,
			CreatedAt:       time.Now(),
		}

		if err := db.Create(refund1).Error; err != nil {
			t.Fatalf("error creating first refund: %v", err)
		}

		if err := db.Create(refund2).Error; err != nil {
			t.Fatalf("error creating second refund: %v", err)
		}

		refunds, err := registry.GetRefundByPaymentID(
			context.Background(),
			payment.ID,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(refunds) != 2 {
			t.Fatalf("expected 2 refunds, got %d", len(refunds))
		}

		if refunds[0].ID != refund2.ID {
			t.Errorf(
				"expected newest refund %v at index 0, got %v",
				refund2.ID,
				refunds[0].ID,
			)
		}

		if refunds[1].ID != refund1.ID {
			t.Errorf(
				"expected older refund %v at index 1, got %v",
				refund1.ID,
				refunds[1].ID,
			)
		}
	})

	t.Run("does not return refunds belonging to another payment", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		paymentSettlementStatus := constants.LedgerSettlementSettled

		payment1 := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &paymentSettlementStatus,
		}

		payment2 := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(2000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &paymentSettlementStatus,
		}

		if err := db.Create(payment1).Error; err != nil {
			t.Fatalf("error creating payment1: %v", err)
		}

		if err := db.Create(payment2).Error; err != nil {
			t.Fatalf("error creating payment2: %v", err)
		}

		status := constants.TransactionStatusPending
		refundReference := utils.GenerateRefundReference()

		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment2.ID,
			MerchantID:      merchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(500),
			Currency:        currency,
			Status:          status,
		}

		if err := db.Create(refund).Error; err != nil {
			t.Fatalf("error creating refund: %v", err)
		}

		refunds, err := registry.GetRefundByPaymentID(
			context.Background(),
			payment1.ID,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(refunds) != 0 {
			t.Fatalf(
				"expected 0 refunds, got %d",
				len(refunds),
			)
		}
	})

	t.Run("returns empty slice when payment has no refunds", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		refunds, err := registry.GetRefundByPaymentID(
			context.Background(),
			utils.GenerateUUID(),
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(refunds) != 0 {
			t.Fatalf(
				"expected 0 refunds, got %d",
				len(refunds),
			)
		}
	})
}

func TestGetPaymentByID(t *testing.T) {
	t.Run("returns payment successfully", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating receiver merchant: %v", err)
		}

		currency := "INR"
		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled
		paymentReference := utils.GeneratePaymentReference()

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   paymentReference,
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("error creating payment: %v", err)
		}

		result, err := registry.GetPaymentByID(
			context.Background(),
			payment.ID,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("expected payment, got nil")
		}

		if result.ID != payment.ID {
			t.Errorf(
				"expected payment ID %v, got %v",
				payment.ID,
				result.ID,
			)
		}

		if result.SenderMerchantID != payment.SenderMerchantID {
			t.Errorf(
				"expected sender merchant ID %v, got %v",
				payment.SenderMerchantID,
				result.SenderMerchantID,
			)
		}

		if result.ReceiverMerchantID != payment.ReceiverMerchantID {
			t.Errorf(
				"expected receiver merchant ID %v, got %v",
				payment.ReceiverMerchantID,
				result.ReceiverMerchantID,
			)
		}

		if !result.Amount.Equal(payment.Amount) {
			t.Errorf(
				"expected amount %s, got %s",
				payment.Amount,
				result.Amount,
			)
		}

		if result.Currency == nil {
			t.Fatal("expected currency, got nil")
		}

		if *result.Currency != *payment.Currency {
			t.Errorf(
				"expected currency %q, got %q",
				*payment.Currency,
				*result.Currency,
			)
		}
	})

	t.Run("payment not found", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		result, err := registry.GetPaymentByID(
			context.Background(),
			utils.GenerateUUID(),
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result != nil {
			t.Errorf(
				"expected nil payment, got %+v",
				result,
			)
		}
	})
}

func TestCreateRefund(t *testing.T) {
	t.Run("refund amount exceeds remaining refundable amount", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		paymentSettlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &paymentSettlementStatus,
		}
		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("error creating payment: %v", err)
		}

		refundedStatus := constants.TransactionStatusRefunded
		refundReference := utils.GenerateRefundReference()
		reason := "previous refund"

		existingRefund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      merchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(300),
			Currency:        currency,
			Status:          refundedStatus,
			Reason:          &reason,
		}

		if err := db.Create(existingRefund).Error; err != nil {
			t.Fatalf("error creating existing refund: %v", err)
		}

		newReason := "additional refund"

		request := &dto.CreateRefundRequest{
			PaymentID:  payment.ID,
			MerchantID: merchant.ID,
			Amount:     decimal.NewFromInt(500),
			Currency:   currency,
			Reason:     &newReason,
		}

		refunds := []*models.Refund{
			existingRefund,
		}

		refund, err := registry.CreateRefund(
			context.Background(),
			request,
			refunds,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "refund amount exceeds refundable amount" {
			t.Errorf(
				"expected error %q, got %q",
				"refund amount exceeds refundable amount",
				err.Error(),
			)
		}

		if refund != nil {
			t.Errorf("expected nil refund, got %+v", refund)
		}
	})

	t.Run("nil and non-refunded refunds are ignored", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		paymentSettlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &paymentSettlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("error creating payment: %v", err)
		}

		processingStatus := constants.TransactionStatusProcessing
		reason := "processing refund"

		refundReference := utils.GenerateRefundReference()

		processingRefund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      merchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(900),
			Currency:        currency,
			Status:          processingStatus,
			Reason:          &reason,
		}

		if err := db.Create(processingRefund).Error; err != nil {
			t.Fatalf("error creating processing refund: %v", err)
		}

		newReason := "new refund"

		request := &dto.CreateRefundRequest{
			PaymentID:  payment.ID,
			MerchantID: merchant.ID,
			Amount:     decimal.NewFromInt(500),
			Currency:   currency,
			Reason:     &newReason,
		}

		refunds := []*models.Refund{
			nil,
			processingRefund,
		}

		refund, err := registry.CreateRefund(
			context.Background(),
			request,
			refunds,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if refund == nil {
			t.Fatal("expected refund, got nil")
		}

		if refund.Amount != request.Amount {
			t.Errorf(
				"expected amount %s, got %s",
				request.Amount,
				refund.Amount,
			)
		}

		if refund.Status == "" {
			t.Fatal("expected refund status, got nil")
		}

		if refund.Status != constants.TransactionStatusProcessing {
			t.Errorf(
				"expected status %q, got %q",
				constants.TransactionStatusProcessing,
				refund.Status,
			)
		}
	})

	t.Run("creates refund successfully", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		paymentSettlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &paymentSettlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("error creating payment: %v", err)
		}

		reason := "customer requested refund"

		request := &dto.CreateRefundRequest{
			PaymentID:  payment.ID,
			MerchantID: merchant.ID,
			Amount:     decimal.NewFromInt(500),
			Currency:   currency,
			Reason:     &reason,
		}

		refund, err := registry.CreateRefund(
			context.Background(),
			request,
			nil,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if refund == nil {
			t.Fatal("expected refund, got nil")
		}

		if refund.ID == uuid.Nil {
			t.Error("expected refund ID to be generated")
		}

		if refund.PaymentID != request.PaymentID {
			t.Errorf(
				"expected payment ID %v, got %v",
				request.PaymentID,
				refund.PaymentID,
			)
		}

		if refund.MerchantID != request.MerchantID {
			t.Errorf(
				"expected merchant ID %v, got %v",
				request.MerchantID,
				refund.MerchantID,
			)
		}

		if !refund.Amount.Equal(request.Amount) {
			t.Errorf(
				"expected amount %s, got %s",
				request.Amount,
				refund.Amount,
			)
		}

		if refund.Currency == "" {
			t.Fatal("expected currency, got nil")
		}

		if refund.Currency != request.Currency {
			t.Errorf(
				"expected currency %q, got %q",
				request.Currency,
				refund.Currency,
			)
		}

		if refund.Reason == nil {
			t.Fatal("expected reason, got nil")
		}

		if *refund.Reason != *request.Reason {
			t.Errorf(
				"expected reason %q, got %q",
				*request.Reason,
				*refund.Reason,
			)
		}

		if refund.Status == "" {
			t.Fatal("expected refund status, got nil")
		}

		if refund.Status != constants.TransactionStatusProcessing {
			t.Errorf(
				"expected status %q, got %q",
				constants.TransactionStatusProcessing,
				refund.Status,
			)
		}

		var persistedRefund models.Refund
		if err := db.
			Where("id = ?", refund.ID).
			First(&persistedRefund).Error; err != nil {
			t.Fatalf("failed to fetch persisted refund: %v", err)
		}

		if persistedRefund.ID != refund.ID {
			t.Errorf(
				"expected persisted refund ID %v, got %v",
				refund.ID,
				persistedRefund.ID,
			)
		}

		if !persistedRefund.Amount.Equal(request.Amount) {
			t.Errorf(
				"expected persisted amount %s, got %s",
				request.Amount,
				persistedRefund.Amount,
			)
		}

		if persistedRefund.Status == "" {
			t.Fatal("expected persisted status, got nil")
		}

		if persistedRefund.Status != constants.TransactionStatusProcessing {
			t.Errorf(
				"expected persisted status %q, got %q",
				constants.TransactionStatusProcessing,
				persistedRefund.Status,
			)
		}
	})
}

func TestRefundFromMerchantWallet(t *testing.T) {
	t.Run("refund is nil", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		payment := &models.Payment{
			ID: utils.GenerateUUID(),
		}

		refund, err := registry.RefundFromMerchantWallet(
			context.Background(),
			nil,
			payment,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if refund != nil {
			t.Errorf("expected nil refund, got %+v", refund)
		}
	})

	t.Run("payment is nil", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		refund := &models.Refund{
			ID:     utils.GenerateUUID(),
			Amount: decimal.NewFromInt(100),
		}

		result, err := registry.RefundFromMerchantWallet(
			context.Background(),
			refund,
			nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if result != nil {
			t.Errorf("expected nil refund, got %+v", result)
		}
	})

	t.Run("successful refund from merchant wallet", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("error creating sender merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating receiver merchant: %v", err)
		}

		currency := "INR"
		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("error creating payment: %v", err)
		}

		statusActive := constants.MerchantStatusActive

		senderWallet := &models.Wallets{
			ID:               utils.GenerateUUID(),
			MerchantID:       receiverMerchant.ID,
			AvailableBalance: decimal.NewFromInt(1000),
			ReservedBalance:  decimal.Zero,
			Status:           statusActive,
		}

		receiverWallet := &models.Wallets{
			ID:               utils.GenerateUUID(),
			MerchantID:       senderMerchant.ID,
			AvailableBalance: decimal.NewFromInt(500),
			ReservedBalance:  decimal.Zero,
			Status:           statusActive,
		}

		if err := db.Create(senderWallet).Error; err != nil {
			t.Fatalf("error creating sender wallet: %v", err)
		}

		if err := db.Create(receiverWallet).Error; err != nil {
			t.Fatalf("error creating receiver wallet: %v", err)
		}

		paymentVaultType := constants.PaymentVault
		vaultStatus := constants.VaultStatusActive

		paymentVault := &models.Vault{
			ID:      utils.GenerateUUID(),
			Type:    paymentVaultType,
			Balance: decimal.NewFromInt(5000),
			Status:  vaultStatus,
		}

		if err := db.Create(paymentVault).Error; err != nil {
			t.Fatalf("error creating payment vault: %v", err)
		}

		ledgerRef := utils.GenerateLedgerReference()
		paymentRef := payment.PaymentReference
		ledgerTransaction := &models.LedgerTransaction{
			ID:             utils.GenerateUUID(),
			TransactionRef: ledgerRef,
			ReferenceID:    *paymentRef,
		}

		if err := db.Create(ledgerTransaction).Error; err != nil {
			t.Fatalf("error creating ledger transaction: %v", err)
		}

		refundReference := utils.GenerateRefundReference()

		refundStatus := constants.TransactionStatusProcessing

		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      senderMerchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(200),
			Currency:        currency,
			Status:          refundStatus,
		}

		if err := db.Create(refund).Error; err != nil {
			t.Fatalf("error creating refund: %v", err)
		}

		result, err := registry.RefundFromMerchantWallet(
			context.Background(),
			refund,
			payment,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("expected refund, got nil")
		}

		if result.Status == "" {
			t.Fatal("expected refund status, got nil")
		}

		if result.Status != constants.TransactionStatusRefunded {
			t.Errorf(
				"expected refund status %q, got %q",
				constants.TransactionStatusRefunded,
				result.Status,
			)
		}

		var updatedSenderWallet models.Wallets
		if err := db.Where("id = ?", senderWallet.ID).
			First(&updatedSenderWallet).Error; err != nil {
			t.Fatalf("failed to fetch sender wallet: %v", err)
		}

		if !updatedSenderWallet.AvailableBalance.Equal(decimal.NewFromInt(800)) {
			t.Errorf(
				"expected sender available balance 800, got %s",
				updatedSenderWallet.AvailableBalance,
			)
		}

		var updatedReceiverWallet models.Wallets
		if err := db.Where("id = ?", receiverWallet.ID).
			First(&updatedReceiverWallet).Error; err != nil {
			t.Fatalf("failed to fetch receiver wallet: %v", err)
		}

		if !updatedReceiverWallet.AvailableBalance.Equal(decimal.NewFromInt(700)) {
			t.Errorf(
				"expected receiver available balance 700, got %s",
				updatedReceiverWallet.AvailableBalance,
			)
		}

		var updatedVault models.Vault
		if err := db.Where("id = ?", paymentVault.ID).
			First(&updatedVault).Error; err != nil {
			t.Fatalf("failed to fetch payment vault: %v", err)
		}

		if !updatedVault.Balance.Equal(decimal.NewFromInt(5000)) {
			t.Errorf(
				"expected payment vault balance 5000, got %s",
				updatedVault.Balance,
			)
		}

		var refundLedgerTransaction models.LedgerTransaction

		if err := db.
			Where("reference_id = ?", refund.RefundReference).
			First(&refundLedgerTransaction).Error; err != nil {
			t.Fatalf(
				"failed to fetch refund ledger transaction: %v",
				err,
			)
		}

		var ledgerEntries []models.LedgerEntry

		if err := db.
			Where("ledger_transaction_id = ?", refundLedgerTransaction.ID).
			Find(&ledgerEntries).Error; err != nil {
			t.Fatalf(
				"failed to fetch refund ledger entries: %v",
				err,
			)
		}

		if len(ledgerEntries) != 4 {
			t.Fatalf(
				"expected 4 ledger entries, got %d",
				len(ledgerEntries),
			)
		}
	})

	t.Run("insufficient sender wallet triggers top up", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("error creating sender merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating receiver merchant: %v", err)
		}

		currency := "INR"
		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("error creating payment: %v", err)
		}

		activeStatus := "active"

		senderWallet := &models.Wallets{
			ID:               utils.GenerateUUID(),
			MerchantID:       receiverMerchant.ID,
			AvailableBalance: decimal.NewFromInt(100),
			ReservedBalance:  decimal.Zero,
			Status:           activeStatus,
		}

		receiverWallet := &models.Wallets{
			ID:               utils.GenerateUUID(),
			MerchantID:       senderMerchant.ID,
			AvailableBalance: decimal.NewFromInt(500),
			ReservedBalance:  decimal.Zero,
			Status:           activeStatus,
		}

		if err := db.Create(senderWallet).Error; err != nil {
			t.Fatalf("error creating sender wallet: %v", err)
		}

		if err := db.Create(receiverWallet).Error; err != nil {
			t.Fatalf("error creating receiver wallet: %v", err)
		}

		accountNumber := "1234567890"
		accountName := "Sender Merchant"
		bankName := "Test Bank"
		ifscCode := "TEST0001234"
		accountType := "savings"
		bankStatus := "active"

		bankAccount := &models.BankAccount{
			ID:            utils.GenerateUUID(),
			AccountNumber: accountNumber,
			AccountName:   accountName,
			BankName:      bankName,
			IFSCCode:      ifscCode,
			AccountType:   accountType,
			Balance:       decimal.NewFromInt(2000),
			Status:        bankStatus,
		}

		if err := db.Create(bankAccount).Error; err != nil {
			t.Fatalf("error creating bank account: %v", err)
		}

		primaryType := constants.BankAccountTypePrimary
		linkedStatus := "active"

		linkedBankAccount := &models.LinkedBankAccount{
			ID:            utils.GenerateUUID(),
			MerchantID:    receiverMerchant.ID,
			BankAccountID: bankAccount.ID,
			Type:          primaryType,
			Status:        linkedStatus,
		}

		if err := db.Create(linkedBankAccount).Error; err != nil {
			t.Fatalf("error creating linked bank account: %v", err)
		}

		paymentVaultType := constants.PaymentVault
		vaultStatus := constants.VaultStatusActive

		paymentVault := &models.Vault{
			ID:      utils.GenerateUUID(),
			Type:    paymentVaultType,
			Balance: decimal.NewFromInt(5000),
			Status:  vaultStatus,
		}

		if err := db.Create(paymentVault).Error; err != nil {
			t.Fatalf("error creating payment vault: %v", err)
		}

		refundStatus := constants.TransactionStatusProcessing
		refundReference := utils.GenerateRefundReference()

		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      receiverMerchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(200),
			Currency:        currency,
			Status:          refundStatus,
		}

		if err := db.Create(refund).Error; err != nil {
			t.Fatalf("error creating refund: %v", err)
		}

		_, err := registry.RefundFromMerchantWallet(
			context.Background(),
			refund,
			payment,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		var updatedBankAccount models.BankAccount
		if err := db.Where("id = ?", bankAccount.ID).
			First(&updatedBankAccount).Error; err != nil {
			t.Fatalf("failed to fetch bank account: %v", err)
		}

		if !updatedBankAccount.Balance.Equal(decimal.NewFromInt(1800)) {
			t.Errorf(
				"expected bank balance 1900, got %s",
				updatedBankAccount.Balance,
			)
		}

		var updatedSenderWallet models.Wallets
		if err := db.Where("id = ?", senderWallet.ID).
			First(&updatedSenderWallet).Error; err != nil {
			t.Fatalf("failed to fetch sender wallet: %v", err)
		}
	})

	t.Run("receiver wallet is missing", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("error creating sender merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating receiver merchant: %v", err)
		}

		currency := "INR"

		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled
		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}

		refundReference := utils.GenerateRefundReference()

		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      receiverMerchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(100),
		}

		_, err := registry.RefundFromMerchantWallet(
			context.Background(),
			refund,
			payment,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("payment vault is missing", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("error creating sender merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating receiver merchant: %v", err)
		}

		activeStatus := "active"

		senderWallet := &models.Wallets{
			ID:               utils.GenerateUUID(),
			MerchantID:       receiverMerchant.ID,
			AvailableBalance: decimal.NewFromInt(1000),
			Status:           activeStatus,
		}

		receiverWallet := &models.Wallets{
			ID:               utils.GenerateUUID(),
			MerchantID:       senderMerchant.ID,
			AvailableBalance: decimal.NewFromInt(500),
			Status:           activeStatus,
		}

		if err := db.Create(senderWallet).Error; err != nil {
			t.Fatalf("error creating sender wallet: %v", err)
		}

		if err := db.Create(receiverWallet).Error; err != nil {
			t.Fatalf("error creating receiver wallet: %v", err)
		}

		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}
		refundReference := utils.GenerateRefundReference()
		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      receiverMerchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(100),
		}

		_, err := registry.RefundFromMerchantWallet(
			context.Background(),
			refund,
			payment,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("payment vault has insufficient balance", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("error creating sender merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating receiver merchant: %v", err)
		}

		activeStatus := "active"

		senderWallet := &models.Wallets{
			ID:               utils.GenerateUUID(),
			MerchantID:       receiverMerchant.ID,
			AvailableBalance: decimal.NewFromInt(1000),
			Status:           activeStatus,
		}

		receiverWallet := &models.Wallets{
			ID:               utils.GenerateUUID(),
			MerchantID:       senderMerchant.ID,
			AvailableBalance: decimal.NewFromInt(500),
			Status:           activeStatus,
		}

		if err := db.Create(senderWallet).Error; err != nil {
			t.Fatalf("error creating sender wallet: %v", err)
		}

		if err := db.Create(receiverWallet).Error; err != nil {
			t.Fatalf("error creating receiver wallet: %v", err)
		}

		paymentVaultType := constants.PaymentVault
		vaultStatus := constants.VaultStatusActive

		paymentVault := &models.Vault{
			ID:      utils.GenerateUUID(),
			Type:    paymentVaultType,
			Balance: decimal.NewFromInt(50),
			Status:  vaultStatus,
		}

		if err := db.Create(paymentVault).Error; err != nil {
			t.Fatalf("error creating payment vault: %v", err)
		}

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
		}

		refundReference := utils.GenerateRefundReference()
		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      receiverMerchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(100),
		}

		_, err := registry.RefundFromMerchantWallet(
			context.Background(),
			refund,
			payment,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestRefundFromPaymentVault(t *testing.T) {
	t.Run("refund is nil", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		result, err := registry.RefundFromPaymentVault(
			context.Background(),
			nil,
			&models.Payment{},
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "refund not found" {
			t.Errorf(
				"expected error %q, got %q",
				"refund not found",
				err.Error(),
			)
		}

		if result != nil {
			t.Errorf("expected nil refund, got %+v", result)
		}
	})

	t.Run("payment is nil", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		refundReference := utils.GenerateRefundReference()
		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(200),
			Currency:        "INR",
			Status:          constants.TransactionStatusProcessing,
		}

		result, err := registry.RefundFromPaymentVault(
			context.Background(),
			refund,
			nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "payment not found" {
			t.Errorf(
				"expected error %q, got %q",
				"payment not found",
				err.Error(),
			)
		}

		if result != nil {
			t.Errorf("expected nil refund, got %+v", result)
		}
	})

	t.Run("refund receiver wallet not found", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		currency := "INR"
		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: utils.GenerateUUID(),
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}

		refundReference := utils.GenerateRefundReference()
		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      senderMerchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(200),
			Currency:        currency,
			Status:          constants.TransactionStatusProcessing,
		}

		_, err := registry.RefundFromPaymentVault(
			context.Background(),
			refund,
			payment,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("payment vault not found", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		activeStatus := constants.MerchantStatusActive

		wallet := &models.Wallets{
			ID:               utils.GenerateUUID(),
			MerchantID:       senderMerchant.ID,
			AvailableBalance: decimal.NewFromInt(500),
			ReservedBalance:  decimal.Zero,
			Status:           activeStatus,
		}

		if err := db.Create(wallet).Error; err != nil {
			t.Fatalf("error creating wallet: %v", err)
		}

		currency := "INR"
		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: utils.GenerateUUID(),
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}

		refundReference := utils.GenerateRefundReference()
		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      senderMerchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(200),
			Currency:        currency,
			Status:          constants.TransactionStatusProcessing,
		}

		_, err := registry.RefundFromPaymentVault(
			context.Background(),
			refund,
			payment,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("insufficient payment vault balance", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("error creating merchant: %v", err)
		}

		activeStatus := constants.MerchantStatusActive

		wallet := &models.Wallets{
			ID:               utils.GenerateUUID(),
			MerchantID:       senderMerchant.ID,
			AvailableBalance: decimal.NewFromInt(500),
			ReservedBalance:  decimal.Zero,
			Status:           activeStatus,
		}

		if err := db.Create(wallet).Error; err != nil {
			t.Fatalf("error creating wallet: %v", err)
		}

		paymentVaultType := constants.PaymentVault
		vaultStatus := constants.VaultStatusActive

		paymentVault := &models.Vault{
			ID:      utils.GenerateUUID(),
			Type:    paymentVaultType,
			Balance: decimal.NewFromInt(100),
			Status:  vaultStatus,
		}

		if err := db.Create(paymentVault).Error; err != nil {
			t.Fatalf("error creating payment vault: %v", err)
		}

		currency := "INR"
		status := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: utils.GenerateUUID(),
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &status,
			SettlementStatus:   &settlementStatus,
		}

		refundReference := utils.GenerateRefundReference()
		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      senderMerchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(200),
			Currency:        currency,
			Status:          constants.TransactionStatusProcessing,
		}

		_, err := registry.RefundFromPaymentVault(
			context.Background(),
			refund,
			payment,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var updatedVault models.Vault
		if err := db.Where("id = ?", paymentVault.ID).
			First(&updatedVault).Error; err != nil {
			t.Fatalf("failed to fetch payment vault: %v", err)
		}

		if !updatedVault.Balance.Equal(decimal.NewFromInt(100)) {
			t.Errorf(
				"expected vault balance 100, got %s",
				updatedVault.Balance,
			)
		}
	})

	t.Run("successful refund from payment vault", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		senderMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(senderMerchant).Error; err != nil {
			t.Fatalf("error creating sender merchant: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("error creating sender merchant: %v", err)
		}

		activeStatus := constants.MerchantStatusActive

		receiverWallet := &models.Wallets{
			ID:               utils.GenerateUUID(),
			MerchantID:       senderMerchant.ID,
			AvailableBalance: decimal.NewFromInt(500),
			ReservedBalance:  decimal.Zero,
			Status:           activeStatus,
		}

		if err := db.Create(receiverWallet).Error; err != nil {
			t.Fatalf("error creating receiver wallet: %v", err)
		}

		paymentVaultType := constants.PaymentVault
		vaultStatus := constants.VaultStatusActive

		paymentVault := &models.Vault{
			ID:      utils.GenerateUUID(),
			Type:    paymentVaultType,
			Balance: decimal.NewFromInt(5000),
			Status:  vaultStatus,
		}

		if err := db.Create(paymentVault).Error; err != nil {
			t.Fatalf("error creating payment vault: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementSettled

		payment := &models.Payment{
			ID:                 utils.GenerateUUID(),
			SenderMerchantID:   senderMerchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   utils.GeneratePaymentReference(),
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &settlementStatus,
		}

		if errP := db.Create(payment).Error; errP != nil {
			t.Fatalf("payment creation failed: %v", errP)
		}

		refundReference := utils.GenerateRefundReference()
		refundStatus := constants.TransactionStatusProcessing

		refund := &models.Refund{
			ID:              utils.GenerateUUID(),
			PaymentID:       payment.ID,
			MerchantID:      senderMerchant.ID,
			RefundReference: *refundReference,
			Amount:          decimal.NewFromInt(200),
			Currency:        currency,
			Status:          refundStatus,
		}

		if err := db.Create(refund).Error; err != nil {
			t.Fatalf("error creating refund: %v", err)
		}

		result, err := registry.RefundFromPaymentVault(
			context.Background(),
			refund,
			payment,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("expected refund, got nil")
		}

		if result.Status != constants.TransactionStatusRefunded {
			t.Errorf(
				"expected refund status %q, got %q",
				constants.TransactionStatusRefunded,
				result.Status,
			)
		}

		var updatedWallet models.Wallets
		if err := db.Where("id = ?", receiverWallet.ID).
			First(&updatedWallet).Error; err != nil {
			t.Fatalf("failed to fetch receiver wallet: %v", err)
		}

		if !updatedWallet.AvailableBalance.Equal(decimal.NewFromInt(700)) {
			t.Errorf(
				"expected receiver available balance 700, got %s",
				updatedWallet.AvailableBalance,
			)
		}

		var updatedVault models.Vault
		if err := db.Where("id = ?", paymentVault.ID).
			First(&updatedVault).Error; err != nil {
			t.Fatalf("failed to fetch payment vault: %v", err)
		}

		if !updatedVault.Balance.Equal(decimal.NewFromInt(4800)) {
			t.Errorf(
				"expected payment vault balance 4800, got %s",
				updatedVault.Balance,
			)
		}

		var refundLedgerTransaction models.LedgerTransaction

		if err := db.
			Where("reference_id = ?", refund.RefundReference).
			First(&refundLedgerTransaction).Error; err != nil {
			t.Fatalf(
				"failed to fetch refund ledger transaction: %v",
				err,
			)
		}

		if refundLedgerTransaction.Status != constants.TransactionStatusCompleted {
			t.Errorf(
				"expected ledger status %q, got %q",
				constants.TransactionStatusCompleted,
				refundLedgerTransaction.Status,
			)
		}

		if refundLedgerTransaction.SettlementStatus != constants.LedgerSettlementSettled {
			t.Errorf(
				"expected ledger settlement status %q, got %q",
				constants.LedgerSettlementSettled,
				refundLedgerTransaction.SettlementStatus,
			)
		}

		var ledgerEntries []models.LedgerEntry

		if err := db.
			Where("ledger_transaction_id = ?", refundLedgerTransaction.ID).
			Find(&ledgerEntries).Error; err != nil {
			t.Fatalf(
				"failed to fetch refund ledger entries: %v",
				err,
			)
		}

		if len(ledgerEntries) != 2 {
			t.Fatalf(
				"expected 2 ledger entries, got %d",
				len(ledgerEntries),
			)
		}

		var debitFound bool
		var creditFound bool

		for _, entry := range ledgerEntries {
			if entry.AccountType == constants.LedgerAccountTypeVault &&
				entry.EntryType == constants.LedgerEntryTypeDebit &&
				entry.Amount.Equal(decimal.NewFromInt(200)) {
				debitFound = true
			}

			if entry.AccountType == constants.LedgerAccountTypeWallet &&
				entry.EntryType == constants.LedgerEntryTypeCredit &&
				entry.Amount.Equal(decimal.NewFromInt(200)) {
				creditFound = true
			}
		}

		if !debitFound {
			t.Error("expected vault debit ledger entry")
		}

		if !creditFound {
			t.Error("expected wallet credit ledger entry")
		}
	})
}
