package repositories

import (
	"errors"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LedgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(
	db *gorm.DB,
) *LedgerRepository {
	return &LedgerRepository{
		db: db,
	}
}

func (wr *LedgerRepository) WithTx(tx *gorm.DB) *LedgerRepository {
	return &LedgerRepository{
		db: tx,
	}
}


func (lr *LedgerRepository) GetWalletTransactions(
    accountType *string,
    accountID uuid.UUID,
    cursor *uuid.UUID,
    direction string,
) ([]dto.LedgerTransactionRow, error) {

    var transactions []dto.LedgerTransactionRow

    query := lr.db.
        Table("ledger_entries AS le").
        Select(`
            lt.id AS ledger_transaction_id,
            lt.transaction_ref,
            lt.type,
            lt.reference_id,
            lt.status,
            lt.settlement_status,
            le.amount,
            le.currency,
            le.entry_type,
            lt.created_at
        `).
        Joins(`
            JOIN ledger_transactions AS lt
                ON lt.id = le.ledger_transaction_id
        `).
        Where(
            "le.account_type = ? AND le.account_id = ?",
            *accountType,
            accountID,
        )

    if cursor == nil {
        query = query.
            Order("lt.created_at DESC, lt.id DESC").
            Limit(11)
    } else {

        var cursorCreatedAt time.Time

        err := lr.db.
            Table("ledger_transactions").
            Select("created_at").
            Where("id = ?", *cursor).
            Scan(&cursorCreatedAt).Error

        if err != nil {
            return nil, err
        }

        if direction == "next" {
            query = query.
                Where(
                    "(lt.created_at, lt.id) <= (?, ?)",
                    cursorCreatedAt,
                    *cursor,
                ).
                Order("lt.created_at DESC, lt.id DESC").
                Limit(11)
        }

        if direction == "previous" {
            query = query.
                Where(
                    "(lt.created_at, lt.id) > (?, ?)",
                    cursorCreatedAt,
                    *cursor,
                ).
                Order("lt.created_at ASC, lt.id ASC").
                Limit(11)
        }
    }

    err := query.Find(&transactions).Error
    if err != nil {
        return nil, err
    }

    return transactions, nil
}

func (lr *LedgerRepository) GetWalletTransactionById(
    accountID uuid.UUID,
    transactionID uuid.UUID,
) (*dto.LedgerTransactionRow, error) {

    var transaction dto.LedgerTransactionRow

    err := lr.db.
        Table("ledger_entries AS le").
        Select(`
            lt.id AS ledger_transaction_id,
            lt.transaction_ref,
            lt.type,
            lt.reference_id,
            lt.status,
            lt.settlement_status,
            le.amount,
            le.currency,
            le.entry_type,
            lt.created_at
        `).
        Joins(`
            JOIN ledger_transactions AS lt
                ON lt.id = le.ledger_transaction_id
        `).
        Where(
            "le.account_id = ? AND lt.id = ?",
            accountID,
            transactionID,
        ).
        First(&transaction).Error

    if err != nil {
        return nil, err
    }

    return &transaction, nil
}

func (lr *LedgerRepository) CreateLedgerEntry(requestEntry *models.LedgerEntry) (*models.LedgerEntry, error){
    ledgerEntry := &models.LedgerEntry{
        ID: utils.GenerateUUID(),
        LedgerTransactionID: requestEntry.LedgerTransactionID,
        AccountType: requestEntry.AccountType,
        EntryType: requestEntry.EntryType,
        AccountID: requestEntry.AccountID,
        Amount: requestEntry.Amount,
        Currency: requestEntry.Currency,
        CreatedAt: time.Now(),
    }
    if err := lr.db.Create(ledgerEntry).Error; err != nil {
		return nil, err
	}

	return ledgerEntry, nil
}

func (lr *LedgerRepository) CreateLedgerTransaction(requestTransaction *models.LedgerTransaction) (*models.LedgerTransaction, error){
    ledgerTransaction := &models.LedgerTransaction{
        ID: utils.GenerateUUID(),
        TransactionRef: utils.GenerateLedgerReference(),
        Type: requestTransaction.Type,
        ReferenceID: requestTransaction.ReferenceID,
        Status: requestTransaction.Status,
        SettlementStatus: requestTransaction.SettlementStatus,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

   if err := lr.db.Create(ledgerTransaction).Error; err != nil{
    return nil, err
   } 

   return ledgerTransaction, nil
}

func (lr *LedgerRepository) ExistingLedgerTransactionByReferenceID(
	referenceID string,
) (bool, error) {

	var ledgerTransaction models.LedgerTransaction

	err := lr.db.
		Where("reference_id = ?", referenceID).
		First(&ledgerTransaction).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (lr *LedgerRepository) UpdateLedgerStatus(ID uuid.UUID, status string, settlementStatus string) (*models.LedgerTransaction, error){
    updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

    if status != ""{
        updates["status"] = status
    }

    if settlementStatus != ""{
        updates["settlement_status"] = settlementStatus
    }

    ledgerTransaction, errLT := lr.GetLedgerTransactionByID(ID)
    if errLT != nil{
        return nil, errLT
    }

    if ledgerTransaction.Status == status && ledgerTransaction.SettlementStatus == settlementStatus{
        return  nil, errors.New("no status change")
    }

    err := lr.db.
		Model(&models.LedgerTransaction{}).
		Where("id = ?", ID).
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

    var updatedLedgerTransaction *models.LedgerTransaction
    errU := lr.db.Where("id = ?", ID).First(&updatedLedgerTransaction).Error
    if errU != nil{
        return nil, errU
    }

    return updatedLedgerTransaction, nil
}

func (lr *LedgerRepository) GetLedgerTransactionByID(ID uuid.UUID) (*models.LedgerTransaction, error){
    var ledgerTransaction *models.LedgerTransaction
    err := lr.db.Where("id = ?", ID).First(&ledgerTransaction).Error
    if err != nil{
        return nil, err
    }

    return ledgerTransaction, nil
}

func (lr *LedgerRepository) GetTransactionsByCreatedAtRange(
	startTimestamp time.Time,
	endTimestamp time.Time,
) ([]*models.LedgerTransaction, error) {

	var transactions []*models.LedgerTransaction

	err := lr.db.
		Where(
			"created_at >= ? AND created_at < ?",
			startTimestamp,
			endTimestamp,
		).
		Order("created_at ASC").
		Find(&transactions).Error

	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (lr *LedgerRepository) GetEntriesByTransactionIDs(
	transactionIDs []uuid.UUID,
) ([]*models.LedgerEntry, error) {

	var entries []*models.LedgerEntry

	err := lr.db.
		Where("ledger_transaction_id IN ?", transactionIDs).
		Order("created_at ASC").
		Find(&entries).Error

	if err != nil {
		return nil, err
	}

	return entries, nil
}