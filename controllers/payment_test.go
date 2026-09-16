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

func TestCreatePayment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid merchant id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST(
			"/payments/:merchant_id",
			testController.CreatePayment(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/payments/invalid-merchant-id",
			bytes.NewBufferString(`{"amount":"100"}`),
		)
		req.Header.Set("Idempotency-Key", "test-key")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				w.Code,
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
			"/payments/:merchant_id",
			testController.CreatePayment(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/payments/"+merchant.ID.String(),
			bytes.NewBufferString(`{"amount":"100"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				w.Code,
			)
		}
	})

	t.Run("existing idempotency key", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		idempotencyKey := "existing-key"

		expectedResponse := dto.Response{
			Success: true,
			Data: map[string]string{
				"message": "payment already processed",
			},
		}

		responseBody, err := json.Marshal(expectedResponse)
		if err != nil {
			t.Fatalf("failed to marshal response: %v", err)
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

		router := gin.New()
		router.POST(
			"/payments/:merchant_id",
			testController.CreatePayment(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/payments/"+merchant.ID.String(),
			nil,
		)
		req.Header.Set("Idempotency-Key", idempotencyKey)

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusOK,
				w.Code,
			)
		}

		var actualResponse dto.Response

		if err := json.Unmarshal(w.Body.Bytes(), &actualResponse); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !actualResponse.Success {
			t.Fatal("expected successful response")
		}
	})

	t.Run("malformed existing idempotency response", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		responseBody := []byte(`{"success":"not-a-bool"}`)
		idempotency := &models.IdempotencyKey{
			ID:             uuid.New(),
			MerchantID:     merchant.ID,
			IdempotencyKey: "malformed-key",
			ResponseBody:   responseBody,
		}

		if err := db.Create(idempotency).Error; err != nil {
			t.Fatalf("failed to create idempotency key: %v", err)
		}

		router := gin.New()
		router.POST(
			"/payments/:merchant_id",
			testController.CreatePayment(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/payments/"+merchant.ID.String(),
			nil,
		)
		req.Header.Set("Idempotency-Key", "malformed-key")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				w.Code,
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
			"/payments/:merchant_id",
			testController.CreatePayment(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/payments/"+merchant.ID.String(),
			bytes.NewBufferString(`{"amount":`),
		)
		req.Header.Set("Idempotency-Key", idempotencyKey)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				w.Code,
			)
		}

		var savedIdempotency models.IdempotencyKey

		if err := db.
			Where(
				"merchant_id = ? AND idempotency_key = ?",
				merchant.ID,
				idempotencyKey,
			).
			First(&savedIdempotency).Error; err != nil {
			t.Fatalf(
				"expected idempotency response to be saved: %v",
				err,
			)
		}

		var savedResponse dto.Response

		if err := json.Unmarshal(
			savedIdempotency.ResponseBody,
			&savedResponse,
		); err != nil {
			t.Fatalf(
				"failed to decode saved response: %v",
				err,
			)
		}

		if savedResponse.Success {
			t.Fatal("expected saved response to be unsuccessful")
		}

		if savedResponse.Error == nil {
			t.Fatal("expected error information in saved response")
		}

		if savedResponse.Error.Code != "BINDING_ISSUE" {
			t.Fatalf(
				"expected error code BINDING_ISSUE, got %s",
				savedResponse.Error.Code,
			)
		}
	})

	t.Run("successful payment", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		wallet := testutils.GenerateTestWallet(merchant.ID)

		if err := db.Create(wallet).Error; err != nil {
			t.Fatalf("failed to create wallet: %v", err)
		}

		receiverMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(receiverMerchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		receiverWallet := testutils.GenerateTestWallet(receiverMerchant.ID)

		if err := db.Create(receiverWallet).Error; err != nil {
			t.Fatalf("failed to create wallet: %v", err)
		}

		paymentVault := &models.Vault{
			ID:      uuid.New(),
			Type:    constants.PaymentVault,
			Balance: decimal.Zero,
		}

		companyVault := &models.Vault{
			ID:      uuid.New(),
			Type:    constants.CompanyVault,
			Balance: decimal.Zero,
		}

		if err := db.Create(paymentVault).Error; err != nil {
			t.Fatalf(
				"failed to create payment vault: %v",
				err,
			)
		}

		if err := db.Create(companyVault).Error; err != nil {
			t.Fatalf(
				"failed to create company vault: %v",
				err,
			)
		}

		router := gin.New()
		router.POST(
			"/payments/:merchant_id",
			testController.CreatePayment(),
		)

		reqBody := dto.CreatePaymentRequest{
			ReceiverMerchantID: receiverWallet.MerchantID,
			Amount:             decimal.NewFromInt(100),
			Currency:           "INR",
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf(
				"failed to marshal request: %v",
				err,
			)
		}

		idempotencyKey := "successful-payment-key"

		req := httptest.NewRequest(
			http.MethodPost,
			"/payments/"+merchant.ID.String(),
			bytes.NewReader(body),
		)
		req.Header.Set("Idempotency-Key", idempotencyKey)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				w.Code,
				w.Body.String(),
			)
		}

		var response struct {
			Success bool           `json:"success"`
			Data    models.Payment `json:"data"`
			Error   *dto.ErrorInfo `json:"error,omitempty"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode payment response: %v", err)
		}

		payment := response.Data

		if payment.PaymentReference == nil {
			t.Fatal("expected payment reference")
		}

		if payment.Status == nil {
			t.Fatal("expected payment status")
		}

		if *payment.Status != constants.TransactionStatusCompleted {
			t.Fatalf(
				"expected payment status %s, got %s",
				constants.TransactionStatusCompleted,
				*payment.Status,
			)
		}

		var savedIdempotency models.IdempotencyKey

		if err := db.
			Where(
				"merchant_id = ? AND idempotency_key = ?",
				merchant.ID,
				idempotencyKey,
			).
			First(&savedIdempotency).Error; err != nil {
			t.Fatalf(
				"expected idempotency response to be saved: %v",
				err,
			)
		}

		var savedResponse dto.Response

		if err := json.Unmarshal(
			savedIdempotency.ResponseBody,
			&savedResponse,
		); err != nil {
			t.Fatalf(
				"failed to decode saved idempotency response: %v",
				err,
			)
		}

		if !savedResponse.Success {
			t.Fatal("expected saved idempotency response to be successful")
		}
	})
}

func TestTriggerSettlement(t *testing.T) {
	t.Run("successful settlement workflow trigger", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST(
			"/settlement",
			testController.TriggerSettlement(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/settlement",
			nil,
		)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				w.Code,
				w.Body.String(),
			)
		}

		var response struct {
			Success bool           `json:"success"`
			Data    string         `json:"data"`
			Error   *dto.ErrorInfo `json:"error,omitempty"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if !response.Success {
			t.Fatalf("expected successful response, got: %s", w.Body.String())
		}

		if response.Data == "" {
			t.Fatal("expected workflow ID in response")
		}

		expectedPrefix := "settlement-"
		if !strings.HasPrefix(response.Data, expectedPrefix) {
			t.Fatalf(
				"expected workflow ID to start with %q, got %q",
				expectedPrefix,
				response.Data,
			)
		}
	})
}

func TestGetPaymentsByMerchantID(t *testing.T) {
	t.Run("successful request returns merchant payments", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		receiver := testutils.GenerateTestMerchant()
		if err := db.Create(receiver).Error; err != nil {
			t.Fatalf("failed to create receiver merchant: %v", err)
		}

		status := constants.TransactionStatusCompleted

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiver.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "Test payment",
		}

		payment1, err := testRepo.CreatePayment(
			merchant.ID,
			request,
			&status,
		)
		if err != nil {
			t.Fatalf("failed to create payment 1: %v", err)
		}

		payment2, err := testRepo.CreatePayment(
			merchant.ID,
			request,
			&status,
		)
		if err != nil {
			t.Fatalf("failed to create payment 2: %v", err)
		}

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id/payments",
			testController.GetPaymentsByMerchantID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/"+merchant.ID.String()+"/payments",
			nil,
		)

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				w.Code,
				w.Body.String(),
			)
		}

		var response struct {
			Success bool              `json:"success"`
			Data    []*models.Payment `json:"data"`
			Error   *dto.ErrorInfo    `json:"error,omitempty"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if !response.Success {
			t.Fatalf("expected successful response, got: %s", w.Body.String())
		}

		if len(response.Data) != 2 {
			t.Fatalf(
				"expected 2 payments, got %d",
				len(response.Data),
			)
		}

		paymentIDs := map[uuid.UUID]bool{
			payment1.ID: false,
			payment2.ID: false,
		}

		for _, payment := range response.Data {
			if _, exists := paymentIDs[payment.ID]; !exists {
				t.Fatalf(
					"unexpected payment returned: %s",
					payment.ID,
				)
			}

			paymentIDs[payment.ID] = true

			if payment.SenderMerchantID != merchant.ID {
				t.Fatalf(
					"expected sender merchant ID %s, got %s",
					merchant.ID,
					payment.SenderMerchantID,
				)
			}
		}

		for paymentID, found := range paymentIDs {
			if !found {
				t.Fatalf(
					"expected payment %s was not returned",
					paymentID,
				)
			}
		}
	})

	t.Run("merchant with no payments returns empty list", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id/payments",
			testController.GetPaymentsByMerchantID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/"+merchant.ID.String()+"/payments",
			nil,
		)

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				w.Code,
				w.Body.String(),
			)
		}

		var response struct {
			Success bool              `json:"success"`
			Data    []*models.Payment `json:"data"`
			Error   *dto.ErrorInfo    `json:"error,omitempty"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if !response.Success {
			t.Fatalf("expected successful response, got: %s", w.Body.String())
		}

		if len(response.Data) != 0 {
			t.Fatalf(
				"expected empty payment list, got %d payments",
				len(response.Data),
			)
		}
	})

	t.Run("invalid merchant ID returns parsing error", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id/payments",
			testController.GetPaymentsByMerchantID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/invalid-uuid/payments",
			nil,
		)

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusInternalServerError,
				w.Code,
				w.Body.String(),
			)
		}

		var response struct {
			Success bool           `json:"success"`
			Data    interface{}    `json:"data"`
			Error   *dto.ErrorInfo `json:"error,omitempty"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Success {
			t.Fatal("expected unsuccessful response")
		}

		if response.Error == nil {
			t.Fatal("expected error information")
		}

		if response.Error.Code != "PARSING_ERROR" {
			t.Fatalf(
				"expected error code %q, got %q",
				"PARSING_ERROR",
				response.Error.Code,
			)
		}
	})
}

func TestGetPaymentByID(t *testing.T) {
	t.Run("successful request returns payment", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		receiver := testutils.GenerateTestMerchant()
		if err := db.Create(receiver).Error; err != nil {
			t.Fatalf("failed to create receiver merchant: %v", err)
		}

		status := constants.TransactionStatusCompleted

		request := &dto.CreatePaymentRequest{
			ReceiverMerchantID: receiver.ID,
			Amount:             decimal.NewFromInt(1000),
			Currency:           "INR",
			Description:        "Test payment",
		}

		payment, err := testRepo.CreatePayment(
			merchant.ID,
			request,
			&status,
		)
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		router := gin.New()
		router.GET(
			"/payments/:payment_id",
			testController.GetPaymentByID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/payments/"+payment.ID.String(),
			nil,
		)

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				w.Code,
				w.Body.String(),
			)
		}

		var response struct {
			Success bool           `json:"success"`
			Data    models.Payment `json:"data"`
			Error   *dto.ErrorInfo `json:"error,omitempty"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if !response.Success {
			t.Fatalf(
				"expected successful response, got: %s",
				w.Body.String(),
			)
		}

		if response.Error != nil {
			t.Fatalf(
				"expected no error, got: %+v",
				response.Error,
			)
		}

		if response.Data.ID != payment.ID {
			t.Fatalf(
				"expected payment ID %s, got %s",
				payment.ID,
				response.Data.ID,
			)
		}

		if response.Data.SenderMerchantID != merchant.ID {
			t.Fatalf(
				"expected sender merchant ID %s, got %s",
				merchant.ID,
				response.Data.SenderMerchantID,
			)
		}

		if response.Data.ReceiverMerchantID != receiver.ID {
			t.Fatalf(
				"expected receiver merchant ID %s, got %s",
				receiver.ID,
				response.Data.ReceiverMerchantID,
			)
		}

		if !response.Data.Amount.Equal(payment.Amount) {
			t.Fatalf(
				"expected amount %s, got %s",
				payment.Amount,
				response.Data.Amount,
			)
		}

		if response.Data.Currency == nil || *response.Data.Currency != "INR" {
			t.Fatalf(
				"expected currency INR, got %v",
				response.Data.Currency,
			)
		}
	})

	t.Run("invalid payment ID returns parsing error", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.GET(
			"/payments/:payment_id",
			testController.GetPaymentByID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/payments/invalid-uuid",
			nil,
		)

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusInternalServerError,
				w.Code,
				w.Body.String(),
			)
		}

		var response struct {
			Success bool           `json:"success"`
			Data    interface{}    `json:"data"`
			Error   *dto.ErrorInfo `json:"error,omitempty"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Success {
			t.Fatal("expected unsuccessful response")
		}

		if response.Error == nil {
			t.Fatal("expected error information")
		}

		if response.Error.Code != "PARSING_ERROR" {
			t.Fatalf(
				"expected error code %q, got %q",
				"PARSING_ERROR",
				response.Error.Code,
			)
		}
	})
}
