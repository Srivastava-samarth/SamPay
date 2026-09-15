package services

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

func TestUpdateUser(t *testing.T) {

	tests := []struct {
		name          string
		request       *dto.UpdateUserRequest
		expectedError string
	}{
		{
			name: "missing first name",
			request: &dto.UpdateUserRequest{
				FirstName:          "",
				LastName:           "Doe",
				PasswordHash:       "hashed-password",
				MustChangePassword: true,
			},
			expectedError: "first_name is required",
		},
		{
			name: "missing last name",
			request: &dto.UpdateUserRequest{
				FirstName:          "John",
				LastName:           "",
				PasswordHash:       "hashed-password",
				MustChangePassword: true,
			},
			expectedError: "last_name is required",
		},
		{
			name: "missing password hash",
			request: &dto.UpdateUserRequest{
				FirstName:          "John",
				LastName:           "Doe",
				PasswordHash:       "",
				MustChangePassword: true,
			},
			expectedError: "password_hash is required",
		},
		{
			name: "must change password is false",
			request: &dto.UpdateUserRequest{
				FirstName:          "John",
				LastName:           "Doe",
				PasswordHash:       "hashed-password",
				MustChangePassword: false,
			},
			expectedError: "must_change_password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testServices.UpdateUser(tt.request, uuid.New())

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Fatalf(
					"expected error %q, got %q",
					tt.expectedError,
					err.Error(),
				)
			}
		})
	}
}

func TestUpdateUserStatus(t *testing.T) {

	tests := []struct {
		name          string
		status        string
		userID        uuid.UUID
		expectedError string
	}{
		{
			name:          "nil user id",
			status:        constants.UserStatusActive,
			userID:        uuid.Nil,
			expectedError: "user_id is required",
		},
		{
			name:          "empty status",
			status:        "",
			userID:        uuid.New(),
			expectedError: "status is required",
		},
		{
			name:          "invalid status",
			status:        "invalid",
			userID:        uuid.New(),
			expectedError: "invalid user status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testServices.UpdateUserStatus(tt.status, tt.userID)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Fatalf(
					"expected error %q, got %q",
					tt.expectedError,
					err.Error(),
				)
			}
		})
	}
}

func TestGetUsersByMerchant(t *testing.T) {

	_, err := testServices.GetUsersByMerchant(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "merchant_id is required" {
		t.Fatalf(
			"expected error %q, got %q",
			"merchant_id is required",
			err.Error(),
		)
	}
}

func TestGetUser(t *testing.T) {

	_, err := testServices.GetUser(uuid.Nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "user_id is required" {
		t.Fatalf(
			"expected error %q, got %q",
			"user_id is required",
			err.Error(),
		)
	}
}
