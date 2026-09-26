package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm/clause"
)

func (ar *Repository) CreateUserSession(
	oauth *models.UserSession,
	) (*models.UserSession, error) {
	userSessionRequest := &models.UserSession{
		ID:               utils.GenerateUUID(),
		UserID:           oauth.UserID,
		RefreshTokenHash: oauth.RefreshTokenHash,
		ExpiresAt:        oauth.ExpiresAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := ar.DB.Create(userSessionRequest).Error; err != nil {
		return nil, err
	}
	return userSessionRequest, nil
}

func (ar *Repository) GetUserSessionByRefreshTokenHash(
	refreshTokenHash string,
	) (*models.UserSession, error) {
	var userSession *models.UserSession
	err := ar.DB.
		Clauses(clause.Locking{
			Strength: "UPDATE",
		}).
		Where("refresh_token_hash = ?", refreshTokenHash).
		First(&userSession).Error

	if err != nil {
		return nil, err
	}
	return userSession, nil
}
