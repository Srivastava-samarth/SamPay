package services

import (
	"errors"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const MinimumBankBalance = 1000

func (ps *Services) ValidatePayoutRequestWalletToBank(
	request *dto.CreateWalletToBankRequest, 
	senderMerchantID uuid.UUID,
	) error {
	if request == nil {
		return errors.New("payment request is required")
	}

	if senderMerchantID == uuid.Nil {
		return errors.New("sender merchant id is required")
	}

	if request.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be greater than 0")
	}

	if request.Currency != "INR" {
		return errors.New("currency should be INR")
	}

	senderWallet, errSW := ps.Repo.GetWalletByMerchantId(senderMerchantID)
	if errSW != nil {
		return errSW
	}

	if request.SenderWalletID == uuid.Nil || senderWallet.ID != request.SenderWalletID {
		return errors.New("sender_wallet_id issue")
	}

	destinationBankAccount, errDBA := ps.Repo.GetBankAccountByID(request.DestinationBankAccountID)
	if errDBA != nil {
		return errDBA
	}

	if destinationBankAccount == nil {
		return errors.New("destination bank account not found")
	}
	return nil
}

func (ps *Services) ValidatePayoutRequestBankToBank(
	request *dto.CreateBankToBankRequest, 
	senderMerchantID uuid.UUID,
	) error {
	if request == nil {
		return errors.New("payment request is required")
	}

	if senderMerchantID == uuid.Nil {
		return errors.New("sender merchant id is required")
	}

	if request.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be greater than 0")
	}

	if request.Currency != "INR" {
		return errors.New("currency should be INR")
	}

	senderBankAccount, errSBA := ps.Repo.GetBankAccountByID(request.SourceBankAccountID)
	if errSBA != nil {
		return errSBA
	}

	linkedSenderBankAccount, errLSBA := ps.Repo.GetBankAccountLinkedByID(senderBankAccount.ID)
	if errLSBA != nil {
		return errLSBA
	}

	if senderMerchantID != linkedSenderBankAccount.MerchantID {
		return errors.New("merchant id doesn't match")
	}

	destinationBankAccount, errDBA := ps.Repo.GetBankAccountByID(request.DestinationBankAccountID)
	if errDBA != nil {
		return errDBA
	}

	if destinationBankAccount == nil {
		return errors.New("destination bank account not found")
	}

	if senderBankAccount.ID == destinationBankAccount.ID {
		return errors.New("accounts must be differnt")
	}
	return nil
}

func (ps *Services) CheckBankBalance(
	ID uuid.UUID,
	amount decimal.Decimal,
) (bool, error) {

	bankAccount, errBA := ps.Repo.GetBankAccountByID(ID)

	if errBA != nil {
		return false, errBA
	}

	if bankAccount == nil {
		return false, errors.New("bank account does not exist")
	}

	minimumBalance := decimal.NewFromInt(MinimumBankBalance)

	return bankAccount.Balance.GreaterThanOrEqual(
		amount.Add(minimumBalance),
	), nil
}

func (ps *Services) GetPayoutsByMerchantID(
	merchantID uuid.UUID,
) ([]*models.Payout, error) {
	if merchantID == uuid.Nil {
		return nil, errors.New("merchant_id is required")
	}

	payouts, errP := ps.Repo.GetPayoutsByMerchantID(merchantID)
	if errP != nil {
		return nil, errP
	}

	return payouts, nil
}

func (ps *Services) GetPayoutByID(
	payoutID uuid.UUID,
) (*models.Payout, error) {
	if payoutID == uuid.Nil {
		return nil, errors.New("payout_id is required")
	}

	payout, errP := ps.Repo.GetPayoutByID(payoutID)
	if errP != nil {
		return nil, errP
	}

	return payout, nil
}
