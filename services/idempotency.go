package services

import (
	"errors"
	"strings"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
)

func (is *Services) GetIdempotencyByKeyAndMerchantId(merchantID uuid.UUID, key string) (*models.IdempotencyKey, error) {
	if merchantID == uuid.Nil {
		return nil, errors.New("merchantID is required")
	}

	if key == "" {
		return nil, errors.New("key is required")
	}

	idempotency, errI := is.Repo.GetIdempotencyByKey(merchantID, key)
	if errI != nil {
		return nil, errI
	}

	return idempotency, nil
}

func (is *Services) CreateIdempotencyKey(
    merchantID uuid.UUID,
    key string,
) (*models.IdempotencyKey, error) {
    if merchantID == uuid.Nil {
        return nil, errors.New("merchantID is required")
    }

    key = strings.TrimSpace(key)
    if key == "" {
        return nil, errors.New("key is required")
    }

    idempotencyKey := &models.IdempotencyKey{
        ID:            utils.GenerateUUID(),
        MerchantID:   merchantID,
        IdempotencyKey: key,
    }

    return is.Repo.CreateIdempotencyKey(idempotencyKey)
}

func (is *Services) UpdateIdempotencyKey(
    idempotencyKey *models.IdempotencyKey,
) (*models.IdempotencyKey, error) {
    if idempotencyKey == nil {
        return nil, errors.New("idempotencyKey is required")
    }

    if idempotencyKey.ID == uuid.Nil {
        return nil, errors.New("idempotencyKey ID is required")
    }

    if idempotencyKey.ResponseBody == nil {
        return nil, errors.New("responseBody is required")
    }

    return is.Repo.UpdateIdempotencyKey(idempotencyKey)
}