package repositories

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestGetBlockedUserByEmail(t *testing.T) {
	email := uuid.NewString() + "@test.com"
	passwordHash := "test-password-hash"
	firstName := "Test"
	lastName := "User"
	status := constants.MerchantStatusSuspended
	mustChangePassword := false

	user := &models.User{
		ID:                 uuid.New(),
		Email:              email,
		PasswordHash:       passwordHash,
		FirstName:          firstName,
		LastName:           lastName,
		Status:             status,
		MustChangePassword: mustChangePassword,
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	result, err := testRepo.GetBlockedUserByEmail(email)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result {
		t.Error("expected blocked user to return true")
	}

	nonExistingEmail := uuid.NewString() + "@test.com"

	result, err = testRepo.GetBlockedUserByEmail(nonExistingEmail)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result {
		t.Error("expected non-existing user to return false")
	}
}

func TestCreateUser(t *testing.T) {

	email := uuid.NewString() + "@test.com"
	passwordHash := "test-password-hash"
	firstName := "Test"
	lastName := "User"
	status := "active"

	user := &models.User{
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Status:       status,
	}

	result, err := testRepo.CreateUser(user)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected generated ID")
	}

	if result.Email != user.Email {
		t.Error("expected email to match")
	}

	if result.PasswordHash != user.PasswordHash {
		t.Error("expected password hash to match")
	}

	if result.MustChangePassword != true {
		t.Error("expected MustChangePassword to be true")
	}

	if result.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if result.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestGetUserByEmail(t *testing.T) {

	email := uuid.NewString() + "@test.com"
	passwordHash := "test-password-hash"
	firstName := "Test"
	lastName := "User"
	status := "active"
	mustChangePassword := false

	user := &models.User{
		ID:                 uuid.New(),
		Email:              email,
		PasswordHash:       passwordHash,
		FirstName:          firstName,
		LastName:           lastName,
		Status:             status,
		MustChangePassword: mustChangePassword,
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	result, err := testRepo.GetUserByEmail(email)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user, got nil")
	}

	if result.ID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, result.ID)
	}

	if result.Email != email {
		t.Errorf("expected email %v, got %v", email, result.Email)
	}

	nonExistingEmail := uuid.NewString() + "@test.com"

	_, err = testRepo.GetUserByEmail(nonExistingEmail)

	if err == nil {
		t.Fatal("expected error for non-existing email")
	}

	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected record not found error, got %v", err)
	}
}

func TestUpdateUser(t *testing.T) {

	email := uuid.NewString() + "@test.com"
	passwordHash := "old-password"
	firstName := "Old"
	lastName := "Name"
	status := "active"
	mustChangePassword := false

	user := &models.User{
		ID:                 uuid.New(),
		Email:              email,
		PasswordHash:       passwordHash,
		FirstName:          firstName,
		LastName:           lastName,
		Status:             status,
		MustChangePassword: mustChangePassword,
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	mustChangePassword = true
	request := &dto.UpdateUserRequest{
		PasswordHash:       "new-password",
		FirstName:          "New",
		LastName:           "User",
		MustChangePassword: &mustChangePassword,
	}

	result, err := testRepo.UpdateUser(user.ID, request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user, got nil")
	}

	if result.PasswordHash != "new-password" {
		t.Errorf("expected password hash to be updated")
	}

	if result.FirstName != "New" {
		t.Errorf("expected first name to be updated")
	}

	if result.LastName != "User" {
		t.Errorf("expected last name to be updated")
	}

	if !result.MustChangePassword {
		t.Error("expected MustChangePassword to be true")
	}

	request = &dto.UpdateUserRequest{}

	result, err = testRepo.UpdateUser(user.ID, request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PasswordHash != "new-password" ||
		result.FirstName != "New" ||
		result.LastName != "User" {
		t.Error("expected existing values to remain unchanged")
	}
}

func TestGetUserByID(t *testing.T) {

	email := uuid.NewString() + "@test.com"
	passwordHash := "test-password-hash"
	firstName := "Test"
	lastName := "User"
	status := "active"
	mustChangePassword := false

	user := &models.User{
		ID:                 uuid.New(),
		Email:              email,
		PasswordHash:       passwordHash,
		FirstName:          firstName,
		LastName:           lastName,
		Status:             status,
		MustChangePassword: mustChangePassword,
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	result, err := testRepo.GetUserByID(user.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user, got nil")
	}

	if result.ID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, result.ID)
	}

	nonExistingID := uuid.New()

	_, err = testRepo.GetUserByID(nonExistingID)

	if err == nil {
		t.Fatal("expected error for non-existing user")
	}

	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected record not found error, got %v", err)
	}
}

func TestGetUsers(t *testing.T) {

	email1 := uuid.NewString() + "@test.com"
	email2 := uuid.NewString() + "@test.com"
	passwordHash := "test-password-hash"
	firstName := "Test"
	lastName := "User"
	status := "active"
	mustChangePassword := false

	user1 := &models.User{
		ID:                 uuid.New(),
		Email:              email1,
		PasswordHash:       passwordHash,
		FirstName:          firstName,
		LastName:           lastName,
		Status:             status,
		MustChangePassword: mustChangePassword,
	}

	user2 := &models.User{
		ID:                 uuid.New(),
		Email:              email2,
		PasswordHash:       passwordHash,
		FirstName:          firstName,
		LastName:           lastName,
		Status:             status,
		MustChangePassword: mustChangePassword,
	}

	if err := db.Create(user1).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := db.Create(user2).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	result, err := testRepo.GetUsers()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) < 2 {
		t.Fatalf("expected at least 2 users, got %d", len(result))
	}

	found1, found2 := false, false

	for _, user := range result {
		if user.ID == user1.ID {
			found1 = true
		}

		if user.ID == user2.ID {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("expected created users to be returned")
	}
}

func TestUpdateUserStatus(t *testing.T) {

	email := uuid.NewString() + "@test.com"
	passwordHash := "test-password-hash"
	firstName := "Test"
	lastName := "User"
	status := "active"
	mustChangePassword := false

	user := &models.User{
		ID:                 uuid.New(),
		Email:              email,
		PasswordHash:       passwordHash,
		FirstName:          firstName,
		LastName:           lastName,
		Status:             status,
		MustChangePassword: mustChangePassword,
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	newStatus := "suspended"

	result, err := testRepo.UpdateUserStatus(newStatus, user.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user, got nil")
	}

	if result.ID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, result.ID)
	}

	if result.Status == "" || result.Status != newStatus {
		t.Errorf("expected status %v, got %v", newStatus, result.Status)
	}
}
