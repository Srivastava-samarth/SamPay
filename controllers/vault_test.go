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

func TestCreateVault(t *testing.T) {
	cleanTestDB(t)

	t.Run("invalid request body", func(t *testing.T) {
		router := gin.New()
		router.POST("/vault", testController.CreateVault())

		req := httptest.NewRequest(
			http.MethodPost,
			"/vault",
			strings.NewReader(`{"type":`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		requestBody := dto.CreateVaultRequest{
			Type:    "payment",
			Balance: decimal.NewFromInt(1000),
			Status:  "invalid-status",
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.POST("/vault", testController.CreateVault())

		req := httptest.NewRequest(
			http.MethodPost,
			"/vault",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		requestBody := dto.CreateVaultRequest{
			Type:    "invalid-type",
			Balance: decimal.NewFromInt(1000),
			Status:  constants.VaultStatusActive,
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.POST("/vault", testController.CreateVault())

		req := httptest.NewRequest(
			http.MethodPost,
			"/vault",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("zero balance", func(t *testing.T) {
		requestBody := dto.CreateVaultRequest{
			Type:    constants.PaymentVault,
			Balance: decimal.Zero,
			Status:  constants.VaultStatusActive,
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.POST("/vault", testController.CreateVault())

		req := httptest.NewRequest(
			http.MethodPost,
			"/vault",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		requestBody := dto.CreateVaultRequest{
			Type:    constants.PaymentVault,
			Balance: decimal.NewFromInt(1000),
			Status:  constants.VaultStatusActive,
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.POST("/vault", testController.CreateVault())

		req := httptest.NewRequest(
			http.MethodPost,
			"/vault",
			bytes.NewReader(body),
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
			Success bool         `json:"success"`
			Data    models.Vault `json:"data"`
		}

		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if response.Data.ID == uuid.Nil {
			t.Errorf("expected vault ID to be generated")
		}

		if response.Data.Balance.Cmp(requestBody.Balance) != 0 {
			t.Errorf(
				"expected balance %s, got %s",
				requestBody.Balance,
				response.Data.Balance,
			)
		}

		if response.Data.Type == "" || response.Data.Type != requestBody.Type {
			t.Errorf("expected vault type %v, got %v", requestBody.Type, response.Data.Type)
		}

		if response.Data.Status == "" || response.Data.Status != requestBody.Status {
			t.Errorf("expected vault status %v, got %v", requestBody.Status, response.Data.Status)
		}

		var vault models.Vault
		if err := db.First(&vault, "id = ?", response.Data.ID).Error; err != nil {
			t.Fatalf("failed to find created vault: %v", err)
		}
	})

	t.Run("duplicate vault type", func(t *testing.T) {
		vaultType := constants.PaymentVault
		vaultStatus := constants.VaultStatusActive

		requestBody := dto.CreateVaultRequest{
			Type:    vaultType,
			Balance: decimal.NewFromInt(2000),
			Status:  vaultStatus,
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.POST("/vault", testController.CreateVault())

		req := httptest.NewRequest(
			http.MethodPost,
			"/vault",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusInternalServerError,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

}

func TestGetVault(t *testing.T) {
	cleanTestDB(t)

	t.Run("invalid vault id", func(t *testing.T) {
		router := gin.New()
		router.GET("/vault/:vault_id", testController.GetVault())

		req := httptest.NewRequest(
			http.MethodGet,
			"/vault/invalid-uuid",
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

	t.Run("vault not found", func(t *testing.T) {
		vaultID := uuid.New()

		router := gin.New()
		router.GET("/vault/:vault_id", testController.GetVault())

		req := httptest.NewRequest(
			http.MethodGet,
			"/vault/"+vaultID.String(),
			nil,
		)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusInternalServerError,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("success", func(t *testing.T) {
		vaultType := constants.PaymentVault
		vaultStatus := constants.VaultStatusActive

		vault := models.Vault{
			ID:      uuid.New(),
			Type:    vaultType,
			Balance: decimal.NewFromInt(1000),
			Status:  vaultStatus,
		}

		if err := db.Create(&vault).Error; err != nil {
			t.Fatalf("failed to create vault: %v", err)
		}

		router := gin.New()
		router.GET("/vault/:vault_id", testController.GetVault())

		req := httptest.NewRequest(
			http.MethodGet,
			"/vault/"+vault.ID.String(),
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
			Success bool         `json:"success"`
			Data    models.Vault `json:"data"`
		}

		err := json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if response.Data.ID != vault.ID {
			t.Errorf(
				"expected vault ID %s, got %s",
				vault.ID,
				response.Data.ID,
			)
		}

		if response.Data.Balance.Cmp(vault.Balance) != 0 {
			t.Errorf(
				"expected balance %s, got %s",
				vault.Balance,
				response.Data.Balance,
			)
		}

		if response.Data.Type == "" || response.Data.Type != vault.Type {
			t.Errorf("expected vault type %v, got %v", vault.Type, response.Data.Type)
		}

		if response.Data.Status == "" || response.Data.Status != vault.Status {
			t.Errorf("expected vault status %v, got %v", vault.Status, response.Data.Status)
		}
	})
}

func TestGetVaults(t *testing.T) {
	cleanTestDB(t)

	t.Run("no vaults", func(t *testing.T) {
		router := gin.New()
		router.GET("/vaults", testController.GetVaults())

		req := httptest.NewRequest(
			http.MethodGet,
			"/vaults",
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
			Success bool            `json:"success"`
			Data    []*models.Vault `json:"data"`
		}

		err := json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if len(response.Data) != 0 {
			t.Errorf(
				"expected no vaults, got %d",
				len(response.Data),
			)
		}
	})

	t.Run("success", func(t *testing.T) {
		vaultType1 := constants.PaymentVault
		vaultStatus1 := constants.VaultStatusActive

		vaultType2 := constants.PayoutVault
		vaultStatus2 := constants.VaultStatusActive

		vault1 := models.Vault{
			ID:      uuid.New(),
			Type:    vaultType1,
			Balance: decimal.NewFromInt(1000),
			Status:  vaultStatus1,
		}

		vault2 := models.Vault{
			ID:      uuid.New(),
			Type:    vaultType2,
			Balance: decimal.NewFromInt(2000),
			Status:  vaultStatus2,
		}

		if err := db.Create(&vault1).Error; err != nil {
			t.Fatalf("failed to create vault1: %v", err)
		}

		if err := db.Create(&vault2).Error; err != nil {
			t.Fatalf("failed to create vault2: %v", err)
		}

		router := gin.New()
		router.GET("/vaults", testController.GetVaults())

		req := httptest.NewRequest(
			http.MethodGet,
			"/vaults",
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
			Success bool            `json:"success"`
			Data    []*models.Vault `json:"data"`
		}

		err := json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if len(response.Data) != 2 {
			t.Fatalf(
				"expected 2 vaults, got %d",
				len(response.Data),
			)
		}

		foundVault1 := false
		foundVault2 := false

		for _, vault := range response.Data {
			if vault.ID == vault1.ID {
				foundVault1 = true

				if vault.Balance.Cmp(vault1.Balance) != 0 {
					t.Errorf(
						"expected vault1 balance %s, got %s",
						vault1.Balance,
						vault.Balance,
					)
				}
			}

			if vault.ID == vault2.ID {
				foundVault2 = true

				if vault.Balance.Cmp(vault2.Balance) != 0 {
					t.Errorf(
						"expected vault2 balance %s, got %s",
						vault2.Balance,
						vault.Balance,
					)
				}
			}
		}

		if !foundVault1 {
			t.Errorf("expected vault1 in response")
		}

		if !foundVault2 {
			t.Errorf("expected vault2 in response")
		}
	})
}

func TestUpdateVaultBalance(t *testing.T) {
	cleanTestDB(t)

	t.Run("merchant id not found in context", func(t *testing.T) {
		router := gin.New()
		router.POST("/vault/balance", testController.UpdateVaultBalance())

		req := httptest.NewRequest(
			http.MethodPost,
			"/vault/balance",
			strings.NewReader(`{"balance":"100","type":"valid-type"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		router := gin.New()
		router.POST("/vault/balance", testController.UpdateVaultBalance())

		req := httptest.NewRequest(
			http.MethodPost,
			"/vault/balance",
			strings.NewReader(`{"bal":`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		c.Set("merchant_id", uuid.New())
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("service error", func(t *testing.T) {
		merchantID := uuid.New()

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		c.Set("merchant_id", merchantID)

		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/vault/balance",
			strings.NewReader(`{"balance":"0","type":"valid-type"}`),
		)
		c.Request.Header.Set("Content-Type", "application/json")

		testController.UpdateVaultBalance()(c)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusInternalServerError,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("success", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		bankAccount := testutils.GenerateTestBankAccount()
		if err := db.Create(bankAccount).Error; err != nil {
			t.Fatalf("failed to create bank account: %v", err)
		}

		linkedBankAccount := &models.LinkedBankAccount{
			MerchantID:    merchant.ID,
			BankAccountID: bankAccount.ID,
			Status:        merchant.Status,
			Type:          bankAccount.AccountType,
		}

		if errL := db.Create(linkedBankAccount).Error; errL != nil {
			t.Fatalf("failed to create linked ban account: %v", errL)
		}

		// Create the primary bank account link here if GenerateTestBankAccount()
		// does not create it automatically.

		vaultType := constants.PaymentVault
		vaultStatus := constants.VaultStatusActive

		vault := models.Vault{
			ID:      uuid.New(),
			Type:    vaultType,
			Balance: decimal.NewFromInt(1000),
			Status:  vaultStatus,
		}

		if err := db.Create(&vault).Error; err != nil {
			t.Fatalf("failed to create vault: %v", err)
		}

		requestBody := dto.UpdateVaultBalance{
			Balance: decimal.NewFromInt(100),
			Type:    vaultType,
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		c.Set("merchant_id", merchant.ID)

		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/vault/balance",
			bytes.NewReader(body),
		)
		c.Request.Header.Set("Content-Type", "application/json")

		testController.UpdateVaultBalance()(c)

		if rec.Code != http.StatusOK {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var response struct {
			Success bool         `json:"success"`
			Data    models.Vault `json:"data"`
		}

		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if response.Data.ID != vault.ID {
			t.Errorf(
				"expected vault ID %s, got %s",
				vault.ID,
				response.Data.ID,
			)
		}

		expectedBalance := decimal.NewFromInt(1100)

		if response.Data.Balance.Cmp(expectedBalance) != 0 {
			t.Errorf(
				"expected vault balance %s, got %s",
				expectedBalance,
				response.Data.Balance,
			)
		}
	})
}

func TestUpdateVaultStatus(t *testing.T) {
	cleanTestDB(t)

	vaultType := constants.PaymentVault
	activeStatus := "active"
	inactiveStatus := "inactive"

	t.Run("invalid request body", func(t *testing.T) {
		router := gin.New()
		router.PUT("/vault/:vault_id/status", testController.UpdateVaultStatus())

		vaultID := uuid.New()

		req := httptest.NewRequest(
			http.MethodPut,
			"/vault/"+vaultID.String()+"/status",
			bytes.NewBufferString(`{"status":`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusBadRequest,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("invalid vault id", func(t *testing.T) {
		body := dto.UpdateVaultStatus{
			Status: activeStatus,
		}

		requestBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.PUT("/vault/:vault_id/status", testController.UpdateVaultStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/vault/invalid-vault-id/status",
			bytes.NewReader(requestBody),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusBadRequest,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		invalidStatus := "invalid"

		body := dto.UpdateVaultStatus{
			Status: invalidStatus,
		}

		requestBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		vaultID := uuid.New()

		router := gin.New()
		router.PUT("/vault/:vault_id/status", testController.UpdateVaultStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/vault/"+vaultID.String()+"/status",
			bytes.NewReader(requestBody),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusInternalServerError,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("vault does not exist", func(t *testing.T) {
		body := dto.UpdateVaultStatus{
			Status: activeStatus,
		}

		requestBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		vaultID := uuid.New()

		router := gin.New()
		router.PUT("/vault/:vault_id/status", testController.UpdateVaultStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/vault/"+vaultID.String()+"/status",
			bytes.NewReader(requestBody),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusInternalServerError,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("same status", func(t *testing.T) {
		vault := models.Vault{
			ID:      uuid.New(),
			Type:    vaultType,
			Balance: decimal.NewFromInt(1000),
			Status:  activeStatus,
		}

		if err := db.Create(&vault).Error; err != nil {
			t.Fatalf("failed to create vault: %v", err)
		}

		body := dto.UpdateVaultStatus{
			Status: activeStatus,
		}

		requestBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.PUT("/vault/:vault_id/status", testController.UpdateVaultStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/vault/"+vault.ID.String()+"/status",
			bytes.NewReader(requestBody),
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
			Success bool         `json:"success"`
			Data    models.Vault `json:"data"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if response.Data.ID != vault.ID {
			t.Errorf(
				"expected vault id %s, got %s",
				vault.ID,
				response.Data.ID,
			)
		}

		if response.Data.Status == "" || response.Data.Status != activeStatus {
			t.Errorf("expected status %s", activeStatus)
		}
	})

	t.Run("success", func(t *testing.T) {
		vault := models.Vault{
			ID:      uuid.New(),
			Type:    constants.PayoutVault,
			Balance: decimal.NewFromInt(2000),
			Status:  activeStatus,
		}

		if err := db.Create(&vault).Error; err != nil {
			t.Fatalf("failed to create vault: %v", err)
		}

		body := dto.UpdateVaultStatus{
			Status: inactiveStatus,
		}

		requestBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.PUT("/vault/:vault_id/status", testController.UpdateVaultStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/vault/"+vault.ID.String()+"/status",
			bytes.NewReader(requestBody),
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
			Success bool         `json:"success"`
			Data    models.Vault `json:"data"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if response.Data.ID != vault.ID {
			t.Errorf(
				"expected vault id %s, got %s",
				vault.ID,
				response.Data.ID,
			)
		}

		if response.Data.Status == "" || response.Data.Status != inactiveStatus {
			t.Errorf(
				"expected status %s, got %v",
				inactiveStatus,
				response.Data.Status,
			)
		}

		var updatedVault models.Vault
		if err := db.First(&updatedVault, "id = ?", vault.ID).Error; err != nil {
			t.Fatalf("failed to fetch updated vault: %v", err)
		}

		if updatedVault.Status == "" || updatedVault.Status != inactiveStatus {
			t.Errorf(
				"expected database status %s, got %v",
				inactiveStatus,
				updatedVault.Status,
			)
		}
	})
}
