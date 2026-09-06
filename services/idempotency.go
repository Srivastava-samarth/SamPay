package services

import (
	"errors"
	"strings"

	models "github.com/Srivastava-samarth/sampay/database/models"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
)

type IdempotencyService struct {
	IdempotencyRepo *repositories.IdempotencyRepository
}

func NewIdempotencyService(
	idempotencyRepo *repositories.IdempotencyRepository,
) *IdempotencyService {
	return &IdempotencyService{
		IdempotencyRepo: idempotencyRepo,
	}
}

func (is *IdempotencyService) GetIdempotencyByKeyAndMerchantId(merchantID uuid.UUID, key string) (*models.IdempotencyKey, error) {
	if merchantID == uuid.Nil {
		return nil, errors.New("merchantID is required")
	}

	if key == "" {
		return nil, errors.New("key is required")
	}

	idempotency, errI := is.IdempotencyRepo.GetIdempotencyByKey(merchantID, key)
	if errI != nil {
		return nil, errI
	}

	return idempotency, nil
}

func (is *IdempotencyService) CreateIdempotencyKey(merchantID uuid.UUID, key string, responseBody []byte) (*models.IdempotencyKey, error) {
	if merchantID == uuid.Nil {
		return nil, errors.New("merchantID is required")
	}

	if strings.TrimSpace(key) == "" {
		return nil, errors.New("key is required")
	}
	key = strings.TrimSpace(key)

	idempotencyKey := &models.IdempotencyKey{
		ID:             utils.GenerateUUID(),
		MerchantID:     merchantID,
		IdempotencyKey: key,
		ResponseBody:   responseBody,
	}

	idempotency, errI := is.IdempotencyRepo.CreateIdempotencyKey(idempotencyKey)
	if errI != nil {
		return nil, errI
	}

	return idempotency, nil
}
