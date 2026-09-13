package activities

import (
	"context"
	"fmt"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PaymentWorkflowRequest struct {
	PaymentReference string
	MerchantID       uuid.UUID
	Fee              decimal.Decimal
	Request          dto.CreatePaymentRequest
}

func (a *Registry) ValidatePaymentRequest(
	ctx context.Context,
	request *dto.CreatePaymentRequest,
	merchantID uuid.UUID,
) error {
	return a.PaymentService.ValidatePaymentRequest(request, merchantID)
}

func (a *Registry) CalculateFees(
	amount decimal.Decimal,
) (decimal.Decimal, error) {
	fee, errF := a.PaymentService.CalculateFees(amount)
	if errF != nil {
		return decimal.Zero, errF
	}
	return fee, nil
}

func (a *Registry) CreatePayment(
	ctx context.Context,
	request *dto.CreatePaymentRequest,
	merchantID uuid.UUID,
	status *string,
) (*models.Payment, error) {
	return a.PaymentService.PaymentRepo.CreatePayment(merchantID, request, status)
}

func (a *Registry) UpdatePaymentStatus(
	ctx context.Context,
	status *string,
	paymentRef string,
) (*models.Payment, error) {
	return a.PaymentService.PaymentRepo.UpdatePaymentStatus(paymentRef, status)
}

func (a *Registry) ExecutePayment(
	ctx context.Context,
	request *PaymentWorkflowRequest,
) (*models.Payment, error) {

	var finalPaymentAfterUpdate *models.Payment
	err := a.DB.Transaction(
		func(tx *gorm.DB) error {

			merchantID := request.MerchantID
			paymentRequest := request.Request
			fees := request.Fee
			totalAmount := paymentRequest.Amount.Add(fees)

			// Transaction-aware repositories
			walletRepo := a.WalletService.WalletRepo.WithTx(tx)
			paymentRepo := a.PaymentService.PaymentRepo.WithTx(tx)
			vaultRepo := a.VaultService.VaultRepo.WithTx(tx)

			// ---------------------------------------------------------
			// 1. Mark payment as PROCESSING
			// ---------------------------------------------------------

			paymentStatus := constants.TransactionStatusProcessing

			payment, err := paymentRepo.UpdatePaymentStatus(
				request.PaymentReference,
				&paymentStatus,
			)
			if err != nil {
				return fmt.Errorf(
					"update payment status to processing: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 2. Check balance
			// ---------------------------------------------------------

			isBalanceSufficient, err := a.PaymentService.CheckBalance(
				merchantID,
				totalAmount,
			)
			if err != nil {
				return fmt.Errorf(
					"balance check failed: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 3. Get sender wallet
			// ---------------------------------------------------------

			senderWallet, err := walletRepo.GetWalletByMerchantId(
				merchantID,
			)
			if err != nil {
				return fmt.Errorf(
					"get sender wallet: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 4. Auto top-up if required
			// ---------------------------------------------------------

			updatedWallet := senderWallet

			if !isBalanceSufficient {
				updatedWallet, err = a.WalletService.TopUpWalletFromPrimaryBank(
					tx,
					senderWallet,
					totalAmount,
				)
				if err != nil {
					return fmt.Errorf(
						"auto top-up failed: %w",
						err,
					)
				}
			}

			// ---------------------------------------------------------
			// 5. Reserve sender balance
			// ---------------------------------------------------------

			reservedWalletAmount := updatedWallet.ReservedBalance.Add(
				totalAmount,
			)

			availableWalletAmount := updatedWallet.AvailableBalance.Sub(
				totalAmount,
			)

			updatedSenderWallet, err := walletRepo.UpdateWalletBalance(
				merchantID,
				&dto.UpdateWallletBalanceRequest{
					ReservedBalance:  reservedWalletAmount,
					AvailableBalance: availableWalletAmount,
				},
			)
			if err != nil {
				return fmt.Errorf(
					"update sender wallet balance: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 6. Get Payment Vault
			// ---------------------------------------------------------

			paymentVault, err := vaultRepo.GetVaultByType(
				constants.PaymentVault,
			)
			if err != nil {
				return fmt.Errorf(
					"get payment vault: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 7. Wallet A -> Payment Vault
			// ---------------------------------------------------------

			_, err = vaultRepo.UpdateVaultBalance(
				paymentVault.Balance.Add(totalAmount),
				constants.PaymentVault,
			)
			if err != nil {
				return fmt.Errorf(
					"update payment vault balance: %w",
					err,
				)
			}

			// Release sender reservation after movement
			_, err = walletRepo.UpdateWalletBalance(
				merchantID,
				&dto.UpdateWallletBalanceRequest{
					ReservedBalance: updatedSenderWallet.ReservedBalance.Sub(
						totalAmount,
					),
				},
			)
			if err != nil {
				return fmt.Errorf(
					"release sender wallet reservation: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 8. Ledger: Wallet A -> Payment Vault
			// ---------------------------------------------------------

			ledgerTransaction, err := a.LedgerService.PostTransaction(
				tx,
				&dto.PostLedgerTransactionRequest{
					ReferenceID:      *payment.PaymentReference,
					Type:             constants.LedgerTransactionTypePayment,
					Status:           constants.TransactionStatusCompleted,
					SettlementStatus: constants.LedgerSettlementPending,
					Currency:         "INR",
				},
				[]*models.LedgerEntry{
					{
						AccountType: constants.LedgerAccountTypeWallet,
						AccountID:   senderWallet.ID,
						EntryType:   constants.LedgerEntryTypeDebit,
						Amount:      totalAmount,
						Currency:    "INR",
					},
					{
						AccountType: constants.LedgerAccountTypeVault,
						AccountID:   paymentVault.ID,
						EntryType:   constants.LedgerEntryTypeCredit,
						Amount:      totalAmount,
						Currency:    "INR",
					},
				},
			)
			if err != nil {
				return fmt.Errorf(
					"create wallet to payment vault ledger: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 9. Get Receiver Wallet
			// ---------------------------------------------------------

			receiverWallet, err := walletRepo.GetWalletByMerchantId(
				paymentRequest.ReceiverMerchantID,
			)
			if err != nil {
				return fmt.Errorf(
					"get receiver wallet: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 10. Payment Vault -> Receiver Wallet
			// ---------------------------------------------------------

			_, err = walletRepo.UpdateWalletBalance(
				paymentRequest.ReceiverMerchantID,
				&dto.UpdateWallletBalanceRequest{
					ReservedBalance: receiverWallet.ReservedBalance.Add(
						paymentRequest.Amount,
					),
				},
			)
			if err != nil {
				return fmt.Errorf(
					"update receiver wallet balance: %w",
					err,
				)
			}

			updatedPaymentVaultBalance := paymentVault.Balance.
				Add(totalAmount).
				Sub(paymentRequest.Amount)

			updatedPaymentVault, err := vaultRepo.UpdateVaultBalance(
				updatedPaymentVaultBalance,
				constants.PaymentVault,
			)
			if err != nil {
				return fmt.Errorf(
					"update payment vault after receiver transfer: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 11. Ledger: Payment Vault -> Receiver Wallet
			// ---------------------------------------------------------

			_, err = a.LedgerService.CreateLedgerEntries(
				tx,
				[]*models.LedgerEntry{
					{
						LedgerTransactionID: ledgerTransaction.ID,
						AccountType:         constants.LedgerAccountTypeVault,
						AccountID:           paymentVault.ID,
						EntryType:           constants.LedgerEntryTypeDebit,
						Amount:              paymentRequest.Amount,
						Currency:            "INR",
					},
					{
						LedgerTransactionID: ledgerTransaction.ID,
						AccountType:         constants.LedgerAccountTypeWallet,
						AccountID:           receiverWallet.ID,
						EntryType:           constants.LedgerEntryTypeCredit,
						Amount:              paymentRequest.Amount,
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

			// ---------------------------------------------------------
			// 12. Payment Vault -> Company Vault
			// ---------------------------------------------------------

			paymentVaultBalanceAfterFees := updatedPaymentVault.Balance.Sub(
				fees,
			)

			_, err = vaultRepo.UpdateVaultBalance(
				paymentVaultBalanceAfterFees,
				constants.PaymentVault,
			)
			if err != nil {
				return fmt.Errorf(
					"deduct fees from payment vault: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 13. Get Company Vault
			// ---------------------------------------------------------

			companyVault, err := vaultRepo.GetVaultByType(
				constants.CompanyVault,
			)
			if err != nil {
				return fmt.Errorf(
					"get company vault: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 14. Add fees to Company Vault
			// ---------------------------------------------------------

			updatedCompanyVault, err := vaultRepo.UpdateVaultBalance(
				companyVault.Balance.Add(fees),
				constants.CompanyVault,
			)
			if err != nil {
				return fmt.Errorf(
					"update company vault balance: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 15. Ledger: Payment Vault -> Company Vault
			// ---------------------------------------------------------

			_, err = a.LedgerService.CreateLedgerEntries(
				tx,
				[]*models.LedgerEntry{
					{
						LedgerTransactionID: ledgerTransaction.ID,
						AccountType:         constants.LedgerAccountTypeVault,
						AccountID:           paymentVault.ID,
						EntryType:           constants.LedgerEntryTypeDebit,
						Amount:              fees,
						Currency:            "INR",
					},
					{
						LedgerTransactionID: ledgerTransaction.ID,
						AccountType:         constants.LedgerAccountTypeVault,
						AccountID:           updatedCompanyVault.ID,
						EntryType:           constants.LedgerEntryTypeCredit,
						Amount:              fees,
						Currency:            "INR",
					},
				},
			)
			if err != nil {
				return fmt.Errorf(
					"create fee ledger entries: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 16. Mark payment COMPLETED
			// ---------------------------------------------------------

			completedStatus := constants.TransactionStatusCompleted

			updatedPayment, err := paymentRepo.UpdatePaymentStatus(
				*payment.PaymentReference,
				&completedStatus,
			)
			if err != nil {
				return fmt.Errorf(
					"update payment status to completed: %w",
					err,
				)
			}

			// Return the updated payment.
			// Returning nil error commits the transaction.
			finalPaymentAfterUpdate = updatedPayment
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return finalPaymentAfterUpdate, nil
}
