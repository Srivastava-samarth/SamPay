package repositories

import (
	"errors"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IdempotencyRepository struct {
	db *gorm.DB
}

func NewIdempotencyRepository(
	db *gorm.DB,
) *IdempotencyRepository{
	return &IdempotencyRepository{
		db:db,
	}
}

func (ir *IdempotencyRepository) GetIdempotencyByKey(merchantID uuid.UUID, key string) (*models.IdempotencyKey, error){
	var idempotency *models.IdempotencyKey
	err := ir.db.Where("merchant_id = ? AND idempotency_key = ?", merchantID, key).First(&idempotency).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil{
		return nil, err
	}

	return idempotency, nil
}

func (ir *IdempotencyRepository) CreateIdempotencyKey(idempotencyKey *models.IdempotencyKey) (*models.IdempotencyKey, error){
	err := ir.db.Create(idempotencyKey).Error
	if err != nil{
		return nil, err
	}

	return idempotencyKey, nil
}