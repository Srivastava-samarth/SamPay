package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCreateRefund(t *testing.T) {
	t.Run("invalid merchant id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST(
			"/merchants/:merchant_id/refunds",
			testController.CreateRefund(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/not-a-uuid/refunds",
			strings.NewReader(`{}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "test-key")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success=false")
		}

		if response.Error == nil {
			t.Fatalf("expected error in response")
		}

		if response.Error.Code != "MERCHANT_ID_PARSING_ERROR" {
			t.Errorf(
				"expected MERCHANT_ID_PARSING_ERROR, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("missing idempotency key", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.POST(
			"/merchants/:merchant_id/refunds",
			testController.CreateRefund(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/"+merchant.ID.String()+"/refunds",
			strings.NewReader(`{}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success=false")
		}

		if response.Error == nil {
			t.Fatalf("expected error in response")
		}

		if response.Error.Code != "IDEMPOTENCY_KEY_NOT_FOUND" {
			t.Errorf(
				"expected IDEMPOTENCY_KEY_NOT_FOUND, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("existing idempotency key", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		idempotencyKey := "existing-refund-key"

		storedResponse := dto.Response{
			Success: true,
			Data:    "existing-refund-response",
		}

		responseBody, err := json.Marshal(storedResponse)
		if err != nil {
			t.Fatalf("failed to marshal stored response: %v", err)
		}

		idempotency := &models.IdempotencyKey{
			ID:             uuid.New(),
			MerchantID:     merchant.ID,
			IdempotencyKey: idempotencyKey,
			ResponseBody:   make([]byte, len(responseBody)),
		}

		for i := range responseBody {
			idempotency.ResponseBody[i] = responseBody[i]
		}

		if err := db.Create(idempotency).Error; err != nil {
			t.Fatalf("failed to create idempotency key: %v", err)
		}

		router := gin.New()
		router.POST(
			"/merchants/:merchant_id/refunds",
			testController.CreateRefund(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/"+merchant.ID.String()+"/refunds",
			strings.NewReader(`{}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success=true")
		}

		if response.Data == nil {
			t.Fatalf("expected existing response data")
		}

		data, ok := response.Data.(string)
		if !ok {
			t.Fatalf("expected response data to be string, got %T", response.Data)
		}

		if data != "existing-refund-response" {
			t.Errorf(
				"expected existing-refund-response, got %s",
				data,
			)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		idempotencyKey := "invalid-body-key"

		router := gin.New()
		router.POST(
			"/merchants/:merchant_id/refunds",
			testController.CreateRefund(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/"+merchant.ID.String()+"/refunds",
			strings.NewReader(`{"payment_id":`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success=false")
		}

		if response.Error == nil {
			t.Fatalf("expected error in response")
		}

		if response.Error.Code != "BINDING_ISSUE" {
			t.Errorf(
				"expected BINDING_ISSUE, got %s",
				response.Error.Code,
			)
		}

		// Controller should save the failed response for idempotency.
		var savedIdempotency models.IdempotencyKey
		if err := db.
			Where("merchant_id = ? AND idempotency_key = ?", merchant.ID, idempotencyKey).
			First(&savedIdempotency).Error; err != nil {
			t.Fatalf(
				"expected idempotency response to be saved: %v",
				err,
			)
		}
	})

	t.Run("successful refund from payment vault", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		senderWallet := testutils.GenerateTestWallet(merchant.ID)
		if err := db.Create(senderWallet).Error; err != nil {
			t.Fatalf("failed to create sender wallet: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		receiverWallet := testutils.GenerateTestWallet(receiverMerchant.ID)
		receiverWallet.AvailableBalance = decimal.Zero

		if err := db.Create(receiverWallet).Error; err != nil {
			t.Fatalf("failed to create receiver wallet: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementPending
		paymentReference := "payment-ref-" + uuid.New().String()

		payment := &models.Payment{
			ID:                 uuid.New(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   &paymentReference,
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &settlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		paymentVault := &models.Vault{
			ID:      uuid.New(),
			Type:    constants.PaymentVault,
			Balance: decimal.NewFromInt(5000),
		}

		if err := db.Create(paymentVault).Error; err != nil {
			t.Fatalf("failed to create payment vault: %v", err)
		}

		reason := "customer requested refund"
		externalReference := "external-ref-" + uuid.New().String()
		idempotencyKey := "successful-refund-key"

		requestBody := dto.CreateRefundRequest{
			PaymentID:         payment.ID,
			MerchantID:        merchant.ID,
			ExternalReference: &externalReference,
			Amount:            decimal.NewFromInt(500),
			Currency:          currency,
			Reason:            &reason,
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		router := gin.New()
		router.POST(
			"/merchants/:merchant_id/refunds",
			testController.CreateRefund(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/"+merchant.ID.String()+"/refunds",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status 200, got %d. response: %s",
				rec.Code,
				rec.Body.String(),
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success=true")
		}

		if response.Data == nil {
			t.Fatalf("expected refund data")
		}

		var refund models.Refund
		refundData, err := json.Marshal(response.Data)
		if err != nil {
			t.Fatalf("failed to marshal refund response: %v", err)
		}

		if err := json.Unmarshal(refundData, &refund); err != nil {
			t.Fatalf("failed to decode refund response: %v", err)
		}

		if refund.PaymentID != payment.ID {
			t.Errorf(
				"expected payment id %s, got %s",
				payment.ID,
				refund.PaymentID,
			)
		}

		if !refund.Amount.Equal(decimal.NewFromInt(500)) {
			t.Errorf(
				"expected refund amount 500, got %s",
				refund.Amount,
			)
		}

		if refund.Status != constants.TransactionStatusRefunded {
			t.Errorf(
				"expected refund status %s, got %s",
				constants.TransactionStatusRefunded,
				refund.Status,
			)
		}

		var updatedVault models.Vault
		if err := db.
			Where("id = ?", paymentVault.ID).
			First(&updatedVault).Error; err != nil {
			t.Fatalf("failed to fetch payment vault: %v", err)
		}

		if !updatedVault.Balance.Equal(decimal.NewFromInt(4500)) {
			t.Errorf(
				"expected payment vault balance 4500, got %s",
				updatedVault.Balance,
			)
		}

		var updatedWallet models.Wallets
		if err := db.
			Where("id = ?", senderWallet.ID).
			First(&updatedWallet).Error; err != nil {
			t.Fatalf("failed to fetch receiver wallet: %v", err)
		}

		if !updatedWallet.AvailableBalance.Equal(decimal.NewFromInt(5500)) {
			t.Errorf(
				"expected receiver wallet balance 500, got %s",
				updatedWallet.AvailableBalance,
			)
		}

		var savedIdempotency models.IdempotencyKey
		if err := db.
			Where("merchant_id = ? AND idempotency_key = ?", merchant.ID, idempotencyKey).
			First(&savedIdempotency).Error; err != nil {
			t.Fatalf(
				"expected successful response to be saved in idempotency table: %v",
				err,
			)
		}
	})
}

func TestGetRefundByReference(t *testing.T) {
	t.Run("refund reference not provided", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.GET(
			"/refunds/:refund_reference",
			testController.GetRefundByReference(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/refunds/",
			nil,
		)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("refund found", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		senderWallet := testutils.GenerateTestWallet(merchant.ID)
		if err := db.Create(senderWallet).Error; err != nil {
			t.Fatalf("failed to create sender wallet: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		receiverWallet := testutils.GenerateTestWallet(receiverMerchant.ID)
		receiverWallet.AvailableBalance = decimal.Zero

		if err := db.Create(receiverWallet).Error; err != nil {
			t.Fatalf("failed to create receiver wallet: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementPending
		paymentReference := "payment-ref-" + uuid.New().String()

		payment := &models.Payment{
			ID:                 uuid.New(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   &paymentReference,
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &settlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}
		refundReference := "refund-" + uuid.New().String()

		reason := "test 1"

		refund := &models.Refund{
			ID:              uuid.New(),
			PaymentID:       payment.ID,
			MerchantID:      merchant.ID,
			RefundReference: refundReference,
			Amount:          decimal.NewFromInt(500),
			Currency:        currency,
			Reason:          &reason,
			Status:          constants.TransactionStatusRefunded,
		}

		if err := db.Create(refund).Error; err != nil {
			t.Fatalf("failed to create refund: %v", err)
		}

		router := gin.New()
		router.GET(
			"/refunds/:refund_reference",
			testController.GetRefundByReference(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/refunds/"+refundReference,
			nil,
		)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status 200, got %d. response: %s",
				rec.Code,
				rec.Body.String(),
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success=true")
		}

		if response.Data == nil {
			t.Fatalf("expected refund data")
		}

		var returnedRefund models.Refund
		refundData, err := json.Marshal(response.Data)
		if err != nil {
			t.Fatalf("failed to marshal refund response: %v", err)
		}

		if err := json.Unmarshal(refundData, &returnedRefund); err != nil {
			t.Fatalf("failed to decode refund response: %v", err)
		}

		if returnedRefund.ID != refund.ID {
			t.Errorf(
				"expected refund ID %s, got %s",
				refund.ID,
				returnedRefund.ID,
			)
		}

		if returnedRefund.RefundReference != refundReference {
			t.Errorf(
				"expected refund reference %s, got %s",
				refundReference,
				returnedRefund.RefundReference,
			)
		}

		if !returnedRefund.Amount.Equal(refund.Amount) {
			t.Errorf(
				"expected refund amount %s, got %s",
				refund.Amount,
				returnedRefund.Amount,
			)
		}
	})
}

func TestGetRefundById(t *testing.T) {

	t.Run("refund id not provided", func(t *testing.T) {
		cleanTestDB(t)

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		c.Params = gin.Params{
			{Key: "refund_id", Value: ""},
		}

		testController.GetRefundById()(c)

		if rec.Code != http.StatusNotFound {
			t.Errorf(
				"expected status 404, got %d",
				rec.Code,
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success=false")
		}

		if response.Error == nil {
			t.Fatalf("expected error in response")
		}

		if response.Error.Code != "REFUND_ID_NOT_FOUND" {
			t.Errorf(
				"expected REFUND_ID_NOT_FOUND, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("invalid refund id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.GET(
			"/refunds/:refund_id",
			testController.GetRefundById(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/refunds/not-a-uuid",
			nil,
		)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success=false")
		}

		if response.Error == nil {
			t.Fatalf("expected error in response")
		}

		if response.Error.Code != "REFUND_ID_PARSING_ERROR" {
			t.Errorf(
				"expected REFUND_ID_PARSING_ERROR, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("refund found", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		senderWallet := testutils.GenerateTestWallet(merchant.ID)
		if err := db.Create(senderWallet).Error; err != nil {
			t.Fatalf("failed to create sender wallet: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		receiverWallet := testutils.GenerateTestWallet(receiverMerchant.ID)
		receiverWallet.AvailableBalance = decimal.Zero

		if err := db.Create(receiverWallet).Error; err != nil {
			t.Fatalf("failed to create receiver wallet: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementPending
		paymentReference := "payment-ref-" + uuid.New().String()

		payment := &models.Payment{
			ID:                 uuid.New(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   &paymentReference,
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &settlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}
		refundReference := "refund-" + uuid.New().String()
		reason := "customer requested refund"

		refund := &models.Refund{
			ID:              uuid.New(),
			PaymentID:       payment.ID,
			MerchantID:      merchant.ID,
			RefundReference: refundReference,
			Amount:          decimal.NewFromInt(500),
			Currency:        currency,
			Reason:          &reason,
			Status:          constants.TransactionStatusRefunded,
		}

		if err := db.Create(refund).Error; err != nil {
			t.Fatalf("failed to create refund: %v", err)
		}

		router := gin.New()
		router.GET(
			"/refunds/:refund_id",
			testController.GetRefundById(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/refunds/"+refund.ID.String(),
			nil,
		)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status 200, got %d. response: %s",
				rec.Code,
				rec.Body.String(),
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success=true")
		}

		if response.Data == nil {
			t.Fatalf("expected refund data")
		}

		var returnedRefund models.Refund
		refundData, err := json.Marshal(response.Data)
		if err != nil {
			t.Fatalf("failed to marshal refund response: %v", err)
		}

		if err := json.Unmarshal(refundData, &returnedRefund); err != nil {
			t.Fatalf("failed to decode refund response: %v", err)
		}

		if returnedRefund.ID != refund.ID {
			t.Errorf(
				"expected refund ID %s, got %s",
				refund.ID,
				returnedRefund.ID,
			)
		}

		if returnedRefund.RefundReference != refund.RefundReference {
			t.Errorf(
				"expected refund reference %s, got %s",
				refund.RefundReference,
				returnedRefund.RefundReference,
			)
		}

		if !returnedRefund.Amount.Equal(refund.Amount) {
			t.Errorf(
				"expected refund amount %s, got %s",
				refund.Amount,
				returnedRefund.Amount,
			)
		}
	})
}

func TestGetRefundByMerchantId(t *testing.T) {
	t.Run("merchant id not provided", func(t *testing.T) {
		cleanTestDB(t)

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		c.Params = gin.Params{
			{Key: "merchant_id", Value: ""},
		}

		testController.GetRefundByMerchantId()(c)

		if rec.Code != http.StatusNotFound {
			t.Errorf(
				"expected status 404, got %d",
				rec.Code,
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success=false")
		}

		if response.Error == nil {
			t.Fatalf("expected error in response")
		}

		if response.Error.Code != "MERCHANT_ID_NOT_FOUND" {
			t.Errorf(
				"expected MERCHANT_ID_NOT_FOUND, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("invalid merchant id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id/refunds",
			testController.GetRefundByMerchantId(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/not-a-uuid/refunds",
			nil,
		)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success=false")
		}

		if response.Error == nil {
			t.Fatalf("expected error in response")
		}

		if response.Error.Code != "PARSING_ERROR" {
			t.Errorf(
				"expected PARSING_ERROR, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("merchant has no refunds", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id/refunds",
			testController.GetRefundByMerchantId(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/"+merchant.ID.String()+"/refunds",
			nil,
		)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status 200, got %d. response: %s",
				rec.Code,
				rec.Body.String(),
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success=true")
		}

		if response.Data == nil {
			t.Fatalf("expected data in response")
		}

		data, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatalf(
				"expected response data to be object, got %T",
				response.Data,
			)
		}

		refunds, ok := data["data"].([]interface{})
		if !ok {
			t.Fatalf(
				"expected refunds data to be array, got %T",
				data["data"],
			)
		}

		if len(refunds) != 0 {
			t.Errorf(
				"expected 0 refunds, got %d",
				len(refunds),
			)
		}
	})

	t.Run("merchant has refunds", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		senderWallet := testutils.GenerateTestWallet(merchant.ID)
		if err := db.Create(senderWallet).Error; err != nil {
			t.Fatalf("failed to create sender wallet: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()
		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		receiverWallet := testutils.GenerateTestWallet(receiverMerchant.ID)
		receiverWallet.AvailableBalance = decimal.Zero

		if err := db.Create(receiverWallet).Error; err != nil {
			t.Fatalf("failed to create receiver wallet: %v", err)
		}

		currency := "INR"
		paymentStatus := constants.TransactionStatusCompleted
		settlementStatus := constants.LedgerSettlementPending
		paymentReference := "payment-ref-" + uuid.New().String()

		payment := &models.Payment{
			ID:                 uuid.New(),
			SenderMerchantID:   merchant.ID,
			ReceiverMerchantID: receiverMerchant.ID,
			PaymentReference:   &paymentReference,
			Amount:             decimal.NewFromInt(1000),
			Currency:           &currency,
			Status:             &paymentStatus,
			SettlementStatus:   &settlementStatus,
		}

		if err := db.Create(payment).Error; err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		reason1 := "customer requested refund"
		reason2 := "duplicate payment"

		refund1 := &models.Refund{
			ID:              uuid.New(),
			PaymentID:       payment.ID,
			MerchantID:      merchant.ID,
			RefundReference: "refund-" + uuid.New().String(),
			Amount:          decimal.NewFromInt(500),
			Currency:        currency,
			Reason:          &reason1,
			Status:          constants.TransactionStatusRefunded,
		}

		refund2 := &models.Refund{
			ID:              uuid.New(),
			PaymentID:       payment.ID,
			MerchantID:      merchant.ID,
			RefundReference: "refund-" + uuid.New().String(),
			Amount:          decimal.NewFromInt(250),
			Currency:        currency,
			Reason:          &reason2,
			Status:          constants.TransactionStatusProcessing,
		}

		if err := db.Create(refund1).Error; err != nil {
			t.Fatalf("failed to create refund1: %v", err)
		}

		if err := db.Create(refund2).Error; err != nil {
			t.Fatalf("failed to create refund2: %v", err)
		}

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id/refunds",
			testController.GetRefundByMerchantId(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/"+merchant.ID.String()+"/refunds",
			nil,
		)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status 200, got %d. response: %s",
				rec.Code,
				rec.Body.String(),
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success=true")
		}

		if response.Data == nil {
			t.Fatalf("expected data in response")
		}

		data, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatalf(
				"expected response data to be object, got %T",
				response.Data,
			)
		}

		refunds, ok := data["data"].([]interface{})
		if !ok {
			t.Fatalf(
				"expected refunds data to be array, got %T",
				data["data"],
			)
		}

		if len(refunds) != 2 {
			t.Errorf(
				"expected 2 refunds, got %d",
				len(refunds),
			)
		}
	})
}
