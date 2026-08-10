package dto

import "github.com/google/uuid"

type CreateMerchantUserRequest struct {
	MerchantID uuid.UUID `json:"merchant_id" gorm:"not null"`
	UserID     uuid.UUID `json:"user_id" gorm:"not null"`
	Role       string    `json:"role" gorm:"not null"`
}

type CreateMerchantUserResponse struct {
	MerchantUserLinkedID uuid.UUID `json:"merchant_user_linked_id" gorm:"not null"`
	MerchantID           uuid.UUID `json:"merchant_id" gorm:"not null"`
	UserID               uuid.UUID `json:"user_id" gorm:"not null"`
	Role                 string    `json:"role" gorm:"not null"`
}
