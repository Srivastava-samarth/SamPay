package services

import (
	"errors"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LedgerService struct {
	LedgerRepo *repositories.LedgerRepository
	db         *gorm.DB
}

func NewLedgerService(
	ledgerRepo *repositories.LedgerRepository,
	db *gorm.DB,
) *LedgerService {
	return &LedgerService{
		LedgerRepo: ledgerRepo,
		db:         db,
	}
}

func (ls *LedgerService) GetTransactions(
	accountType *string,
	accountID uuid.UUID,
	cursor *uuid.UUID,
	direction string,
) ([]dto.LedgerTransactionRow, *dto.TransactionPagination, error) {

	if accountType == nil || *accountType == "" {
		return nil, nil, errors.New("account type is empty")
	}

	if accountID == uuid.Nil {
		return nil, nil, errors.New("account ID is empty")
	}

	if cursor != nil && *cursor == uuid.Nil {
		return nil, nil, errors.New("invalid cursor")
	}

	if direction != "next" && direction != "previous" {
		return nil, nil, errors.New("invalid pagination direction")
	}

	transactions, err := ls.LedgerRepo.GetWalletTransactions(
		accountType,
		accountID,
		cursor,
		direction,
	)

	if err != nil {
		return nil, nil, err
	}

	pagination := &dto.TransactionPagination{
		Records: len(transactions),
	}

	if direction == "previous" {
		reverseTransactions(transactions)
	}

	hasNextPage := len(transactions) > 10

	if hasNextPage {
		pagination.Next = &transactions[10].LedgerTransactionID
		transactions = transactions[:10]
	}

	if cursor != nil && len(transactions) > 0 {
		pagination.Previous = &transactions[0].LedgerTransactionID
	}

	pagination.Records = len(transactions)

	return transactions, pagination, nil
}

func reverseTransactions(
	transactions []dto.LedgerTransactionRow,
) {
	for i, j := 0, len(transactions)-1; i < j; i, j = i+1, j-1 {
		transactions[i], transactions[j] = transactions[j], transactions[i]
	}
}

func (ls *LedgerService) GetTransaction(
	walletId uuid.UUID,
	transactionId uuid.UUID,
) (*dto.LedgerTransactionRow, error) {
	if walletId == uuid.Nil {
		return nil, errors.New("WalletId missing")
	}

	if transactionId == uuid.Nil {
		return nil, errors.New("TransactionId missing")
	}

	transaction, errT := ls.LedgerRepo.GetWalletTransactionById(walletId, transactionId)
	if errT != nil {
		return nil, errT
	}

	return transaction, nil
}

func (ls *LedgerService) CreateLedgerEntries(
	tx *gorm.DB,
	entries []*models.LedgerEntry,
) (int, error) {

	if len(entries) == 0 {
		return 0, errors.New("no entries were provided")
	}

	if tx == nil {
		return 0, errors.New("transaction cannot be nil")
	}

	txLedgerRepo := ls.LedgerRepo.WithTx(tx)

	for _, entry := range entries {
		if entry == nil {
			return 0, errors.New("ledger entry cannot be nil")
		}

		if entry.LedgerTransactionID == uuid.Nil {
			return 0, errors.New("ledger transaction ID is missing")
		}
	}

	for _, entry := range entries {
		if _, err := txLedgerRepo.CreateLedgerEntry(entry); err != nil {
			return 0, err
		}
	}

	return len(entries), nil
}

func (ls *LedgerService) CreateLedgerTransaction(
	transaction *models.LedgerTransaction,
) (*models.LedgerTransaction, error) {
	if transaction == nil {
		return nil, errors.New("transaction not found")
	}

	ledgerTransaction, errLT := ls.LedgerRepo.CreateLedgerTransaction(transaction)
	if errLT != nil {
		return nil, errLT
	}

	return ledgerTransaction, nil
}

func (ls *LedgerService) PostTransaction(
	tx *gorm.DB,
	request *dto.PostLedgerTransactionRequest,
	entries []*models.LedgerEntry,
) (*models.LedgerTransaction, error) {

	if tx == nil {
		return nil, errors.New("transaction cannot be nil")
	}

	if request == nil {
		return nil, errors.New("transaction request cannot be nil")
	}

	if request.ReferenceID == "" {
		return nil, errors.New("reference ID is required")
	}

	if request.Currency != "INR" {
		return nil, errors.New("currency must be INR")
	}

	if len(entries) == 0 {
		return nil, errors.New("no ledger entries were provided")
	}

	for _, entry := range entries {
		if entry == nil {
			return nil, errors.New("ledger entry cannot be nil")
		}
	}

	txLedgerRepo := ls.LedgerRepo.WithTx(tx)

	exists, err := txLedgerRepo.
		ExistingLedgerTransactionByReferenceID(request.ReferenceID)

	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New(
			"ledger transaction with this reference already exists",
		)
	}

	ledgerTransactionPayload := &models.LedgerTransaction{
		Type:             request.Type,
		ReferenceID:      request.ReferenceID,
		Status:           request.Status,
		SettlementStatus: request.SettlementStatus,
	}

	ledgerTransaction, err := txLedgerRepo.CreateLedgerTransaction(
		ledgerTransactionPayload,
	)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {

		entry.LedgerTransactionID = ledgerTransaction.ID

		_, err = txLedgerRepo.CreateLedgerEntry(entry)
		if err != nil {
			return nil, err
		}
	}

	return ledgerTransaction, nil
}

func (ls *LedgerService) GetLedgerTransactionByReferenceID(referenceID string) (*models.LedgerTransaction, error) {
	if referenceID == "" {
		return nil, errors.New("reference id is required")
	}

	ledgerTransaction, errLT := ls.LedgerRepo.GetLedgerTransactionByReferenceID(referenceID)
	if errLT != nil {
		return nil, errLT
	}

	if ledgerTransaction == nil {
		return nil, errors.New("transaction does not exist")
	}

	return ledgerTransaction, nil
}
