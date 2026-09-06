package services

import (
	"errors"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type WalletService struct {
	db *gorm.DB
	WalletRepo     *repositories.WalletRepository
	BankRepo       *repositories.BankRepository
	LinkedBankRepo *repositories.LinkedBankAccountRepository
	LedgerService *LedgerService
}

func NewWalletService(
		db *gorm.DB,
	WalletRepo *repositories.WalletRepository,
	BankRepo *repositories.BankRepository,
	LinkedBankRepo *repositories.LinkedBankAccountRepository,
	LedgerService *LedgerService,

) *WalletService {
	return &WalletService{
		db: db,
		WalletRepo:     WalletRepo,
		BankRepo:       BankRepo,
		LinkedBankRepo: LinkedBankRepo,
		LedgerService: LedgerService,
	}
}

func (ws *WalletService) CreateWalletForMerchant(merchantID uuid.UUID) (*models.Wallets, error) {
	wallet, walletErr := ws.WalletRepo.CreateWalletForMerchant(merchantID)
	if walletErr != nil {
		return nil, walletErr
	}
	return wallet, nil
}

func (ws *WalletService) GetWalletByMerchantID(merchantID uuid.UUID) (*models.Wallets, error) {
	wallet, errW := ws.WalletRepo.GetWalletByMerchantId(merchantID)
	if errW != nil {
		return nil, errW
	}
	return wallet, nil
}

func (ws *WalletService) UpdateWalletStatus(merchantId uuid.UUID, status string) (*models.Wallets, error) {
	updatedWallet, errUW := ws.WalletRepo.UpdateWalletStatus(merchantId, status)
	if errUW != nil {
		return nil, errUW
	}

	return updatedWallet, nil
}

func (ws *WalletService) AutoTopupAmount(
	wallet *models.Wallets,
	amount decimal.Decimal,
) (*models.Wallets, error) {

	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("amount should be greater than 0")
	}

	if wallet == nil {
		return nil, errors.New("wallet not found")
	}

	// 1. Get primary linked bank account
	linkedBankAccount, err := ws.LinkedBankRepo.
		GetPrimaryBankAccountLinkedByMerchantID(wallet.MerchantID)

	if err != nil {
		return nil, err
	}

	if linkedBankAccount == nil {
		return nil, errors.New("no primary linked bank account found")
	}

	// 2. Get bank account
	bankAccount, err := ws.BankRepo.
		GetBankAccountByID(linkedBankAccount.BankAccountID)

	if err != nil {
		return nil, err
	}

	if bankAccount == nil {
		return nil, errors.New("bank account not found")
	}

	// 3. Validate minimum bank balance
	availableBalance := bankAccount.Balance

	canDoTopup := availableBalance.
		Sub(amount).
		GreaterThan(decimal.NewFromInt(1000))

	if !canDoTopup {
		return nil, errors.New(
			"can't proceed with the topup as minimum balance constraints",
		)
	}

	// 4. Calculate new balances
	updatedWalletBalance := wallet.AvailableBalance.Add(amount)
	updatedBankBalance := bankAccount.Balance.Sub(amount)

	// 5. Start transaction
	tx := ws.db.Begin()

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// ------------------------------------
	// 6. Debit bank account
	// ------------------------------------

	updateBankAccountPayload := &dto.UpdateBankAccountRequest{
		Balance: updatedBankBalance,
	}

	updatedBankAccount, err := ws.BankRepo.
		WithTx(tx).
		UpdateBankAccount(
			linkedBankAccount.BankAccountID,
			updateBankAccountPayload,
		)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ------------------------------------
	// 7. Credit wallet
	// ------------------------------------

	updateWalletPayload := &dto.UpdateWallletBalanceRequest{
		AvailableBalance: updatedWalletBalance,
	}

	updatedWallet, err := ws.WalletRepo.
		WithTx(tx).
		UpdateWalletBalance(
			wallet.MerchantID,
			updateWalletPayload,
		)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ------------------------------------
	// 8. Create ledger transaction
	// ------------------------------------

	ledgerRequest := &dto.PostLedgerTransactionRequest{
		ReferenceID:      utils.GenerateAutoTopUpReference(),
		Type:             constants.LedgerTransactionTypeTopup,
		Currency:         "INR",
		Status:           "completed",
		SettlementStatus: "done",
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

	err = ws.LedgerService.PostTransaction(
		tx,
		ledgerRequest,
		ledgerEntries,
	)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ------------------------------------
	// 9. Commit
	// ------------------------------------

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return updatedWallet, nil
}