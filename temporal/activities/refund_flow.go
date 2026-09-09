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
	return a.RefundService.ValidateCreateRefundRequest(request)
}

func (a *Registry) UpdateRefundStatus(
	ctx context.Context,
	refundRef string,
	status string,
) (*models.Refund, error) {
	return a.RefundService.RefundRepo.UpdateRefundStatus(refundRef, status)
}

func (a *Registry) GetRefundByPaymentID(
	ctx context.Context,
	paymentID uuid.UUID,
) ([]*models.Refund, error) {
	return a.RefundService.RefundRepo.GetRefundsByPaymentID(paymentID)
}

func (a *Registry) GetPaymentByID(
	ctx context.Context,
	paymentID uuid.UUID,
) (*models.Payment, error) {
	return a.PaymentService.PaymentRepo.GetPaymentByID(paymentID)
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

	return a.RefundService.RefundRepo.CreateRefund(request)
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
			refundSenderMerchantID := refund.MerchantID
			Amount := refund.Amount

			// Transaction-aware repositories
			walletRepo := a.WalletService.WalletRepo.WithTx(tx)
			vaultRepo := a.VaultService.VaultRepo.WithTx(tx)
			refundRepo := a.RefundService.RefundRepo.WithTx(tx)
			ledgerRepo := a.LedgerService.LedgerRepo.WithTx(tx)

			refundReceiverMerchantID := payment.SenderMerchantID
			refundSenderWallet, errRSW := walletRepo.GetWalletByMerchantId(refundSenderMerchantID)
			if errRSW != nil {
				return errRSW
			}

			if refundSenderWallet == nil {
				return errors.New("refund sender wallet not found")
			}

			refundReceiverWallet, errRRW := walletRepo.GetWalletByMerchantId(refundReceiverMerchantID)
			if errRRW != nil {
				return errRRW
			}

			if refundReceiverWallet == nil {
				return errors.New("refund receiver wallet not found")
			}

			if refundSenderWallet.AvailableBalance.LessThan(Amount) {
				topupSenderWallet, errTSW := a.WalletService.TopUpWalletFromPrimaryBank(tx, refundSenderWallet, Amount)
				if errTSW != nil {
					return errTSW
				}

				refundSenderWallet = topupSenderWallet
			}

			updateSenderAvailableAmount := refundSenderWallet.AvailableBalance.Sub(Amount)
			updatedSenderReservedAmount := refundSenderWallet.ReservedBalance.Add(Amount)
			updatedRefundSenderWallet, errURSW := walletRepo.UpdateWalletBalance(refundSenderMerchantID, &dto.UpdateWallletBalanceRequest{
				AvailableBalance: updateSenderAvailableAmount,
				ReservedBalance:  updatedSenderReservedAmount,
			})

			if errURSW != nil {
				return errURSW
			}

			paymentVault, errPV := vaultRepo.GetVaultByType(constants.PaymentVault)
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
			updatedPaymentVault, errUPV := vaultRepo.UpdateVaultBalance(updatedVaultBalance, constants.PaymentVault)
			if errUPV != nil {
				return errUPV
			}

			updatedRefundSenderWalletAfterBalance, errRSWAB := walletRepo.UpdateWalletBalance(refundSenderMerchantID, &dto.UpdateWallletBalanceRequest{
				ReservedBalance: updatedRefundSenderWallet.ReservedBalance.Sub(Amount),
			})

			if errRSWAB != nil {
				return errRSWAB
			}

			ledgerTransaction, err := a.LedgerService.PostTransaction(
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
			updatePaymentVaultAfterBalanceMove, errUPVAB := vaultRepo.UpdateVaultBalance(updateBalance, constants.PaymentVault)
			if errUPVAB != nil {
				return errUPVAB
			}

			updatedRefundReceiverWallet, errURRW := walletRepo.UpdateWalletBalance(refundReceiverMerchantID, &dto.UpdateWallletBalanceRequest{
				AvailableBalance: refundReceiverWallet.AvailableBalance.Add(Amount),
			})

			if errURRW != nil {
				return errURRW
			}

			_, err = a.LedgerService.CreateLedgerEntries(
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

			_, errULT := ledgerRepo.UpdateLedgerStatus(ledgerTransaction.ID, constants.TransactionStatusCompleted, constants.LedgerSettlementSettled)
			if errULT != nil {
				return errULT
			}

			updateRefund, errUR := refundRepo.UpdateRefundStatus(refund.RefundReference, constants.TransactionStatusRefunded)
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
			walletRepo := a.WalletService.WalletRepo.WithTx(tx)
			vaultRepo := a.VaultService.VaultRepo.WithTx(tx)
			refundRepo := a.RefundService.RefundRepo.WithTx(tx)
			ledgerRepo := a.LedgerService.LedgerRepo.WithTx(tx)

			refundReceiverMerchantID := payment.SenderMerchantID
			refundReceiverWallet, errRRW := walletRepo.GetWalletByMerchantId(refundReceiverMerchantID)
			if errRRW != nil {
				return errRRW
			}

			if refundReceiverWallet == nil {
				return errors.New("refund receiver wallet not found")
			}

			paymentVault, errPV := vaultRepo.GetVaultByType(constants.PaymentVault)
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
			updatedPaymentVault, errUPV := vaultRepo.UpdateVaultBalance(updatedVaultBalance, constants.PaymentVault)
			if errUPV != nil {
				return errUPV
			}

			updatedReceiverWallet, errURW := walletRepo.UpdateWalletBalance(refundReceiverMerchantID, &dto.UpdateWallletBalanceRequest{
				AvailableBalance: refundReceiverWallet.AvailableBalance.Add(Amount),
			})

			if errURW != nil {
				return errURW
			}

			ledgerTransaction, err := a.LedgerService.PostTransaction(
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

			_, errULT := ledgerRepo.UpdateLedgerStatus(ledgerTransaction.ID, constants.TransactionStatusCompleted, constants.LedgerSettlementSettled)
			if errULT != nil {
				return errULT
			}

			updateRefund, errUR := refundRepo.UpdateRefundStatus(refund.RefundReference, constants.TransactionStatusRefunded)
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
