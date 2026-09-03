package repositories

import (
	"time"

	"github.com/Srivastava-samarth/sampay/dto"
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