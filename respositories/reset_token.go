package repositories

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm"
)

func (pr *Repository) CreatePasswordReset(
	resetRequest *models.PasswordResetToken,
	) (*models.PasswordResetToken, error) {
	resetRequestPayload := &models.PasswordResetToken{
		ID:        utils.GenerateUUID(),
		UserID:    resetRequest.UserID,
		TokenHash: resetRequest.TokenHash,
		ExpiresAt: resetRequest.ExpiresAt,
		CreatedAt: resetRequest.CreatedAt,
	}

	if err := pr.DB.Create(resetRequestPayload).Error; err != nil {
		return nil, err
	}
	return resetRequestPayload, nil
}

func (pr *Repository) UpdatePasswordReset(
	resetRequest *models.PasswordResetToken,
) (*models.PasswordResetToken, error) {

	result := pr.DB.
		Model(&models.PasswordResetToken{}).
		Where("id = ?", resetRequest.ID).
		Updates(map[string]interface{}{
			"used_at": resetRequest.UsedAt,
		})

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var updatedResetToken models.PasswordResetToken

	if err := pr.DB.
		Where("id = ?", resetRequest.ID).
		First(&updatedResetToken).Error; err != nil {
		return nil, err
	}

	return &updatedResetToken, nil
}

func (pr *Repository) FindByToken(
	token string,
) (*models.PasswordResetToken, error) {

	tokenHash := utils.HashToken(token)

	var resetToken models.PasswordResetToken

	err := pr.DB.
		Where("token_hash = ?", tokenHash).
		First(&resetToken).Error

	if err != nil {
		return nil, err
	}

	return &resetToken, nil
}
