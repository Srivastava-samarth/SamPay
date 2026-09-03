package dto

import "github.com/google/uuid"

type CreateUserRequest struct {
	Email     string `json:"email" gorm:"unique;not null"`
	FirstName string `json:"first_name" gorm:"not null"`
	LastName  string `json:"last_name" gorm:"not null"`
}

type CreateUserResponse struct {
	ID                 uuid.UUID `json:"id"`
	Email              string    `json:"email"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	Status             string    `json:"status"`
	MustChangePassword bool      `json:"must_change_password"`
}

type UserOnboardingRequest struct {
	MerchantID uuid.UUID `json:"merchant_id" gorm:"not null"`
	Email      string    `json:"email" gorm:"unique;not null"`
	FirstName  string    `json:"first_name" gorm:"not null"`
	LastName   string    `json:"last_name" gorm:"not null"`
	Role       string    `json:"role" gorm:"not null"`
}

type UpdateUserRequest struct {
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	Status             string `json:"status"`
	PasswordHash       string `json:"password_hash"`
	MustChangePassword bool   `json:"must_change_password"`
}
