package repositories

import (
	"errors"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (ir *Repository) GetIdempotencyByKey(merchantID uuid.UUID, key string) (*models.IdempotencyKey, error) {
	var idempotency *models.IdempotencyKey
	err := ir.DB.Where("merchant_id = ? AND idempotency_key = ?", merchantID, key).First(&idempotency).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return idempotency, nil
}

func (ir *Repository) CreateIdempotencyKey(idempotencyKey *models.IdempotencyKey) (*models.IdempotencyKey, error) {
	err := ir.DB.Create(idempotencyKey).Error
	if err != nil {
		return nil, err
	}

	return idempotencyKey, nil
}

func (ir *Repository) UpdateIdempotencyKey(
    idempotencyKey *models.IdempotencyKey,
) (*models.IdempotencyKey, error) {
    err := ir.DB.
        Model(&models.IdempotencyKey{}).
        Where("id = ?", idempotencyKey.ID).
        Updates(map[string]interface{}{
            "response_body": idempotencyKey.ResponseBody,
        }).Error

    if err != nil {
        return nil, err
    }

    return idempotencyKey, nil
}
