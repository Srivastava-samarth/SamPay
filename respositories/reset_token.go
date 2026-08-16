package repositories

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm"
)

type PasswordResetTokenRepository struct {
	DB *gorm.DB
}

func NewPasswordResetTokenRepository(
	db *gorm.DB,
) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{
		DB:db,
	}
}

func (pr *PasswordResetTokenRepository) CreatePasswordReset(resetRequest *models.PasswordResetToken) (*models.PasswordResetToken, error){
	resetRequestPayload := &models.PasswordResetToken{
		ID: utils.GenerateUUID(),
		UserID: resetRequest.UserID,
		TokenHash: resetRequest.TokenHash,
		ExpiresAt: resetRequest.ExpiresAt,
		UsedAt: resetRequest.UsedAt,
		CreatedAt: resetRequest.CreatedAt,
	}

	if err := pr.DB.Create(resetRequestPayload).Error; err!=nil{
		return nil, err
	}
	return resetRequestPayload, nil
}