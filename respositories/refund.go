package repositories

import (
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefundRepository struct {
	DB *gorm.DB
}

func NewRefundRepository(
	db *gorm.DB,
) *RefundRepository {
	return &RefundRepository{
		DB: db,
	}
}

func (pr *RefundRepository) WithTx(tx *gorm.DB) *RefundRepository {
	return &RefundRepository{
		DB: tx,
	}
}

func (rr *RefundRepository) CreateRefund(request *dto.CreateRefundRequest) (*models.Refund, error) {
	createRefundPayload := &models.Refund{
		ID:                utils.GenerateUUID(),
		PaymentID:         request.PaymentID,
		MerchantID:        request.MerchantID,
		ExternalReference: request.ExternalReference,
		RefundReference:   *utils.GenerateRefundReference(),
		Amount:            request.Amount,
		Currency:          request.Currency,
		Reason:            request.Reason,
		Status:            constants.TransactionStatusProcessing,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err := rr.DB.Create(&createRefundPayload).Error
	if err != nil {
		return nil, err
	}

	return createRefundPayload, nil
}

func (rr *RefundRepository) GetRefundsByPaymentID(
	paymentID uuid.UUID,
) ([]*models.Refund, error) {

	var refunds []*models.Refund

	err := rr.DB.
		Where("payment_id = ?", paymentID).
		Order("created_at DESC").
		Find(&refunds).Error

	if err != nil {
		return nil, err
	}

	return refunds, nil
}

func (rr *RefundRepository) GetRefundByReference(
	refundRef string,
) (*models.Refund, error) {
	var refund *models.Refund

	err := rr.DB.
		Where("refund_reference = ?", refundRef).
		Order("created_at DESC").
		Find(&refund).Error

	if err != nil {
		return nil, err
	}

	return refund, nil
}

func (rr *RefundRepository) UpdateRefundStatus(RefundReference string, status string) (*models.Refund, error) {
	var refund *models.Refund
	err := rr.DB.Where("refund_reference = ?", RefundReference).First(&refund).Error
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if status == refund.Status {
		return refund, nil
	}

	updates["status"] = status

	errU := rr.DB.
		Model(&models.Refund{}).
		Where("id = ?", refund.ID).
		Updates(updates).Error

	if errU != nil {
		return nil, errU
	}

	var updatedRefund *models.Refund
	errF := rr.DB.Where("id = ?", refund.ID).First(&updatedRefund).Error
	if errF != nil {
		return nil, errF
	}
	return updatedRefund, nil

}

func (rr *RefundRepository) GetRefundById(
	refundID uuid.UUID,
) (*models.Refund, error) {
	var refund *models.Refund

	err := rr.DB.
		Where("id = ?", refundID).
		Order("created_at DESC").
		Find(&refund).Error

	if err != nil {
		return nil, err
	}

	return refund, nil
}

func (rr *RefundRepository) GetRefundsByMerchantId(
	merchantID uuid.UUID,
) ([]*models.Refund, error) {

	var refunds []*models.Refund

	err := rr.DB.
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Find(&refunds).Error

	if err != nil {
		return nil, err
	}

	return refunds, nil
}
