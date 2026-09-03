package services

import (
	"errors"

	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/google/uuid"
)

type LedgerService struct {
	ledgerRepo *repositories.LedgerRepository
}

func NewLedgerService(
	ledgerRepo *repositories.LedgerRepository,
) *LedgerService {
	return &LedgerService{
		ledgerRepo: ledgerRepo,
	}
}

const transactionPageSize = 10

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

    transactions, err := ls.ledgerRepo.GetWalletTransactions(
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

func (ls *LedgerService) GeTransaction(
	walletId uuid.UUID,
	transactionId uuid.UUID,
) (*dto.LedgerTransactionRow, error){
	if walletId == uuid.Nil{
		return nil, errors.New("WalletId missing")
	}

	if transactionId == uuid.Nil{
		return nil, errors.New("TransactionId missing")
	}

	transaction, errT := ls.ledgerRepo.GetWalletTransactionById(walletId,transactionId)
	if errT != nil{
		return nil,errT
	}

	return transaction, nil
}