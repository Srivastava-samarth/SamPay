package services

import (
	"errors"
	"strings"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type RefundService struct {
	RefundRepo  *repositories.RefundRepository
	PaymentRepo *repositories.PaymentRepository
}

func NewRefundService(
	refundRepo *repositories.RefundRepository,
	paymentRepo *repositories.PaymentRepository,
) *RefundService {
	return &RefundService{
		RefundRepo:  refundRepo,
		PaymentRepo: paymentRepo,
	}
}

func (rs *RefundService) ValidateCreateRefundRequest(
	request *dto.CreateRefundRequest,
) error {
	if request.PaymentID == uuid.Nil {
		return errors.New("payment id not provided")
	}

	if request.MerchantID == uuid.Nil {
		return errors.New("merchant id not provided")
	}

	if request.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be greater than 0")
	}

	if request.Reason == nil || strings.TrimSpace(*request.Reason) == "" {
		return errors.New("reason is necessary")
	}

	payment, err := rs.PaymentRepo.GetPaymentByID(request.PaymentID)
	if err != nil {
		return err
	}

	if payment == nil {
		return errors.New("payment not found")
	}
	if payment.SenderMerchantID != request.MerchantID {
		return errors.New("payment doesn't belong to the merchant")
	}

	if request.Currency == "" {
		return errors.New("currency is required")
	}

	if request.Currency != *payment.Currency {
		return errors.New("refund currency must match payment currency")
	}

	refunds, err := rs.RefundRepo.GetRefundsByPaymentID(request.PaymentID)
	if err != nil {
		return err
	}

	refundedAmount := decimal.Zero

	for _, refund := range refunds {
		if refund == nil {
			continue
		}

		if refund.Status != "" &&
			refund.Status == constants.TransactionStatusRefunded {
			refundedAmount = refundedAmount.Add(refund.Amount)
		}
	}

	refundableAmount := payment.Amount.Sub(refundedAmount)

	if request.Amount.GreaterThan(refundableAmount) {
		return errors.New("refund amount exceeds refundable amount")
	}

	return nil
}

func (rs *RefundService) GetRefundByReference(
	refundRef string,
) (*models.Refund, error) {
	refund, errR := rs.RefundRepo.GetRefundByReference(refundRef)
	if errR != nil {
		return nil, errR
	}

	if refund == nil {
		return nil, errors.New("refund not found")
	}

	return refund, nil
}

func (rs *RefundService) GetRefundById(
	refundID uuid.UUID,
) (*models.Refund, error) {
	refund, errR := rs.RefundRepo.GetRefundById(refundID)
	if errR != nil {
		return nil, errR
	}

	return refund, nil
}

func (rs *RefundService) GetRefundsByMerchantId(
	merchantID uuid.UUID,
) ([]*models.Refund, error) {
	refunds, errR := rs.RefundRepo.GetRefundsByMerchantId(merchantID)
	if errR != nil {
		return nil, errR
	}

	return refunds, nil
}
