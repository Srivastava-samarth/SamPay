package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PayoutRepository struct {
	DB *gorm.DB
}

func NewPayoutRepository(
	db *gorm.DB,
) *PayoutRepository {
	return &PayoutRepository{
		DB: db,
	}
}

func (pr *PayoutRepository) WithTx(tx *gorm.DB) *PayoutRepository {
	return &PayoutRepository{
		DB: tx,
	}
}

func (pr *PayoutRepository) CreatePayoutWalletToBank(merchantID uuid.UUID, request *dto.CreateWalletToBankRequest, status string) (*models.Payout, error) {
	createPayoutPayload := &models.Payout{
		ID:                       utils.GenerateUUID(),
		MerchantID:               merchantID,
		SourceWalletID:           &request.SenderWalletID,
		SourceBankAccountID:      nil,
		DestinationBankAccountID: request.DestinationBankAccountID,
		PayoutReference:          *utils.GeneratePayoutReference(),
		Amount:                   request.Amount,
		Currency:                 request.Currency,
		Description:              request.Description,
		ExternalReference:        utils.GenerateCustomerReference(),
		Status:                   status,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	err := pr.DB.Create(&createPayoutPayload).Error
	if err != nil {
		return nil, err
	}

	return createPayoutPayload, nil
}

func (pr *PayoutRepository) CreatePayoutBankToBank(merchantID uuid.UUID, request *dto.CreateBankToBankRequest, status string) (*models.Payout, error) {
	createPayoutPayload := &models.Payout{
		ID:                       utils.GenerateUUID(),
		MerchantID:               merchantID,
		SourceWalletID:           nil,
		SourceBankAccountID:      &request.SourceBankAccountID,
		DestinationBankAccountID: request.DestinationBankAccountID,
		PayoutReference:          *utils.GeneratePayoutReference(),
		Amount:                   request.Amount,
		Currency:                 request.Currency,
		Description:              request.Description,
		ExternalReference:        utils.GenerateCustomerReference(),
		Status:                   status,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	err := pr.DB.Create(&createPayoutPayload).Error
	if err != nil {
		return nil, err
	}

	return createPayoutPayload, nil
}

func (pr *PayoutRepository) UpdatePayoutStatus(PayoutReference string, status string) (*models.Payout, error) {
	var payout *models.Payout
	err := pr.DB.Where("payout_reference = ?", PayoutReference).First(&payout).Error
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if status == payout.Status {
		return payout, nil
	}

	updates["status"] = status

	errU := pr.DB.
		Model(&models.Payout{}).
		Where("id = ?", payout.ID).
		Updates(updates).Error

	if errU != nil {
		return nil, errU
	}

	var updatedPayout *models.Payout
	errF := pr.DB.Where("id = ?", payout.ID).First(&updatedPayout).Error
	if errF != nil {
		return nil, errF
	}
	return updatedPayout, nil

}
