package repositories

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm"
)

type PasswordResetTokenRepository struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(
	db *gorm.DB,
) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{
		db: db,
	}
}

func (wr *PasswordResetTokenRepository) WithTx(tx *gorm.DB) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{
		db: tx,
	}
}

func (pr *PasswordResetTokenRepository) CreatePasswordReset(resetRequest *models.PasswordResetToken) (*models.PasswordResetToken, error) {
	resetRequestPayload := &models.PasswordResetToken{
		ID:        utils.GenerateUUID(),
		UserID:    resetRequest.UserID,
		TokenHash: resetRequest.TokenHash,
		ExpiresAt: resetRequest.ExpiresAt,
		CreatedAt: resetRequest.CreatedAt,
	}

	if err := pr.db.Create(resetRequestPayload).Error; err != nil {
		return nil, err
	}
	return resetRequestPayload, nil
}

func (pr *PasswordResetTokenRepository) UpdatePasswordReset(
	resetRequest *models.PasswordResetToken,
) (*models.PasswordResetToken, error) {

	result := pr.db.
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

	if err := pr.db.
		Where("id = ?", resetRequest.ID).
		First(&updatedResetToken).Error; err != nil {
		return nil, err
	}

	return &updatedResetToken, nil
}

func (pr *PasswordResetTokenRepository) FindByToken(
	token string,
) (*models.PasswordResetToken, error) {

	tokenHash := utils.HashToken(token)

	var resetToken models.PasswordResetToken

	err := pr.db.
		Where("token_hash = ?", tokenHash).
		First(&resetToken).Error

	if err != nil {
		return nil, err
	}

	return &resetToken, nil
}
