package database

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID `json:"id" gorm:"primaryKey"`
	Email              string    `json:"email" gorm:"unique;not null"`
	PasswordHash       string    `json:"password_hash" gorm:"not null"`
	FirstName          string    `json:"first_name" gorm:"not null"`
	LastName           string    `json:"last_name" gorm:"not null"`
	Status             string    `json:"status" gorm:"default:active;not null"`
	MustChangePassword bool      `json:"must_change_password" gorm:"default:true"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
