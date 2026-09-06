package services

import (
	"errors"

	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentService struct {
	PaymentRepo *repositories.PaymentRepository
	MerchantRepo *repositories.MerchantRepository
	walletSrvc *WalletService
}

func NewPaymentService(
	paymentRepo *repositories.PaymentRepository,
	merchantRepo *repositories.MerchantRepository,
		walletSrvc *WalletService,
) *PaymentService{
	return &PaymentService{
		PaymentRepo: paymentRepo,
		MerchantRepo: merchantRepo,
		walletSrvc: walletSrvc,
	}
}

func (ps *PaymentService) ValidatePaymentRequest(
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

    senderMerchant, err := ps.MerchantRepo.GetMerchantByID(senderMerchantID)
    if err != nil {
        return err
    }

    if senderMerchant == nil {
        return errors.New("sender does not exist")
    }

    senderWallet, err := ps.walletSrvc.GetWalletByMerchantID(senderMerchantID)
    if err != nil {
        return err
    }

    if senderWallet == nil {
        return errors.New("sender wallet not found")
    }

    receiverMerchant, err := ps.MerchantRepo.GetMerchantByID(
        request.ReceiverMerchantID,
    )
    if err != nil {
        return err
    }

    if receiverMerchant == nil {
        return errors.New("receiver does not exist")
    }

    receiverWallet, err := ps.walletSrvc.GetWalletByMerchantID(
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

func (ps *PaymentService) CalculateFees(amount decimal.Decimal) (decimal.Decimal, error) {
    if amount.LessThanOrEqual(decimal.Zero) {
        return decimal.Zero, errors.New("amount must be greater than 0")
    }

    feeRate := decimal.NewFromInt(1).Div(decimal.NewFromInt(100))
    fee := amount.Mul(feeRate)

    return fee, nil
}

func (ps *PaymentService) CheckBalance(merchantID uuid.UUID, amount decimal.Decimal) (bool,error){
	wallet, errW := ps.walletSrvc.GetWalletByMerchantID(merchantID)
	if errW != nil{
		return false, errW
	}

	if wallet == nil{
		return false, errors.New("wallet does not exist")
	}
	return true, nil
}
