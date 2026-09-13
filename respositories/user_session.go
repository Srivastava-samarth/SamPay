package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserSessionRepository struct {
	db *gorm.DB
}

func NewUserSessionRepository(
	db *gorm.DB,
) *UserSessionRepository {
	return &UserSessionRepository{
		db: db,
	}
}

func (wr *UserSessionRepository) WithTx(tx *gorm.DB) *UserSessionRepository {
	return &UserSessionRepository{
		db: tx,
	}
}

func (ar *UserSessionRepository) CreateUserSession(oauth *models.UserSession) (*models.UserSession, error) {
	userSessionRequest := &models.UserSession{
		ID:               utils.GenerateUUID(),
		UserID:           oauth.UserID,
		RefreshTokenHash: oauth.RefreshTokenHash,
		ExpiresAt:        oauth.ExpiresAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := ar.db.Create(userSessionRequest).Error; err != nil {
		return nil, err
	}
	return userSessionRequest, nil
}

func (ar *UserSessionRepository) GetUserSessionByRefreshTokenHash(refreshTokenHash string) (*models.UserSession, error) {
	var userSession *models.UserSession
	err := ar.db.
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
