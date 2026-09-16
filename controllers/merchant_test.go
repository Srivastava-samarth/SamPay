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
)

func TestGetMerchantByID(t *testing.T) {
	t.Run("merchant id not provided", func(t *testing.T) {
		cleanTestDB(t)

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		c.Params = gin.Params{
			{Key: "merchant_id", Value: ""},
		}

		testController.GetMerchantByID()(c)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status 400, got %d",
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
			"/merchants/:merchant_id",
			testController.GetMerchantByID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/not-a-uuid",
			nil,
		)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf(
				"expected status 500, got %d",
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

		if response.Error.Code != "PARSING_ERROR" {
			t.Errorf(
				"expected PARSING_ERROR, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("merchant not found", func(t *testing.T) {
		cleanTestDB(t)

		merchantID := uuid.New()

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id",
			testController.GetMerchantByID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/"+merchantID.String(),
			nil,
		)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf(
				"expected status 404, got %d. response: %s",
				rec.Code,
				rec.Body.String(),
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

		if response.Error.Code != "MERCHANT_NOT_FOUND" {
			t.Errorf(
				"expected MERCHANT_NOT_FOUND, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("merchant found", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.GET(
			"/merchants/:merchant_id",
			testController.GetMerchantByID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/"+merchant.ID.String(),
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
			t.Fatalf("expected merchant data in response")
		}

		merchantData, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatalf(
				"expected merchant data to be object, got %T",
				response.Data,
			)
		}

		returnedID, ok := merchantData["id"].(string)
		if !ok {
			t.Fatalf(
				"expected merchant id to be string, got %T",
				merchantData["id"],
			)
		}

		if returnedID != merchant.ID.String() {
			t.Errorf(
				"expected merchant id %s, got %s",
				merchant.ID.String(),
				returnedID,
			)
		}
	})
}

func TestGetMerchants(t *testing.T) {
	t.Run("no merchants found", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.GET(
			"/merchants",
			testController.GetMerchants(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants",
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
			t.Fatalf("expected merchants data in response")
		}

		merchants, ok := response.Data.([]interface{})
		if !ok {
			t.Fatalf(
				"expected merchants data to be array, got %T",
				response.Data,
			)
		}

		if len(merchants) != 0 {
			t.Errorf(
				"expected 0 merchants, got %d",
				len(merchants),
			)
		}
	})

	t.Run("merchants found", func(t *testing.T) {
		cleanTestDB(t)

		merchant1 := testutils.GenerateTestMerchant()
		merchant2 := testutils.GenerateTestMerchant()

		if err := db.Create(merchant1).Error; err != nil {
			t.Fatalf("failed to create merchant1: %v", err)
		}

		if err := db.Create(merchant2).Error; err != nil {
			t.Fatalf("failed to create merchant2: %v", err)
		}

		router := gin.New()
		router.GET(
			"/merchants",
			testController.GetMerchants(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants",
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
			t.Fatalf("expected merchants data in response")
		}

		merchants, ok := response.Data.([]interface{})
		if !ok {
			t.Fatalf(
				"expected merchants data to be array, got %T",
				response.Data,
			)
		}

		if len(merchants) != 2 {
			t.Errorf(
				"expected 2 merchants, got %d",
				len(merchants),
			)
		}
	})
}

func TestCreateMerchant(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST(
			"/merchants",
			testController.CreateMerchant(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants",
			strings.NewReader(`{"merchant_type":`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status 400, got %d",
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

		if response.Error.Code != "INVALID_REQUEST" {
			t.Errorf(
				"expected INVALID_REQUEST, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("merchant already exists", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		request := map[string]interface{}{
			"merchant_type": "individual",
			"merchant_name": "New Merchant",
			"email":         merchant.Email,
			"phone_number":  "9876543210",
			"individual":    map[string]interface{}{},
		}

		body, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.POST(
			"/merchants",
			testController.CreateMerchant(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status 400, got %d. response: %s",
				rec.Code,
				rec.Body.String(),
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

		if response.Error.Code != "MERCHANT_ALREADY_EXIST" {
			t.Errorf(
				"expected MERCHANT_ALREADY_EXIST, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("individual details not provided", func(t *testing.T) {
		cleanTestDB(t)

		request := map[string]interface{}{
			"merchant_type": "individual",
			"merchant_name": "Test Individual",
			"email":         "individual@test.com",
			"phone_number":  "9876543210",
		}

		body, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.POST(
			"/merchants",
			testController.CreateMerchant(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status 400, got %d",
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

		if response.Error.Code != "INVALID_REQUEST" {
			t.Errorf(
				"expected INVALID_REQUEST, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("company details not provided", func(t *testing.T) {
		cleanTestDB(t)

		request := map[string]interface{}{
			"merchant_type": "company",
			"merchant_name": "Test Company",
			"email":         "company@test.com",
			"phone_number":  "9876543210",
		}

		body, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.POST(
			"/merchants",
			testController.CreateMerchant(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status 400, got %d",
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

		if response.Error.Code != "INVALID_REQUEST" {
			t.Errorf(
				"expected INVALID_REQUEST, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("successfully creates merchant", func(t *testing.T) {
		cleanTestDB(t)

		request := map[string]interface{}{
			"merchant_type": "individual",
			"merchant_name": "Test Merchant",
			"email":         "newmerchant@test.com",
			"phone_number":  "9876543210",
			"individual":    map[string]interface{}{},
		}

		body, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.POST(
			"/merchants",
			testController.CreateMerchant(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf(
				"expected status 202, got %d. response: %s",
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
			t.Fatalf("expected merchant data in response")
		}

		merchantData, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatalf(
				"expected merchant data to be object, got %T",
				response.Data,
			)
		}

		merchantID, ok := merchantData["id"].(string)
		if !ok {
			t.Fatalf(
				"expected merchant id to be string, got %T",
				merchantData["id"],
			)
		}

		if merchantID == "" {
			t.Errorf("expected merchant id to be non-empty")
		}

		merchantName, ok := merchantData["merchant_name"].(string)
		if !ok {
			t.Fatalf(
				"expected merchant_name to be string, got %T",
				merchantData["merchant_name"],
			)
		}

		if merchantName != "Test Merchant" {
			t.Errorf(
				"expected merchant name Test Merchant, got %s",
				merchantName,
			)
		}

		// Verify merchant was actually persisted.
		var merchant models.Merchant
		if err := db.First(
			&merchant,
			"id = ?",
			merchantID,
		).Error; err != nil {
			t.Fatalf(
				"failed to find created merchant: %v",
				err,
			)
		}

		if merchant.Email != "newmerchant@test.com" {
			t.Errorf(
				"expected email newmerchant@test.com, got %s",
				merchant.Email,
			)
		}

		if merchant.Status != constants.MerchantStatusPending {
			t.Errorf(
				"expected merchant status %s, got %s",
				constants.MerchantStatusPending,
				merchant.Status,
			)
		}

		if merchant.ComplianceStatus != constants.ComplianceStatusPending {
			t.Errorf(
				"expected compliance status %s, got %s",
				constants.ComplianceStatusPending,
				merchant.ComplianceStatus,
			)
		}
	})
}

func TestReattemptOnboardingKyc(t *testing.T) {
	t.Run("merchant id not provided", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/merchants/:merchant_id/kyc/reattempt", testController.ReattemptOnboardingKyc())

		reqBody := `{
			"individual": {}
		}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants//kyc/reattempt",
			strings.NewReader(reqBody),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "MERCHANT_ID_NOT_FOUND" {
			t.Errorf(
				"expected error code MERCHANT_ID_NOT_FOUND, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("invalid merchant id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/merchants/:merchant_id/kyc/reattempt", testController.ReattemptOnboardingKyc())

		reqBody := `{
			"individual": {}
		}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/invalid-merchant-id/kyc/reattempt",
			strings.NewReader(reqBody),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "PARSING_ERROR" {
			t.Errorf(
				"expected error code PARSING_ERROR, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/merchants/:merchant_id/kyc/reattempt", testController.ReattemptOnboardingKyc())

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		reqBody := `{
			"individual":
		`

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/"+merchant.ID.String()+"/kyc/reattempt",
			strings.NewReader(reqBody),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "INVALID_REQUEST" {
			t.Errorf(
				"expected error code INVALID_REQUEST, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("successful kyc reattempt", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		merchant.MerchantType = constants.MerchantTypeIndividual

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		reqBody := `{
			"individual": {
				"first_name": "Samarth",
				"last_name": "Srivastava",
				"email": "samarth.test@example.com",
				"date_of_birth": "2000-01-01",
				"country": "India"
			}
		}`

		router := gin.New()
		router.POST("/merchants/:merchant_id/kyc/reattempt", testController.ReattemptOnboardingKyc())

		req := httptest.NewRequest(
			http.MethodPost,
			"/merchants/"+merchant.ID.String()+"/kyc/reattempt",
			strings.NewReader(reqBody),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusAccepted,
				rec.Code,
				rec.Body.String(),
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf(
				"expected success to be true, got error: %+v",
				response.Error,
			)
		}
	})
}

func TestUpdateMerchantInfo(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		merchant.Status = constants.MerchantStatusActive

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id",
			testController.UpdateMerchantInfo(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String(),
			strings.NewReader(`{"merchant_name":`),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "INVALID_REQUEST" {
			t.Errorf(
				"expected error code INVALID_REQUEST, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("merchant id not provided", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id",
			testController.UpdateMerchantInfo(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/",
			strings.NewReader(`{
				"merchant_name": "Updated Merchant",
				"phone_number": "9999999999"
			}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusNotFound,
				rec.Code,
			)
		}
	})

	t.Run("invalid merchant id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id",
			testController.UpdateMerchantInfo(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/invalid-id",
			strings.NewReader(`{
				"merchant_name": "Updated Merchant",
				"phone_number": "9999999999"
			}`),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "PARSING_ERROR" {
			t.Errorf(
				"expected error code PARSING_ERROR, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("merchant not found", func(t *testing.T) {
		cleanTestDB(t)

		merchantID := uuid.New()

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id",
			testController.UpdateMerchantInfo(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchantID.String(),
			strings.NewReader(`{
				"merchant_name": "Updated Merchant",
				"phone_number": "9999999999"
			}`),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "MERCHANT_NOT_FOUND" {
			t.Errorf(
				"expected error code MERCHANT_NOT_FOUND, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("merchant is not active", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		merchant.Status = constants.MerchantStatusPending

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id",
			testController.UpdateMerchantInfo(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String(),
			strings.NewReader(`{
				"merchant_name": "Updated Merchant",
				"phone_number": "9999999999"
			}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusConflict,
				rec.Code,
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "MERCHANT_NOT_ACTIVE" {
			t.Errorf(
				"expected error code MERCHANT_NOT_ACTIVE, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("successful merchant update", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		merchant.Status = constants.MerchantStatusActive

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		originalName := merchant.MerchantName
		originalPhone := merchant.PhoneNumber

		newMerchantName := "Updated Merchant"
		newPhoneNumber := "3295654589"

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id",
			testController.UpdateMerchantInfo(),
		)

		reqBody := `{
			"merchant_name": "` + newMerchantName + `",
			"phone_number": "` + newPhoneNumber + `"
		}`

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String(),
			strings.NewReader(reqBody),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf(
				"expected success to be true, got error: %+v",
				response.Error,
			)
		}

		var updatedMerchant models.Merchant
		if err := db.First(&updatedMerchant, "id = ?", merchant.ID).Error; err != nil {
			t.Fatalf("failed to fetch updated merchant: %v", err)
		}

		if updatedMerchant.MerchantName != newMerchantName {
			t.Errorf(
				"expected merchant name %q, got %q",
				newMerchantName,
				updatedMerchant.MerchantName,
			)
		}

		if updatedMerchant.PhoneNumber != newPhoneNumber {
			t.Errorf(
				"expected phone number %q, got %q",
				newPhoneNumber,
				updatedMerchant.PhoneNumber,
			)
		}

		if updatedMerchant.MerchantName == originalName {
			t.Errorf("merchant name was not updated")
		}

		if updatedMerchant.PhoneNumber == originalPhone {
			t.Errorf("phone number was not updated")
		}
	})
}

func TestUpdateMerchantStatus(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id/status",
			testController.UpdateMerchantStatus(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String()+"/status",
			strings.NewReader(`{"status":`),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "INVALID_REQUEST" {
			t.Errorf(
				"expected error code INVALID_REQUEST, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("merchant id not provided", func(t *testing.T) {
		cleanTestDB(t)
		handler := testController.UpdateMerchantStatus()

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(
			http.MethodPut,
			"/merchants/status",
			strings.NewReader(`{"status":"active"}`),
		)
		c.Request.Header.Set("Content-Type", "application/json")

		handler(c)

		if c.Writer.Status() != http.StatusBadRequest {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				c.Writer.Status(),
			)
		}
	})

	t.Run("invalid merchant id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id/status",
			testController.UpdateMerchantStatus(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/invalid-id/status",
			strings.NewReader(`{"status":"active"}`),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "PARSING_ERROR" {
			t.Errorf(
				"expected error code PARSING_ERROR, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("invalid merchant status", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id/status",
			testController.UpdateMerchantStatus(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String()+"/status",
			strings.NewReader(`{"status":"invalid_status"}`),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "UPDATION_MERCHANT_FAILED" {
			t.Errorf(
				"expected error code UPDATION_MERCHANT_FAILED, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("status unchanged", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		merchant.Status = constants.MerchantStatusActive

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id/status",
			testController.UpdateMerchantStatus(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String()+"/status",
			strings.NewReader(`{"status":"active"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusOK,
				rec.Code,
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf(
				"expected success to be true, got error: %+v",
				response.Error,
			)
		}

		var updatedMerchant models.Merchant
		if err := db.First(&updatedMerchant, "id = ?", merchant.ID).Error; err != nil {
			t.Fatalf("failed to fetch merchant: %v", err)
		}

		if updatedMerchant.Status != constants.MerchantStatusActive {
			t.Errorf(
				"expected merchant status %q, got %q",
				constants.MerchantStatusActive,
				updatedMerchant.Status,
			)
		}
	})

	t.Run("status update fails when no user is linked", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		merchant.Status = constants.MerchantStatusPending

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id/status",
			testController.UpdateMerchantStatus(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String()+"/status",
			strings.NewReader(`{"status":"active"}`),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "UPDATION_MERCHANT_FAILED" {
			t.Errorf(
				"expected error code UPDATION_MERCHANT_FAILED, got %s",
				response.Error.Code,
			)
		}

		var unchangedMerchant models.Merchant
		if err := db.First(&unchangedMerchant, "id = ?", merchant.ID).Error; err != nil {
			t.Fatalf("failed to fetch merchant: %v", err)
		}

		if unchangedMerchant.Status != constants.MerchantStatusPending {
			t.Errorf(
				"expected merchant status to rollback to %q, got %q",
				constants.MerchantStatusPending,
				unchangedMerchant.Status,
			)
		}
	})

	t.Run("successful status update with linked user", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		merchant.Status = constants.MerchantStatusPending

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		user := &models.User{
			ID:                 utils.GenerateUUID(),
			FirstName:          "Test",
			LastName:           "Merchant",
			Status:             constants.UserStatusActive,
			MustChangePassword: false,
			PasswordHash:       "$2a$10$EV4UdjJqM0mVvINQWTQRsOvUxKGQWyrpDd29bssvUT8xi5HLZ0mxK",
		}

		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		merchantUser := &models.MerchantUser{
			ID:         utils.GenerateUUID(),
			MerchantID: merchant.ID,
			UserID:     user.ID,
			Role:       "owner",
		}

		if err := db.Create(merchantUser).Error; err != nil {
			t.Fatalf("failed to create merchant user: %v", err)
		}

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id/status",
			testController.UpdateMerchantStatus(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String()+"/status",
			strings.NewReader(`{"status":"active"}`),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf(
				"expected success to be true, got error: %+v",
				response.Error,
			)
		}

		var updatedMerchant models.Merchant
		if err := db.First(&updatedMerchant, "id = ?", merchant.ID).Error; err != nil {
			t.Fatalf("failed to fetch merchant: %v", err)
		}

		if updatedMerchant.Status != constants.MerchantStatusActive {
			t.Errorf(
				"expected merchant status %q, got %q",
				constants.MerchantStatusActive,
				updatedMerchant.Status,
			)
		}
	})
}

func TestUpdateMerchantKyc(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		merchant.Status = constants.MerchantStatusActive

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id/kyc",
			testController.UpdateMerchantKyc(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String()+"/kyc",
			strings.NewReader(`{"individual_update":`),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "INVALID_REQUEST" {
			t.Errorf(
				"expected error code INVALID_REQUEST, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("merchant id not provided", func(t *testing.T) {
		cleanTestDB(t)

		handler := testController.UpdateMerchantKyc()

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(
			http.MethodPut,
			"/merchants/kyc",
			strings.NewReader(`{
				"individual_update": {
					"first_name": "Samarth",
					"last_name": "Srivastava",
					"date_of_birth": "2000-01-01",
					"country": "IN",
					"tax_id": "TEST123",
					"address": "Test Address"
				}
			}`),
		)
		c.Request.Header.Set("Content-Type", "application/json")

		handler(c)

		if c.Writer.Status() != http.StatusBadRequest {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				c.Writer.Status(),
			)
		}
	})

	t.Run("invalid merchant id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id/kyc",
			testController.UpdateMerchantKyc(),
		)

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/invalid-id/kyc",
			strings.NewReader(`{
				"individual_update": {
					"first_name": "Samarth",
					"last_name": "Srivastava",
					"date_of_birth": "2000-01-01",
					"country": "IN",
					"tax_id": "TEST123",
					"address": "Test Address"
				}
			}`),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "PARSING_ERROR" {
			t.Errorf(
				"expected error code PARSING_ERROR, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("merchant not active", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		merchant.Status = constants.MerchantStatusPending
		merchant.MerchantType = constants.MerchantTypeIndividual

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id/kyc",
			testController.UpdateMerchantKyc(),
		)

		reqBody := `{
			"individual_update": {
				"first_name": "Samarth",
				"last_name": "Srivastava",
				"date_of_birth": "2000-01-01",
				"country": "IN",
				"tax_id": "TEST123",
				"address": "Test Address"
			}
		}`

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String()+"/kyc",
			strings.NewReader(reqBody),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "UPDATION_MERCHANT_FAILED" {
			t.Errorf(
				"expected error code UPDATION_MERCHANT_FAILED, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("individual update is required", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		merchant.Status = constants.MerchantStatusActive
		merchant.MerchantType = constants.MerchantTypeIndividual

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id/kyc",
			testController.UpdateMerchantKyc(),
		)

		reqBody := `{
			"company_update": {
				"legal_name": "Test Company",
				"tax_id": "TAX123"
			}
		}`

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String()+"/kyc",
			strings.NewReader(reqBody),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "UPDATION_MERCHANT_FAILED" {
			t.Errorf(
				"expected error code UPDATION_MERCHANT_FAILED, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("company update is required", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()
		merchant.Status = constants.MerchantStatusActive
		merchant.MerchantType = constants.MerrchantTypeCorporate

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.PUT(
			"/merchants/:merchant_id/kyc",
			testController.UpdateMerchantKyc(),
		)

		reqBody := `{
			"individual_update": {
				"first_name": "Samarth",
				"date_of_birth": "2000-01-01",
				"country": "IN",
				"tax_id": "TAX123",
				"address": "Test Address"
			}
		}`

		req := httptest.NewRequest(
			http.MethodPut,
			"/merchants/"+merchant.ID.String()+"/kyc",
			strings.NewReader(reqBody),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Errorf("expected success to be false")
		}

		if response.Error.Code != "UPDATION_MERCHANT_FAILED" {
			t.Errorf(
				"expected error code UPDATION_MERCHANT_FAILED, got %s",
				response.Error.Code,
			)
		}
	})
}
