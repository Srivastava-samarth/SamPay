package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (a *Registry) ValidateRefundRequest(
	ctx context.Context,
	request *dto.CreateRefundRequest,
) error {
	return a.Services.ValidateCreateRefundRequest(request)
}

func (a *Registry) UpdateRefundStatus(
	ctx context.Context,
	refundRef string,
	status string,
) (*models.Refund, error) {
	return a.Repo.UpdateRefundStatus(refundRef, status)
}

func (a *Registry) GetRefundByPaymentID(
	ctx context.Context,
	paymentID uuid.UUID,
) ([]*models.Refund, error) {
	return a.Repo.GetRefundsByPaymentID(paymentID)
}

func (a *Registry) GetPaymentByID(
	ctx context.Context,
	paymentID uuid.UUID,
) (*models.Payment, error) {
	return a.Repo.GetPaymentByID(paymentID)
}

func (a *Registry) CreateRefund(
	ctx context.Context,
	request *dto.CreateRefundRequest,
	refunds []*models.Refund,
) (*models.Refund, error) {
	Amount := request.Amount
	refundedAmount := decimal.Zero

	for _, refund := range refunds {
		if refund == nil || refund.Status == "" {
			continue
		}

		if refund.Status == constants.TransactionStatusRefunded {
			refundedAmount = refundedAmount.Add(refund.Amount)
		}
	}

	remainingRefundable := Amount.Sub(refundedAmount)

	if request.Amount.GreaterThan(remainingRefundable) {
		return nil, errors.New("refund amount exceeds refundable amount")
	}

	return a.Repo.CreateRefund(request)
}

func (a *Registry) RefundFromMerchantWallet(
	ctx context.Context,
	refund *models.Refund,
	payment *models.Payment,
) (*models.Refund, error) {
	var finalRefund *models.Refund
	if refund == nil {
		return nil, errors.New("refund not found")
	}

	if payment == nil {
		return nil, errors.New("payment not found")
	}
	err := a.DB.Transaction(
		func(tx *gorm.DB) error {
			refundSenderMerchantID := payment.ReceiverMerchantID
			Amount := refund.Amount

			// Transaction-aware repositories
			repo := a.Repo.WithTx(tx)
			refundReceiverMerchantID := refund.MerchantID
			refundSenderWallet, errRSW := repo.GetWalletByMerchantId(refundSenderMerchantID)
			if errRSW != nil {
				return errRSW
			}

			if refundSenderWallet == nil {
				return errors.New("refund sender wallet not found")
			}

			refundReceiverWallet, errRRW := repo.GetWalletByMerchantId(refundReceiverMerchantID)
			if errRRW != nil {
				return errRRW
			}

			if refundReceiverWallet == nil {
				return errors.New("refund receiver wallet not found")
			}

			if refundSenderWallet.AvailableBalance.LessThan(Amount) {
				topupSenderWallet, errTSW := a.Services.TopUpWalletFromPrimaryBank(tx, refundSenderWallet, Amount)
				if errTSW != nil {
					return errTSW
				}

				refundSenderWallet = topupSenderWallet
			}

			updateSenderAvailableAmount := refundSenderWallet.AvailableBalance.Sub(Amount)
			updatedSenderReservedAmount := refundSenderWallet.ReservedBalance.Add(Amount)
			updatedRefundSenderWallet, errURSW := repo.UpdateWalletBalance(refundSenderMerchantID, &dto.UpdateWalletBalanceRequest{
				AvailableBalance: &updateSenderAvailableAmount,
				ReservedBalance:  &updatedSenderReservedAmount,
			})

			if errURSW != nil {
				return errURSW
			}

			paymentVault, errPV := repo.GetVaultByType(constants.PaymentVault)
			if errPV != nil {
				return errPV
			}

			if paymentVault == nil {
				return errors.New("payment vault not found")
			}

			if paymentVault.Balance.LessThan(Amount) {
				return errors.New("insufficient payment vault balance for refund")
			}

			updatedVaultBalance := paymentVault.Balance.Add(Amount)
			updatedPaymentVault, errUPV := repo.UpdateVaultBalance(updatedVaultBalance, constants.PaymentVault)
			if errUPV != nil {
				return errUPV
			}

			updatedRefundSenderWalletAfterBalanceResvBal := updatedRefundSenderWallet.ReservedBalance.Sub(Amount)
			updatedRefundSenderWalletAfterBalance, errRSWAB := repo.UpdateWalletBalance(refundSenderMerchantID, &dto.UpdateWalletBalanceRequest{
				ReservedBalance: &updatedRefundSenderWalletAfterBalanceResvBal,
			})

			if errRSWAB != nil {
				return errRSWAB
			}

			ledgerTransaction, err := a.Services.PostTransaction(
				tx,
				&dto.PostLedgerTransactionRequest{
					ReferenceID:      refund.RefundReference,
					Type:             constants.LedgerTransactionTypeRefund,
					Status:           constants.TransactionStatusProcessing,
					SettlementStatus: constants.LedgerSettlementSettled,
					Currency:         "INR",
				},
				[]*models.LedgerEntry{
					{
						AccountType: constants.LedgerAccountTypeWallet,
						AccountID:   updatedRefundSenderWalletAfterBalance.ID,
						EntryType:   constants.LedgerEntryTypeDebit,
						Amount:      Amount,
						Currency:    "INR",
					},
					{
						AccountType: constants.LedgerAccountTypeVault,
						AccountID:   updatedPaymentVault.ID,
						EntryType:   constants.LedgerEntryTypeCredit,
						Amount:      Amount,
						Currency:    "INR",
					},
				},
			)
			if err != nil {
				return fmt.Errorf(
					"create wallet to payout vault ledger: %w",
					err,
				)
			}

			updateBalance := updatedPaymentVault.Balance.Sub(Amount)
			updatePaymentVaultAfterBalanceMove, errUPVAB := repo.UpdateVaultBalance(updateBalance, constants.PaymentVault)
			if errUPVAB != nil {
				return errUPVAB
			}
			updatedRefundReceiverWalletAvalBal := refundReceiverWallet.AvailableBalance.Add(Amount)
			updatedRefundReceiverWallet, errURRW := repo.UpdateWalletBalance(refundReceiverMerchantID, &dto.UpdateWalletBalanceRequest{
				AvailableBalance: &updatedRefundReceiverWalletAvalBal,
			})

			if errURRW != nil {
				return errURRW
			}

			_, err = a.Services.CreateLedgerEntries(
				tx,
				[]*models.LedgerEntry{
					{
						LedgerTransactionID: ledgerTransaction.ID,
						AccountType:         constants.LedgerAccountTypeVault,
						AccountID:           updatePaymentVaultAfterBalanceMove.ID,
						EntryType:           constants.LedgerEntryTypeDebit,
						Amount:              Amount,
						Currency:            "INR",
					},
					{
						LedgerTransactionID: ledgerTransaction.ID,
						AccountType:         constants.LedgerAccountTypeWallet,
						AccountID:           updatedRefundReceiverWallet.ID,
						EntryType:           constants.LedgerEntryTypeCredit,
						Amount:              Amount,
						Currency:            "INR",
					},
				},
			)
			if err != nil {
				return fmt.Errorf(
					"create receiver ledger entries: %w",
					err,
				)
			}

			_, errULT := repo.UpdateLedgerStatus(ledgerTransaction.ID, constants.TransactionStatusCompleted, constants.LedgerSettlementSettled)
			if errULT != nil {
				return errULT
			}

			updateRefund, errUR := repo.UpdateRefundStatus(refund.RefundReference, constants.TransactionStatusRefunded)
			if errUR != nil {
				return errUR
			}

			finalRefund = updateRefund
			return nil
		},
	)
	if err != nil {
		updateRefund, errUR := a.UpdateRefundStatus(ctx, refund.RefundReference, constants.TransactionStatusFailed)
		if errUR != nil {
			return nil, errUR
		}

		return updateRefund, err
	}
	return finalRefund, nil
}

func (a *Registry) RefundFromPaymentVault(
	ctx context.Context,
	refund *models.Refund,
	payment *models.Payment,
) (*models.Refund, error) {
	var finalRefund *models.Refund
	if refund == nil {
		return nil, errors.New("refund not found")
	}

	if payment == nil {
		return nil, errors.New("payment not found")
	}
	err := a.DB.Transaction(
		func(tx *gorm.DB) error {
			Amount := refund.Amount

			// Transaction-aware repositories
			repo := a.Repo.WithTx(tx)

			refundReceiverMerchantID := payment.SenderMerchantID
			refundReceiverWallet, errRRW := repo.GetWalletByMerchantId(refundReceiverMerchantID)
			if errRRW != nil {
				return errRRW
			}

			if refundReceiverWallet == nil {
				return errors.New("refund receiver wallet not found")
			}

			paymentVault, errPV := repo.GetVaultByType(constants.PaymentVault)
			if errPV != nil {
				return errPV
			}

			if paymentVault == nil {
				return errors.New("payment vault not found")
			}

			if paymentVault.Balance.LessThan(Amount) {
				return errors.New("insufficient payment vault balance for refund")
			}

			updatedVaultBalance := paymentVault.Balance.Sub(Amount)
			updatedPaymentVault, errUPV := repo.UpdateVaultBalance(updatedVaultBalance, constants.PaymentVault)
			if errUPV != nil {
				return errUPV
			}
			updatedReceiverWalletNewAvalBal := refundReceiverWallet.AvailableBalance.Add(Amount)
			updatedReceiverWallet, errURW := repo.UpdateWalletBalance(refundReceiverMerchantID, &dto.UpdateWalletBalanceRequest{
				AvailableBalance: &updatedReceiverWalletNewAvalBal,
			})

			if errURW != nil {
				return errURW
			}

			ledgerTransaction, err := a.Services.PostTransaction(
				tx,
				&dto.PostLedgerTransactionRequest{
					ReferenceID:      refund.RefundReference,
					Type:             constants.LedgerTransactionTypeRefund,
					Status:           constants.TransactionStatusProcessing,
					SettlementStatus: constants.LedgerSettlementSettled,
					Currency:         "INR",
				},
				[]*models.LedgerEntry{
					{
						AccountType: constants.LedgerAccountTypeVault,
						AccountID:   updatedPaymentVault.ID,
						EntryType:   constants.LedgerEntryTypeDebit,
						Amount:      Amount,
						Currency:    "INR",
					},
					{
						AccountType: constants.LedgerAccountTypeWallet,
						AccountID:   updatedReceiverWallet.ID,
						EntryType:   constants.LedgerEntryTypeCredit,
						Amount:      Amount,
						Currency:    "INR",
					},
				},
			)
			if err != nil {
				return fmt.Errorf(
					"create refund ledger: %w",
					err,
				)
			}

			_, errULT := repo.UpdateLedgerStatus(ledgerTransaction.ID, constants.TransactionStatusCompleted, constants.LedgerSettlementSettled)
			if errULT != nil {
				return errULT
			}

			updateRefund, errUR := repo.UpdateRefundStatus(refund.RefundReference, constants.TransactionStatusRefunded)
			if errUR != nil {
				return errUR
			}

			finalRefund = updateRefund
			return nil
		},
	)
	if err != nil {
		updateRefund, errUR := a.UpdateRefundStatus(ctx, refund.RefundReference, constants.TransactionStatusFailed)
		if errUR != nil {
			return nil, errUR
		}

		return updateRefund, err
	}
	return finalRefund, nil
}
