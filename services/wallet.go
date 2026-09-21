package services

import (
	"errors"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (ws *Services) CreateWalletForMerchant(merchantID uuid.UUID) (*models.Wallets, error) {
	if merchantID == uuid.Nil {
		return nil, errors.New("merchant id is required")
	}

	existingWallet, errEW := ws.Repo.GetWalletByMerchantId(merchantID)
	if errEW != nil {
		return nil, errEW
	}

	if existingWallet != nil {
		return nil, errors.New("wallet already exists for this merchant")
	}

	wallet, walletErr := ws.Repo.CreateWalletForMerchant(merchantID)
	if walletErr != nil {
		return nil, walletErr
	}
	return wallet, nil
}

func (ws *Services) GetWalletByMerchantID(merchantID uuid.UUID) (*models.Wallets, error) {
	if merchantID == uuid.Nil {
		return nil, errors.New("merchant id is required")
	}

	wallet, errW := ws.Repo.GetWalletByMerchantId(merchantID)
	if errW != nil {
		return nil, errW
	}
	return wallet, nil
}

func (ws *Services) UpdateWalletStatus(merchantId uuid.UUID, status string) (*models.Wallets, error) {
	if merchantId == uuid.Nil {
		return nil, errors.New("merchant id is required")
	}
	updatedWallet, errUW := ws.Repo.UpdateWalletStatus(merchantId, status)
	if errUW != nil {
		return nil, errUW
	}

	return updatedWallet, nil
}

func (ws *Services) TopUpWalletFromPrimaryBank(
	tx *gorm.DB,
	wallet *models.Wallets,
	amount decimal.Decimal,
) (*models.Wallets, error) {

	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("amount should be greater than 0")
	}

	if wallet == nil {
		return nil, errors.New("wallet not found")
	}

	// Use transaction-aware repositories.
	repo := ws.Repo.WithTx(tx)

	linkedBankAccount, err := repo.
		GetPrimaryBankAccountLinkedByMerchantID(wallet.MerchantID)
	if err != nil {
		return nil, err
	}

	if linkedBankAccount == nil {
		return nil, errors.New("no primary linked bank account found")
	}

	bankAccount, err := repo.
		GetBankAccountByID(linkedBankAccount.BankAccountID)
	if err != nil {
		return nil, err
	}

	if bankAccount == nil {
		return nil, errors.New("bank account not found")
	}

	// Merchant's bank account must retain at least ₹1000.
	minimumBalance := decimal.NewFromInt(1000)

	remainingBankBalance := bankAccount.Balance.Sub(amount)

	if remainingBankBalance.LessThan(minimumBalance) {
		return nil, errors.New(
			"can't proceed with the topup as minimum balance constraint",
		)
	}

	updatedBankAccount, err := repo.UpdateBankAccount(
		linkedBankAccount.BankAccountID,
		&dto.UpdateBankAccountRequest{
			Balance: remainingBankBalance,
		},
	)
	if err != nil {
		return nil, err
	}

	updatedAvalBal := wallet.AvailableBalance.Add(amount)
	updatedWallet, err := repo.UpdateWalletBalance(
		wallet.MerchantID,
		&dto.UpdateWalletBalanceRequest{
			AvailableBalance: &updatedAvalBal,
		},
	)
	if err != nil {
		return nil, err
	}

	ledgerRequest := &dto.PostLedgerTransactionRequest{
		ReferenceID:      utils.GenerateAutoTopUpReference(),
		Type:             constants.LedgerTransactionTypeTopup,
		Currency:         "INR",
		Status:           constants.TransactionStatusCompleted,
		SettlementStatus: constants.LedgerSettlementSettled,
	}

	ledgerEntries := []*models.LedgerEntry{
		{
			AccountType: constants.LedgerAccountTypeBankAccount,
			AccountID:   updatedBankAccount.ID,
			EntryType:   constants.LedgerEntryTypeDebit,
			Amount:      amount,
			Currency:    "INR",
		},
		{
			AccountType: constants.LedgerAccountTypeWallet,
			AccountID:   updatedWallet.ID,
			EntryType:   constants.LedgerEntryTypeCredit,
			Amount:      amount,
			Currency:    "INR",
		},
	}

	_, err = ws.PostTransaction(
		tx,
		ledgerRequest,
		ledgerEntries,
	)
	if err != nil {
		return nil, err
	}

	return updatedWallet, nil
}
