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
)

func TestCreateBankAccount(t *testing.T) {
	t.Run("invalid request", func(t *testing.T) {
		cleanTestDB(t)

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.POST("/bank-accounts", testController.CreateBankAccount())

		req := httptest.NewRequest(
			http.MethodPost,
			"/bank-accounts",
			strings.NewReader(`{"account_name":`),
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

	t.Run("service failure when merchant already has 2 bank accounts", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		bankAccount1 := testutils.GenerateTestBankAccount()
		if err := db.Create(bankAccount1).Error; err != nil {
			t.Fatalf("failed to create bank account 1: %v", err)
		}

		bankAccount2 := testutils.GenerateTestSecondaryBankAccount()
		if err := db.Create(bankAccount2).Error; err != nil {
			t.Fatalf("failed to create bank account 2: %v", err)
		}

		linkedBankAccount1 := &models.LinkedBankAccount{
			ID:            uuid.New(),
			MerchantID:    merchant.ID,
			BankAccountID: bankAccount1.ID,
			Type:          bankAccount1.AccountType,
			Status:        bankAccount1.Status,
		}

		if err := db.Create(linkedBankAccount1).Error; err != nil {
			t.Fatalf(
				"failed to create linked bank account 1: %v",
				err,
			)
		}

		linkedBankAccount2 := &models.LinkedBankAccount{
			ID:            uuid.New(),
			MerchantID:    merchant.ID,
			BankAccountID: bankAccount2.ID,
			Type:          bankAccount2.AccountType,
			Status:        bankAccount2.Status,
		}

		if err := db.Create(linkedBankAccount2).Error; err != nil {
			t.Fatalf(
				"failed to create linked bank account 2: %v",
				err,
			)
		}

		request := &dto.CreateBankAccountRequest{
			MerchantID:  merchant.ID,
			AccountName: "Third Bank Account",
		}

		body, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.POST("/bank-accounts", testController.CreateBankAccount())

		req := httptest.NewRequest(
			http.MethodPost,
			"/bank-accounts",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

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

	t.Run("success", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		request := &dto.CreateBankAccountRequest{
			MerchantID:  merchant.ID,
			AccountName: "Primary Test Account",
		}

		body, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.POST("/bank-accounts", testController.CreateBankAccount())

		req := httptest.NewRequest(
			http.MethodPost,
			"/bank-accounts",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var bankAccount models.BankAccount
		if err := db.Where(
			"account_name = ?",
			request.AccountName,
		).First(&bankAccount).Error; err != nil {
			t.Fatalf(
				"expected bank account to be created: %v",
				err,
			)
		}

		if bankAccount.AccountType != constants.BankAccountTypeSecondary {
			t.Fatalf(
				"expected account type %s, got %s",
				constants.BankAccountTypeSecondary,
				bankAccount.AccountType,
			)
		}

		var linkedBankAccount models.LinkedBankAccount
		if err := db.Where(
			"merchant_id = ? AND bank_account_id = ?",
			merchant.ID,
			bankAccount.ID,
		).First(&linkedBankAccount).Error; err != nil {
			t.Fatalf(
				"expected linked bank account to be created: %v",
				err,
			)
		}

		if linkedBankAccount.MerchantID != merchant.ID {
			t.Fatalf("merchant ID was not linked correctly")
		}

		if linkedBankAccount.BankAccountID != bankAccount.ID {
			t.Fatalf("bank account ID was not linked correctly")
		}
	})
}

func TestUpdateBankAccount(t *testing.T) {
	t.Run("invalid request", func(t *testing.T) {
		cleanTestDB(t)

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.PUT("/bank-accounts", testController.UpdateBankAccount())

		req := httptest.NewRequest(
			http.MethodPut,
			"/bank-accounts",
			strings.NewReader(`{"id":`),
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

	t.Run("id is empty", func(t *testing.T) {
		cleanTestDB(t)

		request := &dto.UpdateBankAccountRequest{
			ID:          uuid.Nil,
			AccountName: "Updated Account",
			AccountType: constants.BankAccountTypeSecondary,
			Status:      constants.MerchantStatusActive,
		}

		body, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.PUT("/bank-accounts", testController.UpdateBankAccount())

		req := httptest.NewRequest(
			http.MethodPut,
			"/bank-accounts",
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

	t.Run("service failure", func(t *testing.T) {
		cleanTestDB(t)

		request := &dto.UpdateBankAccountRequest{
			ID:          uuid.New(),
			AccountName: "Updated Account",
			AccountType: constants.BankAccountTypeSecondary,
			Status:      constants.MerchantStatusActive,
		}

		body, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.PUT("/bank-accounts", testController.UpdateBankAccount())

		req := httptest.NewRequest(
			http.MethodPut,
			"/bank-accounts",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusInternalServerError,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("success", func(t *testing.T) {
		cleanTestDB(t)

		bankAccount := testutils.GenerateTestBankAccount()

		if err := db.Create(bankAccount).Error; err != nil {
			t.Fatalf("failed to create bank account: %v", err)
		}

		merchant := testutils.GenerateTestMerchant()

		if errM := db.Create(merchant).Error; errM != nil {
			t.Fatalf("failed to create merchant: %v", errM)
		}

		linkedBankAccount := &models.LinkedBankAccount{
			ID:            uuid.New(),
			MerchantID:    merchant.ID,
			BankAccountID: bankAccount.ID,
			Type:          bankAccount.AccountType,
			Status:        bankAccount.Status,
		}

		if err := db.Create(linkedBankAccount).Error; err != nil {
			t.Fatalf(
				"failed to create linked bank account: %v",
				err,
			)
		}

		request := &dto.UpdateBankAccountRequest{
			ID:          bankAccount.ID,
			AccountName: "Updated Bank Account",
			AccountType: bankAccount.AccountType,
			Status:      bankAccount.Status,
		}

		body, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.PUT("/bank-accounts", testController.UpdateBankAccount())

		req := httptest.NewRequest(
			http.MethodPut,
			"/bank-accounts",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var updatedBankAccount models.BankAccount

		if err := db.First(
			&updatedBankAccount,
			bankAccount.ID,
		).Error; err != nil {
			t.Fatalf(
				"failed to fetch updated bank account: %v",
				err,
			)
		}

		if updatedBankAccount.AccountName != request.AccountName {
			t.Fatalf(
				"expected account name %s, got %s",
				request.AccountName,
				updatedBankAccount.AccountName,
			)
		}

		var updatedLinkedBankAccount models.LinkedBankAccount

		if err := db.Where(
			"bank_account_id = ?",
			bankAccount.ID,
		).First(&updatedLinkedBankAccount).Error; err != nil {
			t.Fatalf(
				"failed to fetch updated linked bank account: %v",
				err,
			)
		}

		if updatedLinkedBankAccount.Type != request.AccountType {
			t.Fatalf(
				"expected linked bank account type %s, got %s",
				request.AccountType,
				updatedLinkedBankAccount.Type,
			)
		}

		if updatedLinkedBankAccount.Status != request.Status {
			t.Fatalf(
				"expected linked bank account status %s, got %s",
				request.Status,
				updatedLinkedBankAccount.Status,
			)
		}
	})
}

func TestGetBankAccounts(t *testing.T) {
	t.Run("merchant id not found", func(t *testing.T) {
		cleanTestDB(t)

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id/bank-accounts",
			testController.GetBankAccounts(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants//bank-accounts",
			nil,
		)

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

	t.Run("invalid merchant id", func(t *testing.T) {
		cleanTestDB(t)

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id/bank-accounts",
			testController.GetBankAccounts(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/invalid-uuid/bank-accounts",
			nil,
		)

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

	t.Run("no bank account found", func(t *testing.T) {
		cleanTestDB(t)

		merchantID := uuid.New()

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id/bank-accounts",
			testController.GetBankAccounts(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/"+merchantID.String()+"/bank-accounts",
			nil,
		)

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

	t.Run("success", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf(
				"failed to create merchant: %v",
				err,
			)
		}

		bankAccount := testutils.GenerateTestBankAccount()

		if err := db.Create(bankAccount).Error; err != nil {
			t.Fatalf(
				"failed to create bank account: %v",
				err,
			)
		}

		linkedBankAccount := &models.LinkedBankAccount{
			ID:            uuid.New(),
			MerchantID:    merchant.ID,
			BankAccountID: bankAccount.ID,
			Type:          bankAccount.AccountType,
			Status:        bankAccount.Status,
		}

		if err := db.Create(linkedBankAccount).Error; err != nil {
			t.Fatalf(
				"failed to create linked bank account: %v",
				err,
			)
		}

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id/bank-accounts",
			testController.GetBankAccounts(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/"+merchant.ID.String()+"/bank-accounts",
			nil,
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var bankAccounts []models.BankAccount

		if err := db.
			Where("id = ?", bankAccount.ID).
			Find(&bankAccounts).Error; err != nil {
			t.Fatalf(
				"failed to fetch bank account: %v",
				err,
			)
		}

		if len(bankAccounts) != 1 {
			t.Fatalf(
				"expected 1 bank account in database, got %d",
				len(bankAccounts),
			)
		}

		if bankAccounts[0].ID != bankAccount.ID {
			t.Fatalf("returned bank account ID does not match")
		}
	})
}

func TestGetBankAccount(t *testing.T) {
	t.Run("invalid bank account id", func(t *testing.T) {
		cleanTestDB(t)

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.GET(
			"/bank-accounts/:bank_account_id",
			testController.GetBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/bank-accounts/invalid-uuid",
			nil,
		)

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

	t.Run("bank account not found", func(t *testing.T) {
		cleanTestDB(t)

		bankAccountID := uuid.New()

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.GET(
			"/bank-accounts/:bank_account_id",
			testController.GetBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/bank-accounts/"+bankAccountID.String(),
			nil,
		)

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

	t.Run("success", func(t *testing.T) {
		cleanTestDB(t)

		bankAccount := testutils.GenerateTestBankAccount()

		if err := db.Create(bankAccount).Error; err != nil {
			t.Fatalf(
				"failed to create bank account: %v",
				err,
			)
		}

		gin.SetMode(gin.TestMode)

		router := gin.New()
		router.GET(
			"/bank-accounts/:bank_account_id",
			testController.GetBankAccount(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/bank-accounts/"+bankAccount.ID.String(),
			nil,
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var fetchedBankAccount models.BankAccount

		if err := db.First(
			&fetchedBankAccount,
			bankAccount.ID,
		).Error; err != nil {
			t.Fatalf(
				"failed to fetch bank account: %v",
				err,
			)
		}

		if fetchedBankAccount.ID != bankAccount.ID {
			t.Fatalf(
				"expected bank account ID %s, got %s",
				bankAccount.ID,
				fetchedBankAccount.ID,
			)
		}
	})
}
