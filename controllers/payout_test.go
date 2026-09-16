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
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestWalletToBankAccount(t *testing.T) {
	t.Run("invalid merchant ID", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST(
			"/merchants/:merchant_id/payout",
			testController.WalletToBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/invalid-uuid/payout",
			strings.NewReader(`{}`),
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

		var response dto.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Success {
			t.Fatal("expected unsuccessful response")
		}

		if response.Error == nil {
			t.Fatal("expected error information")
		}

		if response.Error.Code != "MERCHANT_ID_PARSING_ERROR" {
			t.Fatalf(
				"expected error code %q, got %q",
				"MERCHANT_ID_PARSING_ERROR",
				response.Error.Code,
			)
		}
	})

	t.Run("missing idempotency key", func(t *testing.T) {
		cleanTestDB(t)

		merchantID := uuid.New()

		router := gin.New()
		router.POST(
			"/merchants/:merchant_id/payout",
			testController.WalletToBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/"+merchantID.String()+"/payout",
			strings.NewReader(`{}`),
		)

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusBadRequest,
				w.Code,
				w.Body.String(),
			)
		}

		var response dto.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Success {
			t.Fatal("expected unsuccessful response")
		}

		if response.Error == nil {
			t.Fatal("expected error information")
		}

		if response.Error.Code != "IDEMPOTENCY_KEY_NOT_FOUND" {
			t.Fatalf(
				"expected error code %q, got %q",
				"IDEMPOTENCY_KEY_NOT_FOUND",
				response.Error.Code,
			)
		}
	})

	t.Run("existing idempotency key returns stored response", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		var savedMerchant models.Merchant
		if err := db.First(&savedMerchant, "id = ?", merchant.ID).Error; err != nil {
			t.Fatalf("merchant was not persisted: %v", err)
		}

		merchantID := merchant.ID
		idempotencyKey := "wallet-bank-test-key"

		storedResponse := dto.Response{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "PAYOUT_FAILED",
				Message: "stored payout failure",
			},
		}

		responseBody, err := json.Marshal(storedResponse)
		if err != nil {
			t.Fatalf("failed to marshal response: %v", err)
		}

		if _, err := testServices.CreateIdempotencyKey(
			merchant.ID,
			idempotencyKey,
			responseBody,
		); err != nil {
			t.Fatalf("failed to create idempotency key: %v", err)
		}

		router := gin.New()
		router.POST(
			"/merchants/:merchant_id/payout",
			testController.WalletToBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/"+merchantID.String()+"/payout",
			strings.NewReader(`{}`),
		)

		req.Header.Set("Idempotency-Key", idempotencyKey)

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

		var response dto.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Success {
			t.Fatal("expected stored failure response")
		}

		if response.Error == nil {
			t.Fatal("expected error information")
		}

		if response.Error.Code != "PAYOUT_FAILED" {
			t.Fatalf(
				"expected error code %q, got %q",
				"PAYOUT_FAILED",
				response.Error.Code,
			)
		}

		if response.Error.Message != "stored payout failure" {
			t.Fatalf(
				"expected error message %q, got %q",
				"stored payout failure",
				response.Error.Message,
			)
		}
	})

	t.Run("invalid stored idempotency response", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		var savedMerchant models.Merchant
		if err := db.First(&savedMerchant, "id = ?", merchant.ID).Error; err != nil {
			t.Fatalf("merchant was not persisted: %v", err)
		}

		merchantID := merchant.ID
		idempotencyKey := "wallet-bank-invalid-response"

		responseBody := []byte(`{"success":"not-a-bool"}`)

		if _, err := testServices.CreateIdempotencyKey(
			merchantID,
			idempotencyKey,
			responseBody,
		); err != nil {
			t.Fatalf("failed to create idempotency key: %v", err)
		}

		router := gin.New()
		router.POST(
			"/merchants/:merchant_id/payout",
			testController.WalletToBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/"+merchantID.String()+"/payout",
			strings.NewReader(`{}`),
		)

		req.Header.Set("Idempotency-Key", idempotencyKey)

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

		var response dto.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Success {
			t.Fatal("expected unsuccessful response")
		}

		if response.Error == nil {
			t.Fatal("expected error information")
		}

		if response.Error.Code != "FAILED_UNMARSHALLING" {
			t.Fatalf(
				"expected error code %q, got %q",
				"FAILED_UNMARSHALLING",
				response.Error.Code,
			)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		var savedMerchant models.Merchant
		if err := db.First(&savedMerchant, "id = ?", merchant.ID).Error; err != nil {
			t.Fatalf("merchant was not persisted: %v", err)
		}

		merchantID := merchant.ID
		idempotencyKey := "wallet-bank-binding-error"

		router := gin.New()
		router.POST(
			"/merchants/:merchant_id/payout",
			testController.WalletToBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/"+merchantID.String()+"/payout",
			strings.NewReader(`{"invalid-json"`),
		)

		req.Header.Set("Idempotency-Key", idempotencyKey)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusBadRequest,
				w.Code,
				w.Body.String(),
			)
		}

		var response dto.Response
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Success {
			t.Fatal("expected unsuccessful response")
		}

		if response.Error == nil {
			t.Fatal("expected error information")
		}

		if response.Error.Code != "BINDING_ISSUE" {
			t.Fatalf(
				"expected error code %q, got %q",
				"BINDING_ISSUE",
				response.Error.Code,
			)
		}

		stored, err := testServices.GetIdempotencyByKeyAndMerchantId(
			merchantID,
			idempotencyKey,
		)

		if err != nil {
			t.Fatalf("failed to retrieve idempotency key: %v", err)
		}

		if stored == nil {
			t.Fatal("expected idempotency response to be saved")
		}
	})

	t.Run("successful wallet to bank payout", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		var savedMerchant models.Merchant
		if err := db.First(&savedMerchant, "id = ?", merchant.ID).Error; err != nil {
			t.Fatalf("merchant was not persisted: %v", err)
		}

		wallet := testutils.GenerateTestWallet(merchant.ID)
		wallet.AvailableBalance = decimal.NewFromInt(20000)
		wallet.ReservedBalance = decimal.Zero

		if err := db.Create(wallet).Error; err != nil {
			t.Fatalf("failed to create wallet: %v", err)
		}

		destinationBankAccount := testutils.GenerateTestBankAccount()
		destinationBankAccount.Balance = decimal.Zero

		if err := db.Create(destinationBankAccount).Error; err != nil {
			t.Fatalf("failed to create destination bank account: %v", err)
		}

		payoutVaultType := constants.PayoutVault
		payoutVault := &models.Vault{
			ID:      uuid.New(),
			Type:    payoutVaultType,
			Balance: decimal.NewFromInt(1000000),
		}

		if err := db.Create(payoutVault).Error; err != nil {
			t.Fatalf("failed to create payout vault: %v", err)
		}

		companyVaultType := constants.CompanyVault
		companyVault := &models.Vault{
			ID:      uuid.New(),
			Type:    companyVaultType,
			Balance: decimal.Zero,
		}

		if err := db.Create(companyVault).Error; err != nil {
			t.Fatalf("failed to create company vault: %v", err)
		}

		currency := "INR"
		externalReference := "wallet-bank-success"
		description := "test wallet to bank payout"

		requestBody := dto.CreateWalletToBankRequest{
			SenderWalletID:           wallet.ID,
			DestinationBankAccountID: destinationBankAccount.ID,
			ExternalReference:        &externalReference,
			Amount:                   decimal.NewFromInt(10000),
			Currency:                 currency,
			Description:              &description,
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()

		router.POST(
			"/:merchant_id/payout/wallet-to-bank",
			testController.WalletToBankAccount(),
		)

		recorder := httptest.NewRecorder()

		req := httptest.NewRequest(
			http.MethodPost,
			"/"+merchant.ID.String()+"/payout/wallet-to-bank",
			bytes.NewReader(body),
		)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "wallet-bank-success-key")

		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf(
				"expected status 200, got %d: %s",
				recorder.Code,
				recorder.Body.String(),
			)
		}

		var response struct {
			Success bool           `json:"success"`
			Data    models.Payout  `json:"data"`
			Error   *dto.ErrorInfo `json:"error,omitempty"`
		}

		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Fatalf(
				"expected successful payout, got error: %+v",
				response.Error,
			)
		}

		if response.Data.ID == uuid.Nil {
			t.Fatal("expected payout ID in response")
		}

		if response.Data.PayoutReference == "" {
			t.Fatal("expected payout reference in response")
		}

		if response.Data.Status == "" {
			t.Fatal("expected payout status in response")
		}

		if response.Data.Status != constants.TransactionStatusCompleted {
			t.Fatalf(
				"expected payout status %q, got %q",
				constants.TransactionStatusCompleted,
				response.Data.Status,
			)
		}

		var updatedWallet models.Wallets
		if err := db.First(&updatedWallet, "id = ?", wallet.ID).Error; err != nil {
			t.Fatalf("failed to fetch updated wallet: %v", err)
		}

		expectedWalletBalance := decimal.NewFromInt(9900)

		if !updatedWallet.AvailableBalance.Equal(expectedWalletBalance) {
			t.Fatalf(
				"expected wallet available balance %s, got %s",
				expectedWalletBalance,
				updatedWallet.AvailableBalance,
			)
		}

		var updatedBankAccount models.BankAccount
		if err := db.First(
			&updatedBankAccount,
			"id = ?",
			destinationBankAccount.ID,
		).Error; err != nil {
			t.Fatalf("failed to fetch updated bank account: %v", err)
		}

		expectedBankBalance := decimal.NewFromInt(10000)

		if !updatedBankAccount.Balance.Equal(expectedBankBalance) {
			t.Fatalf(
				"expected bank balance %s, got %s",
				expectedBankBalance,
				updatedBankAccount.Balance,
			)
		}

		var updatedPayoutVault models.Vault
		if err := db.First(
			&updatedPayoutVault,
			"id = ?",
			payoutVault.ID,
		).Error; err != nil {
			t.Fatalf("failed to fetch updated payout vault: %v", err)
		}

		var updatedCompanyVault models.Vault
		if err := db.First(
			&updatedCompanyVault,
			"id = ?",
			companyVault.ID,
		).Error; err != nil {
			t.Fatalf("failed to fetch updated company vault: %v", err)
		}

		expectedFee := decimal.NewFromInt(100)

		if !updatedCompanyVault.Balance.Equal(expectedFee) {
			t.Fatalf(
				"expected company vault balance %s, got %s",
				expectedFee,
				updatedCompanyVault.Balance,
			)
		}
	})
}

func TestBankToBankAccount(t *testing.T) {
	t.Run("invalid merchant id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST(
			"/:merchant_id/payout/bank-to-bank",
			testController.BankToBankAccount(),
		)

		requestBody := dto.CreateBankToBankRequest{
			SourceBankAccountID:      uuid.New(),
			DestinationBankAccountID: uuid.New(),
			Amount:                   decimal.NewFromInt(1000),
			Currency:                 "INR",
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest(
			http.MethodPost,
			"/invalid-merchant-id/payout/bank-to-bank",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "test-idempotency-key")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
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
			"/:merchant_id/payout/bank-to-bank",
			testController.BankToBankAccount(),
		)

		requestBody := dto.CreateBankToBankRequest{
			SourceBankAccountID:      uuid.New(),
			DestinationBankAccountID: uuid.New(),
			Amount:                   decimal.NewFromInt(1000),
			Currency:                 "INR",
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest(
			http.MethodPost,
			"/"+merchant.ID.String()+"/payout/bank-to-bank",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("existing idempotency key", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		idempotencyKey := "existing-bank-to-bank-key"

		response := dto.Response{
			Success: true,
			Data: map[string]string{
				"message": "existing response",
			},
		}

		responseBody, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("failed to marshal response: %v", err)
		}

		_, err = testServices.CreateIdempotencyKey(
			merchant.ID,
			idempotencyKey,
			responseBody,
		)
		if err != nil {
			t.Fatalf("failed to create idempotency key: %v", err)
		}

		router := gin.New()
		router.POST(
			"/:merchant_id/payout/bank-to-bank",
			testController.BankToBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/"+merchant.ID.String()+"/payout/bank-to-bank",
			bytes.NewReader([]byte(`{}`)),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusOK,
				rec.Code,
			)
		}

		var actualResponse dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &actualResponse); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !actualResponse.Success {
			t.Fatal("expected success to be true")
		}
	})

	t.Run("invalid stored idempotency response", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		idempotencyKey := "invalid-bank-to-bank-response"

		invalidResponse := []byte(`{"success":"not-a-bool"}`)

		_, err := testServices.CreateIdempotencyKey(
			merchant.ID,
			idempotencyKey,
			invalidResponse,
		)
		if err != nil {
			t.Fatalf("failed to create idempotency key: %v", err)
		}

		router := gin.New()
		router.POST(
			"/:merchant_id/payout/bank-to-bank",
			testController.BankToBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/"+merchant.ID.String()+"/payout/bank-to-bank",
			bytes.NewReader([]byte(`{}`)),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.POST(
			"/:merchant_id/payout/bank-to-bank",
			testController.BankToBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/"+merchant.ID.String()+"/payout/bank-to-bank",
			bytes.NewReader([]byte(`{"amount":`)),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "invalid-body-key")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("successful bank to bank payout", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		sourceBankAccount := testutils.GenerateTestBankAccount()
		sourceBankAccount.Balance = decimal.NewFromInt(20000)

		if err := db.Create(sourceBankAccount).Error; err != nil {
			t.Fatalf("failed to create source bank account: %v", err)
		}

		linkedSourceBankAccountPayload := &models.LinkedBankAccount{
			ID:            uuid.New(),
			MerchantID:    merchant.ID,
			BankAccountID: sourceBankAccount.ID,
			Type:          sourceBankAccount.AccountType,
			Status:        merchant.Status,
		}

		if err := db.Create(linkedSourceBankAccountPayload).Error; err != nil {
			t.Fatalf("failed to create link bank account: %v", err)
		}

		destinationBankAccount := testutils.GenerateTestSecondaryBankAccount()
		destinationBankAccount.Balance = decimal.Zero

		if err := db.Create(destinationBankAccount).Error; err != nil {
			t.Fatalf("failed to create destination bank account: %v", err)
		}

		requestBody := dto.CreateBankToBankRequest{
			SourceBankAccountID:      sourceBankAccount.ID,
			DestinationBankAccountID: destinationBankAccount.ID,
			Amount:                   decimal.NewFromInt(10000),
			Currency:                 "INR",
			ExternalReference:        utils.GenerateCustomerReference(),
			Description:              func() *string { s := "Test bank to bank payout"; return &s }(),
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		router := gin.New()
		router.POST(
			"/:merchant_id/payout/bank-to-bank",
			testController.BankToBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/"+merchant.ID.String()+"/payout/bank-to-bank",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "successful-bank-to-bank-key")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, response: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var response struct {
			Success bool           `json:"success"`
			Data    models.Payout  `json:"data"`
			Error   *dto.ErrorInfo `json:"error,omitempty"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Fatalf("expected successful response, got error: %+v", response.Error)
		}

		if response.Data.ID == uuid.Nil {
			t.Fatal("expected payout ID to be set")
		}

		if response.Data.PayoutReference == "" {
			t.Fatal("expected payout reference to be set")
		}

		if response.Data.Status == "" ||
			response.Data.Status != constants.TransactionStatusCompleted {
			t.Fatalf(
				"expected payout status %q, got %v",
				constants.TransactionStatusCompleted,
				response.Data.Status,
			)
		}

		var updatedSource models.BankAccount
		if err := db.
			Where("id = ?", sourceBankAccount.ID).
			First(&updatedSource).Error; err != nil {
			t.Fatalf("failed to reload source bank account: %v", err)
		}

		if !updatedSource.Balance.Equal(decimal.NewFromInt(10000)) {
			t.Fatalf(
				"expected source balance 9900, got %s",
				updatedSource.Balance,
			)
		}

		var updatedDestination models.BankAccount
		if err := db.
			Where("id = ?", destinationBankAccount.ID).
			First(&updatedDestination).Error; err != nil {
			t.Fatalf("failed to reload destination bank account: %v", err)
		}

		if !updatedDestination.Balance.Equal(decimal.NewFromInt(10000)) {
			t.Fatalf(
				"expected destination balance 10000, got %s",
				updatedDestination.Balance,
			)
		}
	})
}

func TestGetPayoutsByMerchantID(t *testing.T) {
	t.Run("invalid merchant id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.GET(
			"/:merchant_id/payouts",
			testController.GetPayoutsByMerchantID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/invalid-merchant-id/payouts",
			nil,
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d, response: %s",
				http.StatusBadRequest,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("merchant has no payouts", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.GET(
			"/:merchant_id/payouts",
			testController.GetPayoutsByMerchantID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/"+merchant.ID.String()+"/payouts",
			nil,
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, response: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var response struct {
			Success bool            `json:"success"`
			Data    []models.Payout `json:"data"`
			Error   *dto.ErrorInfo  `json:"error,omitempty"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Fatalf(
				"expected successful response, got error: %+v",
				response.Error,
			)
		}

		if len(response.Data) != 0 {
			t.Fatalf(
				"expected 0 payouts, got %d",
				len(response.Data),
			)
		}
	})

	t.Run("successful payout retrieval", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		sourceWallet := testutils.GenerateTestWallet(merchant.ID)
		if errW := db.Create(sourceWallet).Error; errW != nil {
			t.Fatalf("failed to create wallet: %v", errW)
		}

		destinationBankAccount := testutils.GenerateTestBankAccount()
		if errB := db.Create(destinationBankAccount).Error; errB != nil {
			t.Fatalf("failed to create bank account: %v", errB)
		}

		payout1Reference := "payout-test-reference-1"
		payout2Reference := "payout-test-reference-2"

		currency := "INR"
		status := constants.TransactionStatusCompleted

		payout1 := models.Payout{
			ID:                       uuid.New(),
			MerchantID:               merchant.ID,
			SourceWalletID:           &sourceWallet.ID,
			SourceBankAccountID:      nil,
			DestinationBankAccountID: destinationBankAccount.ID,
			PayoutReference:          payout1Reference,
			Amount:                   decimal.NewFromInt(100),
			Currency:                 currency,
			Status:                   status,
		}

		payout2 := models.Payout{
			ID:                       uuid.New(),
			MerchantID:               merchant.ID,
			SourceWalletID:           &sourceWallet.ID,
			SourceBankAccountID:      nil,
			DestinationBankAccountID: destinationBankAccount.ID,
			PayoutReference:          payout2Reference,
			Amount:                   decimal.NewFromInt(50),
			Currency:                 currency,
			Status:                   status,
		}

		if err := db.Create(&payout1).Error; err != nil {
			t.Fatalf("failed to create payout1: %v", err)
		}

		if err := db.Create(&payout2).Error; err != nil {
			t.Fatalf("failed to create payout2: %v", err)
		}

		// Create a payout belonging to another merchant.
		otherMerchant := testutils.GenerateTestMerchant()

		if err := db.Create(otherMerchant).Error; err != nil {
			t.Fatalf("failed to create other merchant: %v", err)
		}

		otherSourceWallet := testutils.GenerateTestWallet(otherMerchant.ID)
		if errW := db.Create(otherSourceWallet).Error; errW != nil {
			t.Fatalf("failed to create wallet: %v", errW)
		}

		otherDestinationBankAccount := testutils.GenerateTestBankAccount()
		if errB := db.Create(otherDestinationBankAccount).Error; errB != nil {
			t.Fatalf("failed to create bank account: %v", errB)
		}

		otherPayoutReference := "payout-test-reference-other"

		otherPayout := models.Payout{
			ID:                       uuid.New(),
			MerchantID:               otherMerchant.ID,
			SourceWalletID:           &otherSourceWallet.ID,
			SourceBankAccountID:      nil,
			DestinationBankAccountID: otherDestinationBankAccount.ID,
			PayoutReference:          otherPayoutReference,
			Amount:                   decimal.NewFromInt(20000),
			Currency:                 currency,
			Status:                   status,
		}

		if err := db.Create(&otherPayout).Error; err != nil {
			t.Fatalf("failed to create other payout: %v", err)
		}

		router := gin.New()
		router.GET(
			"/:merchant_id/payouts",
			testController.GetPayoutsByMerchantID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/"+merchant.ID.String()+"/payouts",
			nil,
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, response: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var response struct {
			Success bool            `json:"success"`
			Data    []models.Payout `json:"data"`
			Error   *dto.ErrorInfo  `json:"error,omitempty"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Fatalf(
				"expected successful response, got error: %+v",
				response.Error,
			)
		}

		if len(response.Data) != 2 {
			t.Fatalf(
				"expected 2 payouts, got %d",
				len(response.Data),
			)
		}

		for _, payout := range response.Data {
			if payout.MerchantID != merchant.ID {
				t.Fatalf(
					"expected payout merchant ID %s, got %s",
					merchant.ID,
					payout.MerchantID,
				)
			}
		}
	})
}

func TestGetPayoutByID(t *testing.T) {
	t.Run("invalid payout id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.GET(
			"/payouts/:payout_id",
			testController.GetPayoutByID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/payouts/invalid-payout-id",
			nil,
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d, response: %s",
				http.StatusBadRequest,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("successful payout retrieval", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		sourceWallet := testutils.GenerateTestWallet(merchant.ID)
		if errW := db.Create(sourceWallet).Error; errW != nil {
			t.Fatalf("failed to create wallet: %v", errW)
		}

		destinationBankAccount := testutils.GenerateTestBankAccount()
		if errB := db.Create(destinationBankAccount).Error; errB != nil {
			t.Fatalf("failed to create bank account: %v", errB)
		}

		payoutReference := "payout-test-reference-single"
		currency := "INR"
		status := constants.TransactionStatusCompleted

		payout := models.Payout{
			ID:                       uuid.New(),
			MerchantID:               merchant.ID,
			SourceWalletID:           &sourceWallet.ID,
			SourceBankAccountID:      nil,
			DestinationBankAccountID: destinationBankAccount.ID,
			PayoutReference:          payoutReference,
			Amount:                   decimal.NewFromInt(10000),
			Currency:                 currency,
			Status:                   status,
		}

		if err := db.Create(&payout).Error; err != nil {
			t.Fatalf("failed to create payout: %v", err)
		}

		router := gin.New()
		router.GET(
			"/payouts/:payout_id",
			testController.GetPayoutByID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/payouts/"+payout.ID.String(),
			nil,
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, response: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var response struct {
			Success bool           `json:"success"`
			Data    models.Payout  `json:"data"`
			Error   *dto.ErrorInfo `json:"error,omitempty"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Fatalf(
				"expected successful response, got error: %+v",
				response.Error,
			)
		}

		if response.Data.ID != payout.ID {
			t.Fatalf(
				"expected payout ID %s, got %s",
				payout.ID,
				response.Data.ID,
			)
		}

		if response.Data.MerchantID != merchant.ID {
			t.Fatalf(
				"expected merchant ID %s, got %s",
				merchant.ID,
				response.Data.MerchantID,
			)
		}

		if response.Data.Amount.Cmp(payout.Amount) != 0 {
			t.Fatalf(
				"expected amount %s, got %s",
				payout.Amount,
				response.Data.Amount,
			)
		}

		if response.Data.PayoutReference == "" ||
			response.Data.PayoutReference != payoutReference {
			t.Fatalf(
				"expected payout reference %q, got %v",
				payoutReference,
				response.Data.PayoutReference,
			)
		}
	})
}
