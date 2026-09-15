package services

import (
	"errors"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (ps *Services) ValidatePaymentRequest(
	request *dto.CreatePaymentRequest,
	senderMerchantID uuid.UUID,
) error {

	if request == nil {
		return errors.New("payment request is required")
	}

	if senderMerchantID == uuid.Nil {
		return errors.New("sender merchant id is required")
	}

	if request.ReceiverMerchantID == uuid.Nil {
		return errors.New("receiver merchant id is required")
	}

	if senderMerchantID == request.ReceiverMerchantID {
		return errors.New("sender and receiver can't be same")
	}

	if request.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be greater than 0")
	}

	if request.Currency != "INR" {
		return errors.New("currency should be INR")
	}

	senderMerchant, err := ps.Repo.GetMerchantByID(senderMerchantID)
	if err != nil {
		return err
	}

	if senderMerchant == nil {
		return errors.New("sender does not exist")
	}

	senderWallet, err := ps.GetWalletByMerchantID(senderMerchantID)
	if err != nil {
		return err
	}

	if senderWallet == nil {
		return errors.New("sender wallet not found")
	}

	receiverMerchant, err := ps.Repo.GetMerchantByID(
		request.ReceiverMerchantID,
	)
	if err != nil {
		return err
	}

	if receiverMerchant == nil {
		return errors.New("receiver does not exist")
	}

	receiverWallet, err := ps.GetWalletByMerchantID(
		request.ReceiverMerchantID,
	)
	if err != nil {
		return err
	}

	if receiverWallet == nil {
		return errors.New("receiver wallet does not exist")
	}

	if senderWallet.ID == receiverWallet.ID {
		return errors.New("sender and receiver can't be same")
	}

	return nil
}

func (ps *Services) CalculateFees(amount decimal.Decimal) (decimal.Decimal, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, errors.New("amount must be greater than 0")
	}

	feeRate := decimal.NewFromInt(1).Div(decimal.NewFromInt(100))
	fee := amount.Mul(feeRate)

	return fee, nil
}

func (ps *Services) CheckBalance(merchantID uuid.UUID, amount decimal.Decimal) (bool, error) {
	wallet, errW := ps.GetWalletByMerchantID(merchantID)
	if errW != nil {
		return false, errW
	}

	if wallet == nil {
		return false, errors.New("wallet does not exist")
	}
	return wallet.AvailableBalance.GreaterThanOrEqual(amount), nil
}

func (ps *Services) GetPaymentsByMerchantID(merchantID uuid.UUID) ([]*models.Payment, error) {
	if merchantID == uuid.Nil {
		return nil, errors.New("merchant_id is required")
	}

	payments, errP := ps.Repo.GetPaymentsByMerchantID(merchantID)
	if errP != nil {
		return nil, errP
	}

	return payments, nil
}

func (ps *Services) GetPaymentByID(paymentID uuid.UUID) (*models.Payment, error) {
	if paymentID == uuid.Nil {
		return nil, errors.New("payment_id is required")
	}

	payment, errP := ps.Repo.GetPaymentByID(paymentID)
	if errP != nil {
		return nil, errP
	}

	return payment, nil
}
