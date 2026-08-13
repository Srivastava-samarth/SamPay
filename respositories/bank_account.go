package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BankRepository struct{
	db *gorm.DB
}

func NewBankRepository(db *gorm.DB) *BankRepository{
	return &BankRepository{
		db: db,
	}
}

func (wr *BankRepository) WithTx(tx *gorm.DB) *BankRepository {
	return &BankRepository{
		db: tx,
	}
}

func(br *BankRepository) CreateBankAccount(
	bankAccount *models.BankAccount,
) (*models.BankAccount, error) {

	newBankAccount := &models.BankAccount{
		ID:            utils.GenerateUUID(),
		AccountNumber: utils.GenerateBankAccountNumber(),
		AccountName:   bankAccount.AccountName,
		BankName:      "SamPay Bank",
		IFSCCode:      "SAMP5917AY",
		AccountType:   bankAccount.AccountType,
		Balance:       decimal.Zero,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := br.db.Create(newBankAccount).Error; err != nil {
		return nil, err
	}

	return newBankAccount, nil
}

func(br *BankRepository) GetBankAccountByID(ID uuid.UUID) (*models.BankAccount, error) {
	var bankAccount models.BankAccount
	err := br.db.Where("id = ?", ID).First(&bankAccount).Error
	if err != nil {
		return nil, err
	}
	return &bankAccount, nil
}