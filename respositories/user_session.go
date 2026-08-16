package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm"
)

type UserSessionRepository struct {
	db *gorm.DB
}

func NewUserSessionRepository(
	db *gorm.DB,
) *UserSessionRepository{
	return &UserSessionRepository{
		db: db,
	}
}

func(ar *UserSessionRepository) CreateUserSession(oauth *models.UserSession) (*models.UserSession, error){
	userSessionRequest := &models.UserSession{
		ID: utils.GenerateUUID(),
		UserID: oauth.UserID,
		RefreshTokenHash: oauth.RefreshTokenHash,
		ExpiresAt: oauth.ExpiresAt,
		CreatedAt: time.Now(),
		RevokedAt: nil,
		UpdatedAt: time.Now(),
	}
	if err := ar.db.Create(userSessionRequest).Error; err != nil {
		return nil, err
	}
	return userSessionRequest, nil
}