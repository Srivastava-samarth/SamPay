package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (a *Registry) GetUnsettledPayment(
	ctx context.Context,
) ([]*models.Payment, error) {
	return a.Repo.GetPaymentsBySettlementStatus(constants.LedgerSettlementPending)
}

func (a *Registry) ExecuteSettlementByMerchant(
	ctx context.Context,
	merchantID uuid.UUID,
	payment *models.Payment,
) error {
	settledStatus := constants.LedgerSettlementSettled
	if payment.SettlementStatus == &settledStatus {
		return nil
	}
	return a.DB.Transaction(func(tx *gorm.DB) error {
		repo := a.Repo.WithTx(tx)

		paymentVault, errPV := repo.GetVaultByType(constants.PaymentVault)
		if errPV != nil {
			return errPV
		}

		wallet, errW := repo.GetWalletByMerchantId(merchantID)
		if errW != nil {
			return errW
		}

		ledgerTransaction, errLT := repo.GetLedgerTransactionByReferenceID(*payment.PaymentReference)
		if errLT != nil {
			return errLT
		}

		if paymentVault.Balance.LessThan(payment.Amount) {
			return errors.New("insufficient balance of payment vault")
		}

		_, errUPV := repo.UpdateVaultBalance(paymentVault.Balance.Sub(payment.Amount), constants.PaymentVault)
		if errUPV != nil {
			return errUPV
		}

		updatedWallet, errUW := repo.UpdateWalletBalance(merchantID, &dto.UpdateWalletBalanceRequest{
			AvailableBalance: wallet.AvailableBalance.Add(payment.Amount),
		})

		if errUW != nil {
			return errUW
		}

		_, err := a.Services.CreateLedgerEntries(
			tx,
			[]*models.LedgerEntry{
				{
					LedgerTransactionID: ledgerTransaction.ID,
					AccountType:         constants.LedgerAccountTypeVault,
					AccountID:           paymentVault.ID,
					EntryType:           constants.LedgerEntryTypeDebit,
					Amount:              payment.Amount,
					Currency:            "INR",
				},
				{
					LedgerTransactionID: ledgerTransaction.ID,
					AccountType:         constants.LedgerAccountTypeWallet,
					AccountID:           updatedWallet.ID,
					EntryType:           constants.LedgerEntryTypeCredit,
					Amount:              payment.Amount,
					Currency:            "INR",
				},
			},
		)
		if err != nil {
			return fmt.Errorf(
				"create settlement ledger entries: %w",
				err,
			)
		}

		settlementStatus := constants.LedgerSettlementSettled
		_, errUSS := repo.UpdateSettlementStatusByID(
			payment.ID,
			&settlementStatus,
		)
		if errUSS != nil {
			return fmt.Errorf(
				"failed to update settlement status for payment %s: %w",
				payment.ID,
				errUSS,
			)
		}

		return nil
	})
}
