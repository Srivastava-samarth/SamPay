package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestUserOnboarding(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST(
			"/user/onboarding",
			testController.UserOnboarding(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/user/onboarding",
			bytes.NewReader([]byte(`{"merchant_id":`)),
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Fatal("expected success to be false")
		}

		if response.Error == nil {
			t.Fatal("expected error information")
		}

		if response.Error.Code != "INVALID_REQUEST" {
			t.Fatalf(
				"expected error code INVALID_REQUEST, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("successful user onboarding", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateCorporateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		requestBody := dto.UserOnboardingRequest{
			MerchantID: merchant.ID,
			Email:      uuid.NewString() + "@test.com",
		}

		body, err := json.Marshal(requestBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		router := gin.New()
		router.POST(
			"/user/onboarding",
			testController.UserOnboarding(),
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/user/onboarding",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Fatalf(
				"expected success to be true, got error: %+v",
				response.Error,
			)
		}

		var returnedEmail string
		dataBytes, err := json.Marshal(response.Data)
		if err != nil {
			t.Fatalf("failed to marshal response data: %v", err)
		}

		if err := json.Unmarshal(dataBytes, &returnedEmail); err != nil {
			t.Fatalf("failed to decode returned email: %v", err)
		}

		if returnedEmail != requestBody.Email {
			t.Fatalf(
				"expected email %s, got %s",
				requestBody.Email,
				returnedEmail,
			)
		}
	})
}

func TestGetUserByID(t *testing.T) {
	t.Run("invalid user id", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.GET(
			"/users/:user_id",
			testController.GetUserByID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/users/invalid-user-id",
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

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Fatal("expected success to be false")
		}

		if response.Error == nil {
			t.Fatal("expected error information")
		}

		if response.Error.Code != "PARSING_ERROR" {
			t.Fatalf(
				"expected error code PARSING_ERROR, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.GET(
			"/users/:user_id",
			testController.GetUserByID(),
		)

		userID := uuid.New()

		req := httptest.NewRequest(
			http.MethodGet,
			"/users/"+userID.String(),
			nil,
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusNotFound,
				rec.Code,
			)
		}

		var response dto.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Success {
			t.Fatal("expected success to be false")
		}

		if response.Error == nil {
			t.Fatal("expected error information")
		}

		if response.Error.Code != "USER_NOT_FOUND" {
			t.Fatalf(
				"expected error code USER_NOT_FOUND, got %s",
				response.Error.Code,
			)
		}
	})

	t.Run("successful get user", func(t *testing.T) {
		cleanTestDB(t)

		// Use your existing user generator here.
		user := &models.User{
			ID:                 uuid.New(),
			Email:              "test@gmail.com",
			FirstName:          "Test",
			LastName:           "Merchant",
			PasswordHash:       "$2a$10$w5t2jGhkssbV0qThlRRiYuLDkfY528p/xeGuZLA66aS5zPCR94jvC",
			Status:             constants.UserStatusActive,
			MustChangePassword: false,
		}

		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		router := gin.New()
		router.GET(
			"/users/:user_id",
			testController.GetUserByID(),
		)

		req := httptest.NewRequest(
			http.MethodGet,
			"/users/"+user.ID.String(),
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
			Data    models.User    `json:"data"`
			Error   *dto.ErrorInfo `json:"error,omitempty"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Fatalf(
				"expected success to be true, got error: %+v",
				response.Error,
			)
		}

		if response.Data.ID != user.ID {
			t.Fatalf(
				"expected user ID %s, got %s",
				user.ID,
				response.Data.ID,
			)
		}
	})
}

func TestGetUsers(t *testing.T) {
	cleanTestDB(t)

	t.Run("no users", func(t *testing.T) {
		router := gin.New()
		router.GET("/users", testController.GetUsers())

		req := httptest.NewRequest(
			http.MethodGet,
			"/users",
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
			Success bool           `json:"success"`
			Data    []*models.User `json:"data"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if len(response.Data) != 0 {
			t.Errorf(
				"expected 0 users, got %d",
				len(response.Data),
			)
		}
	})

	t.Run("success", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		user1 := &models.User{
			ID:                 uuid.New(),
			Email:              "test1@gmail.com",
			FirstName:          "Test",
			LastName:           "Merchant1",
			PasswordHash:       "$2a$10$w5t2jGhkssbV0qThlRRiYuLDkfY528p/xeGuZLA66aS5zPCR94jvC",
			Status:             constants.UserStatusActive,
			MustChangePassword: false,
		}

		user2 := &models.User{
			ID:                 uuid.New(),
			Email:              "test2@gmail.com",
			FirstName:          "Test",
			LastName:           "Merchant2",
			PasswordHash:       "$2a$10$w5t2jGhkssbV0qThlRRiYuLDkfY528p/xeGuZLA66aS5zPCR94jvC",
			Status:             constants.UserStatusActive,
			MustChangePassword: false,
		}

		if err := db.Create(user1).Error; err != nil {
			t.Fatalf("failed to create user1: %v", err)
		}

		if err := db.Create(user2).Error; err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		router := gin.New()
		router.GET("/users", testController.GetUsers())

		req := httptest.NewRequest(
			http.MethodGet,
			"/users",
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
			Success bool           `json:"success"`
			Data    []*models.User `json:"data"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if len(response.Data) != 2 {
			t.Errorf(
				"expected 2 users, got %d",
				len(response.Data),
			)
		}

		foundUser1 := false
		foundUser2 := false

		for _, user := range response.Data {
			if user.ID == user1.ID {
				foundUser1 = true
			}

			if user.ID == user2.ID {
				foundUser2 = true
			}
		}

		if !foundUser1 {
			t.Errorf("user1 was not returned")
		}

		if !foundUser2 {
			t.Errorf("user2 was not returned")
		}
	})
}

func TestUpdateUserStatus(t *testing.T) {
	cleanTestDB(t)

	activeStatus := "active"
	inactiveStatus := "inactive"

	t.Run("invalid request body", func(t *testing.T) {
		userID := uuid.New()

		router := gin.New()
		router.PUT("/users/:user_id/status", testController.UpdateUserStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/users/"+userID.String()+"/status",
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

	t.Run("invalid user id", func(t *testing.T) {
		body := dto.UpdateMerchantStatusRequest{
			Status: inactiveStatus,
		}

		requestBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.PUT("/users/:user_id/status", testController.UpdateUserStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/users/invalid-user-id/status",
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

	t.Run("missing status", func(t *testing.T) {
		body := dto.UpdateMerchantStatusRequest{}

		requestBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		userID := uuid.New()

		router := gin.New()
		router.PUT("/users/:user_id/status", testController.UpdateUserStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/users/"+userID.String()+"/status",
			bytes.NewReader(requestBody),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status %d, got %d, body: %s",
				http.StatusInternalServerError,
				rec.Code,
				rec.Body.String(),
			)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		invalidStatus := "invalid-status"

		body := dto.UpdateMerchantStatusRequest{
			Status: invalidStatus,
		}

		requestBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		userID := uuid.New()

		router := gin.New()
		router.PUT("/users/:user_id/status", testController.UpdateUserStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/users/"+userID.String()+"/status",
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

	t.Run("user does not exist", func(t *testing.T) {
		body := dto.UpdateMerchantStatusRequest{
			Status: inactiveStatus,
		}

		requestBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		userID := uuid.New()

		router := gin.New()
		router.PUT("/users/:user_id/status", testController.UpdateUserStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/users/"+userID.String()+"/status",
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
		user := &models.User{
			ID:     uuid.New(),
			Status: activeStatus,
		}

		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		body := dto.UpdateMerchantStatusRequest{
			Status: activeStatus,
		}

		requestBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.PUT("/users/:user_id/status", testController.UpdateUserStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/users/"+user.ID.String()+"/status",
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
			Success bool        `json:"success"`
			Data    models.User `json:"data"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if response.Data.ID != user.ID {
			t.Errorf(
				"expected user id %s, got %s",
				user.ID,
				response.Data.ID,
			)
		}

		if response.Data.Status != activeStatus {
			t.Errorf(
				"expected status %s, got %s",
				activeStatus,
				response.Data.Status,
			)
		}
	})

	t.Run("success", func(t *testing.T) {
		user := &models.User{
			ID:                 uuid.New(),
			Email:              "test1@gmail.com",
			FirstName:          "Test",
			LastName:           "Merchant1",
			PasswordHash:       "$2a$10$w5t2jGhkssbV0qThlRRiYuLDkfY528p/xeGuZLA66aS5zPCR94jvC",
			Status:             constants.UserStatusActive,
			MustChangePassword: false,
		}

		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		body := dto.UpdateMerchantStatusRequest{
			Status: inactiveStatus,
		}

		requestBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		router := gin.New()
		router.PUT("/users/:user_id/status", testController.UpdateUserStatus())

		req := httptest.NewRequest(
			http.MethodPut,
			"/users/"+user.ID.String()+"/status",
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
			Success bool        `json:"success"`
			Data    models.User `json:"data"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if response.Data.ID != user.ID {
			t.Errorf(
				"expected user id %s, got %s",
				user.ID,
				response.Data.ID,
			)
		}

		if response.Data.Status != inactiveStatus {
			t.Errorf(
				"expected status %s, got %s",
				inactiveStatus,
				response.Data.Status,
			)
		}

		var updatedUser models.User
		if err := db.First(&updatedUser, "id = ?", user.ID).Error; err != nil {
			t.Fatalf("failed to fetch updated user: %v", err)
		}

		if updatedUser.Status != inactiveStatus {
			t.Errorf(
				"expected database status %s, got %s",
				inactiveStatus,
				updatedUser.Status,
			)
		}
	})
}

func TestGetUsersByMerchant(t *testing.T) {
	cleanTestDB(t)

	t.Run("merchant id not provided", func(t *testing.T) {
		router := gin.New()
		router.GET("/merchants/:merchant_id/users", testController.GetUsersByMerchant())

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants//users",
			nil,
		)

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

	t.Run("invalid merchant id", func(t *testing.T) {
		router := gin.New()
		router.GET("/merchants/:merchant_id/users", testController.GetUsersByMerchant())

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/invalid-merchant-id/users",
			nil,
		)

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

	t.Run("merchant has no users", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		router := gin.New()
		router.GET("/merchants/:merchant_id/users", testController.GetUsersByMerchant())

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/"+merchant.ID.String()+"/users",
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
			Success bool           `json:"success"`
			Data    []*models.User `json:"data"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if len(response.Data) != 0 {
			t.Errorf(
				"expected 0 users, got %d",
				len(response.Data),
			)
		}
	})

	t.Run("success", func(t *testing.T) {
		merchant1 := testutils.GenerateCorporateTestMerchant()

		if err := db.Create(merchant1).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		user1 := &models.User{
			ID:     uuid.New(),
			Status: "active",
			Email:  "test12@gmail.com",
		}

		user2 := &models.User{
			ID:     uuid.New(),
			Status: "active",
			Email:  "test13@gmail.com",
		}

		if err := db.Create(user1).Error; err != nil {
			t.Fatalf("failed to create user1: %v", err)
		}

		if err := db.Create(user2).Error; err != nil {
			t.Fatalf("failed to create user2: %v", err)
		}

		merchantUser1 := &models.MerchantUser{
			MerchantID: merchant1.ID,
			UserID:     user1.ID,
		}

		merchantUser2 := &models.MerchantUser{
			MerchantID: merchant1.ID,
			UserID:     user2.ID,
		}

		if err := db.Create(merchantUser1).Error; err != nil {
			t.Fatalf("failed to create merchant user 1: %v", err)
		}

		if err := db.Create(merchantUser2).Error; err != nil {
			t.Fatalf("failed to create merchant user 2: %v", err)
		}

		router := gin.New()
		router.GET("/merchants/:merchant_id/users", testController.GetUsersByMerchant())

		req := httptest.NewRequest(
			http.MethodGet,
			"/merchants/"+merchant1.ID.String()+"/users",
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
			Success bool           `json:"success"`
			Data    []*models.User `json:"data"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Errorf("expected success to be true")
		}

		if len(response.Data) != 2 {
			t.Errorf(
				"expected 2 users, got %d",
				len(response.Data),
			)
		}

		foundUser1 := false
		foundUser2 := false

		for _, user := range response.Data {
			if user.ID == user1.ID {
				foundUser1 = true
			}

			if user.ID == user2.ID {
				foundUser2 = true
			}
		}

		if !foundUser1 {
			t.Errorf("user1 was not returned")
		}

		if !foundUser2 {
			t.Errorf("user2 was not returned")
		}
	})
}
