package controllers

import (
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

func TestGetWallet(t *testing.T) {
	cleanTestDB(t)

	merchant := testutils.GenerateTestMerchant()
	err := db.Create(merchant).Error
	if err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	wallet := testutils.GenerateTestWallet(merchant.ID)

	err = db.Create(wallet).Error
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	t.Run("invalid merchant id", func(t *testing.T) {
		router := gin.New()
		router.GET("/:merchant_id/wallet", testController.GetWallet())

		req := httptest.NewRequest(
			http.MethodGet,
			"/invalid-uuid/wallet",
			nil,
		)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})

	t.Run("success", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		errM := db.Create(merchant).Error
		if errM != nil {
			t.Fatalf("failed to create merchant: %v", errM)
		}

		wallet := testutils.GenerateTestWallet(merchant.ID)

		errW := db.Create(wallet).Error
		if errW != nil {
			t.Fatalf("failed to create wallet: %v", errW)
		}

		router := gin.New()
		router.GET("/:merchant_id/wallet", testController.GetWallet())

		req := httptest.NewRequest(
			http.MethodGet,
			"/"+merchant.ID.String()+"/wallet",
			nil,
		)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusOK,
				rec.Code,
			)
		}

		var response struct {
			Success bool           `json:"success"`
			Data    models.Wallets `json:"data"`
			Error   *dto.ErrorInfo `json:"error,omitempty"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode payment response: %v", err)
		}

		if response.Data.ID != wallet.ID {
			t.Errorf(
				"expected wallet ID %s, got %s",
				wallet.ID,
				response.Data.ID,
			)
		}

		if response.Data.MerchantID != merchant.ID {
			t.Errorf(
				"expected merchant ID %s, got %s",
				merchant.ID,
				response.Data.MerchantID,
			)
		}
	})
}

func TestGetWalletTransaction(t *testing.T) {
	cleanTestDB(t)

	merchant := testutils.GenerateTestMerchant()
	err := db.Create(merchant).Error
	if err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	wallet := testutils.GenerateTestWallet(merchant.ID)
	errW := db.Create(wallet).Error
	if errW != nil {
		t.Fatalf("failed to create wallet: %v", errW)
	}

	t.Run("merchant id not provided", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		req := httptest.NewRequest(
			http.MethodGet,
			"/wallet/transactions/transaction-id",
			nil,
		)
		c.Request = req

		c.Params = gin.Params{
			{
				Key:   "merchant_id",
				Value: "",
			},
			{
				Key:   "transaction_id",
				Value: uuid.New().String(),
			},
		}

		testController.GetWalletTransaction()(c)

		if rec.Code != http.StatusNotFound {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusNotFound,
				rec.Code,
			)
		}
	})

	t.Run("transaction id not provided", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		req := httptest.NewRequest(
			http.MethodGet,
			"/wallet/transactions",
			nil,
		)
		c.Request = req

		c.Params = gin.Params{
			{
				Key:   "merchant_id",
				Value: merchant.ID.String(),
			},
			{
				Key:   "transaction_id",
				Value: "",
			},
		}

		testController.GetWalletTransaction()(c)

		if rec.Code != http.StatusNotFound {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusNotFound,
				rec.Code,
			)
		}
	})

	t.Run("invalid merchant id", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		err := db.Create(merchant).Error
		if err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		wallet := testutils.GenerateTestWallet(merchant.ID)
		errW := db.Create(wallet).Error
		if errW != nil {
			t.Fatalf("failed to create wallet: %v", errW)
		}
		router := gin.New()
		router.GET(
			"/:merchant_id/wallet/transactions/:transaction_id",
			testController.GetWalletTransaction(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/invalid-uuid/wallet/transactions/"+uuid.New().String(),
			nil,
		)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})

	t.Run("invalid transaction id", func(t *testing.T) {
		router := gin.New()
		router.GET(
			"/:merchant_id/wallet/transactions/:transaction_id",
			testController.GetWalletTransaction(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/"+merchant.ID.String()+"/wallet/transactions/invalid-uuid",
			nil,
		)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})

	t.Run("transaction not found", func(t *testing.T) {
		router := gin.New()
		router.GET(
			"/:merchant_id/wallet/transactions/:transaction_id",
			testController.GetWalletTransaction(),
		)

		randomTransactionID := uuid.New()

		req := httptest.NewRequest(
			http.MethodGet,
			"/"+merchant.ID.String()+"/wallet/transactions/"+randomTransactionID.String(),
			nil,
		)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})

	t.Run("success", func(t *testing.T) {
		ledgerTransaction := models.LedgerTransaction{
			ID:               uuid.New(),
			TransactionRef:   utils.GenerateLedgerReference(),
			Type:             constants.LedgerEntryTypeDebit,
			ReferenceID:      "transaction-test-reference",
			Status:           constants.TransactionStatusCompleted,
			SettlementStatus: constants.LedgerSettlementSettled,
		}

		if err := db.Create(&ledgerTransaction).Error; err != nil {
			t.Fatalf("failed to create ledger transaction: %v", err)
		}

		ledgerEntry := models.LedgerEntry{
			ID:                  uuid.New(),
			LedgerTransactionID: ledgerTransaction.ID,
			AccountID:           wallet.ID,
			Amount:              decimal.NewFromInt(50),
			EntryType:           constants.LedgerEntryTypeDebit,
			AccountType:         constants.LedgerAccountTypeWallet,
			Currency:            "INR",
		}

		if err := db.Create(&ledgerEntry).Error; err != nil {
			t.Fatalf("failed to create ledger entry: %v", err)
		}

		router := gin.New()
		router.GET(
			"/:merchant_id/wallet/transactions/:transaction_id",
			testController.GetWalletTransaction(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/"+merchant.ID.String()+
				"/wallet/transactions/"+
				ledgerTransaction.ID.String(),
			nil,
		)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var response struct {
			Success *bool                    `json:"success"`
			Data    dto.LedgerTransactionRow `json:"data"`
		}

		err := json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success == nil || !*response.Success {
			t.Errorf("expected success to be true")
		}

		if response.Data.LedgerTransactionID != ledgerTransaction.ID {
			t.Errorf(
				"expected transaction ID %s, got %s",
				ledgerTransaction.ID,
				response.Data.LedgerTransactionID,
			)
		}

		if response.Data.Amount.Cmp(decimal.NewFromInt(50)) != 0 {
			t.Errorf(
				"expected amount 50, got %s",
				response.Data.Amount,
			)
		}

		if response.Data.Currency == "" || response.Data.Currency != "INR" {
			t.Errorf("expected currency INR")
		}
	})
}

func TestTopUpWallet(t *testing.T) {
	cleanTestDB(t)

	merchant := testutils.GenerateTestMerchant()
	err := db.Create(merchant).Error
	if err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	wallet := testutils.GenerateTestWallet(merchant.ID)
	errW := db.Create(wallet).Error
	if errW != nil {
		t.Fatalf("failed to create wallet: %v", errW)
	}

	t.Run("invalid merchant id", func(t *testing.T) {
		router := gin.New()
		router.POST("/:merchant_id/wallet/topup", testController.TopUpWallet())

		req := httptest.NewRequest(
			http.MethodPost,
			"/invalid-uuid/wallet/topup",
			strings.NewReader(`{"amount":"100"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		router := gin.New()
		router.POST("/:merchant_id/wallet/topup", testController.TopUpWallet())

		req := httptest.NewRequest(
			http.MethodPost,
			"/"+merchant.ID.String()+"/wallet/topup",
			strings.NewReader(`invalid-json`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("wallet not found", func(t *testing.T) {
		router := gin.New()
		router.POST("/:merchant_id/wallet/topup", testController.TopUpWallet())

		randomMerchantID := uuid.New()

		req := httptest.NewRequest(
			http.MethodPost,
			"/"+randomMerchantID.String()+"/wallet/topup",
			strings.NewReader(`{"amount":"100"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})

	t.Run("topup amount is zero", func(t *testing.T) {
		router := gin.New()
		router.POST("/:merchant_id/wallet/topup", testController.TopUpWallet())

		req := httptest.NewRequest(
			http.MethodPost,
			"/"+merchant.ID.String()+"/wallet/topup",
			strings.NewReader(`{"amount":"0"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})

	t.Run("success", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		sourceWallet := testutils.GenerateTestWallet(merchant.ID)
		if errW := db.Create(sourceWallet).Error; errW != nil {
			t.Fatalf("failed to create wallet: %v", errW)
		}

		bankAccount := testutils.GenerateTestBankAccount()
		if errB := db.Create(bankAccount).Error; errB != nil {
			t.Fatalf("failed to create bank account: %v", errB)
		}

		linkedBankAccount := &models.LinkedBankAccount{
			BankAccountID: bankAccount.ID,
			MerchantID:    merchant.ID,
			Type:          constants.BankAccountTypePrimary,
			Status:        merchant.Status,
		}

		if errL := db.Create(linkedBankAccount).Error; errL != nil {
			t.Fatalf("failed to create linked bank account: %v", errL)
		}

		initialWalletBalance := sourceWallet.AvailableBalance
		topupAmount := decimal.NewFromInt(100)

		router := gin.New()
		router.POST("/:merchant_id/wallet/topup", testController.TopUpWallet())

		req := httptest.NewRequest(
			http.MethodPost,
			"/"+merchant.ID.String()+"/wallet/topup",
			strings.NewReader(`{"amount":"100"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var response struct {
			Success *bool          `json:"success"`
			Data    models.Wallets `json:"data"`
		}

		err := json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success == nil || !*response.Success {
			t.Errorf("expected success to be true")
		}

		expectedWalletBalance := initialWalletBalance.Add(topupAmount)

		if response.Data.AvailableBalance.Cmp(expectedWalletBalance) != 0 {
			t.Errorf(
				"expected wallet balance %s, got %s",
				expectedWalletBalance,
				response.Data.AvailableBalance,
			)
		}
	})
}
