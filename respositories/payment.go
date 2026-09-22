package repositories

import (
	"errors"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (pr *Repository) CreatePayment(merchantID uuid.UUID, request *dto.CreatePaymentRequest, status *string) (*models.Payment, error) {
	settlementStatus := constants.LedgerSettlementPending
	createPaymentPayload := &models.Payment{
		ID:                 utils.GenerateUUID(),
		PaymentReference:   utils.GeneratePaymentReference(),
		SenderMerchantID:   merchantID,
		ReceiverMerchantID: request.ReceiverMerchantID,
		Amount:             request.Amount,
		Currency:           &request.Currency,
		Description:        &request.Description,
		CustomerReference:  utils.GenerateCustomerReference(),
		Status:             status,
		SettlementStatus:   &settlementStatus,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := pr.DB.Create(&createPaymentPayload).Error
	if err != nil {
		return nil, err
	}

	return createPaymentPayload, nil
}

func (pr *Repository) UpdatePaymentStatus(
	paymentReference string,
	status *string,
) (*models.Payment, error) {

	var payment models.Payment

	err := pr.DB.
		Where("payment_reference = ?", paymentReference).
		First(&payment).Error

	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if *status == *payment.Status {
	return &payment, nil
}

	updates["status"] = status

	err = pr.DB.
		Model(&models.Payment{}).
		Where("id = ?", payment.ID).
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

	var updatedPayment models.Payment

	err = pr.DB.
		Where("id = ?", payment.ID).
		First(&updatedPayment).Error

	if err != nil {
		return nil, err
	}

	return &updatedPayment, nil
}

func (pr *Repository) UpdateSettlementStatusByID(
	paymentID uuid.UUID,
	settlementStatus *string,
) (*models.Payment, error) {

	updates := map[string]interface{}{
		"settlement_status": settlementStatus,
		"updated_at":        time.Now(),
	}

	errU := pr.DB.
		Model(&models.Payment{}).
		Where("id = ?", paymentID).
		Updates(updates).Error

	if errU != nil {
		return nil, errU
	}

	var payment *models.Payment
	errP := pr.DB.Where("id = ?", paymentID).First(&payment).Error
	if errP != nil {
		return nil, errP
	}

	return payment, nil
}

func (pr *Repository) GetPaymentByID(paymentID uuid.UUID) (*models.Payment, error) {
	var payment *models.Payment
	err := pr.DB.Where("id = ?", paymentID).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return payment, nil
}

func (pr *Repository) GetPaymentsBySettlementStatus(status string) ([]*models.Payment, error) {
	var payments []*models.Payment
	err := pr.DB.Where("settlement_status = ? AND status = ?", status, constants.TransactionStatusCompleted).Find(&payments).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return payments, nil
}

func (pr *Repository) GetPaymentsByMerchantID(merchantID uuid.UUID) ([]*models.Payment, error) {
	var payments []*models.Payment
	err := pr.DB.Where("sender_merchant_id = ?", merchantID).Find(&payments).Error
	if err != nil {
		return nil, err
	}

	return payments, nil
}
