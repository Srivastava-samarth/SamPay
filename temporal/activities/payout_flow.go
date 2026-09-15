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

type PayoutWorkflowWalletToBanRequest struct {
	PayoutReference string
	MerchantID      uuid.UUID
	Fee             decimal.Decimal
	Request         dto.CreateWalletToBankRequest
}

type PayoutWorkflowBankToBanRequest struct {
	PayoutReference string
	MerchantID      uuid.UUID
	Fee             decimal.Decimal
	Request         dto.CreateBankToBankRequest
}

func (a *Registry) ValidatePayoutWalletToBankRequest(
	ctx context.Context,
	request *dto.CreateWalletToBankRequest,
	merchantID uuid.UUID,
) error {
	return a.Services.ValidatePayoutRequestWalletToBank(request, merchantID)
}

func (a *Registry) ValidatePayoutBankToBankRequest(
	ctx context.Context,
	request *dto.CreateBankToBankRequest,
	merchantID uuid.UUID,
) error {
	return a.Services.ValidatePayoutRequestBankToBank(request, merchantID)
}

func (a *Registry) CreatePayoutWalletToBank(
	ctx context.Context,
	request *dto.CreateWalletToBankRequest,
	merchantID uuid.UUID,
	status string,
) (*models.Payout, error) {
	return a.Repo.CreatePayoutWalletToBank(merchantID, request, status)
}

func (a *Registry) CreatePayoutBankToBank(
	ctx context.Context,
	request *dto.CreateBankToBankRequest,
	merchantID uuid.UUID,
	status string,
) (*models.Payout, error) {
	return a.Repo.CreatePayoutBankToBank(merchantID, request, status)
}

func (a *Registry) UpdatePayoutStatus(
	ctx context.Context,
	status string,
	paymentRef string,
) (*models.Payout, error) {
	return a.Repo.UpdatePayoutStatus(paymentRef, status)
}

func (a *Registry) ExecutePayoutWalletToBank(
	ctx context.Context,
	request *PayoutWorkflowWalletToBanRequest,
) (*models.Payout, error) {

	var finalPayoutAfterUpdate *models.Payout
	err := a.DB.Transaction(
		func(tx *gorm.DB) error {

			merchantID := request.MerchantID
			payoutRequest := request.Request
			fees := request.Fee
			totalAmount := payoutRequest.Amount.Add(fees)

			// Transaction-aware repositories
			repo := a.Repo.WithTx(tx)

			// ---------------------------------------------------------
			// 1. Mark payment as PROCESSING
			// ---------------------------------------------------------

			payoutStatus := constants.TransactionStatusProcessing

			payout, err := repo.UpdatePayoutStatus(
				request.PayoutReference,
				payoutStatus,
			)
			if err != nil {
				return fmt.Errorf(
					"update payout status to processing: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 2. Check balance
			// ---------------------------------------------------------

			isBalanceSufficient, err := a.Services.CheckBalance(
				merchantID,
				totalAmount,
			)
			if err != nil {
				return fmt.Errorf(
					"balance check failed: %w",
					err,
				)
			}

			if !isBalanceSufficient {
				return fmt.Errorf(
					"balance is insufficient: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 3. Get sender wallet
			// ---------------------------------------------------------

			senderWallet, err := repo.GetWalletByMerchantId(
				merchantID,
			)
			if err != nil {
				return fmt.Errorf(
					"get sender wallet: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 5. Reserve sender balance
			// ---------------------------------------------------------

			reservedWalletAmount := senderWallet.ReservedBalance.Add(
				totalAmount,
			)

			availableWalletAmount := senderWallet.AvailableBalance.Sub(
				totalAmount,
			)

			updatedSenderWallet, err := repo.UpdateWalletBalance(
				merchantID,
				&dto.UpdateWalletBalanceRequest{
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
			// 6. Get Payout Vault
			// ---------------------------------------------------------

			payoutVault, err := repo.GetVaultByType(
				constants.PayoutVault,
			)
			if err != nil {
				return fmt.Errorf(
					"get payout vault: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 7. Wallet A -> Payout Vault
			// ---------------------------------------------------------

			_, err = repo.UpdateVaultBalance(
				payoutVault.Balance.Add(totalAmount),
				constants.PayoutVault,
			)
			if err != nil {
				return fmt.Errorf(
					"update payout vault balance: %w",
					err,
				)
			}

			// Release sender reservation after movement
			_, err = repo.UpdateWalletBalance(
				merchantID,
				&dto.UpdateWalletBalanceRequest{
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
			// 8. Ledger: Wallet A -> Payout Vault
			// ---------------------------------------------------------

			ledgerTransaction, err := a.Services.PostTransaction(
				tx,
				&dto.PostLedgerTransactionRequest{
					ReferenceID:      payout.PayoutReference,
					Type:             constants.LedgerTransactionTypePayout,
					Status:           constants.TransactionStatusCompleted,
					SettlementStatus: constants.LedgerSettlementSettled,
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
						AccountID:   payoutVault.ID,
						EntryType:   constants.LedgerEntryTypeCredit,
						Amount:      totalAmount,
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

			// ---------------------------------------------------------
			// 9. Get Receiver Wallet
			// ---------------------------------------------------------

			receiverBankAccount, err := repo.GetBankAccountByID(
				payoutRequest.DestinationBankAccountID,
			)
			if err != nil {
				return fmt.Errorf(
					"get receiver bank account: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 10. Payout Vault -> Receiver bank account
			// ---------------------------------------------------------

			_, err = repo.UpdateBankAccount(
				receiverBankAccount.ID,
				&dto.UpdateBankAccountRequest{
					Balance: receiverBankAccount.Balance.Add(
						payoutRequest.Amount,
					),
				},
			)
			if err != nil {
				return fmt.Errorf(
					"update receiver bank account balance: %w",
					err,
				)
			}

			updatedPayoutVaultBalance := payoutVault.Balance.
				Add(totalAmount).
				Sub(payoutRequest.Amount)

			updatedPayoutVault, err := repo.UpdateVaultBalance(
				updatedPayoutVaultBalance,
				constants.PayoutVault,
			)
			if err != nil {
				return fmt.Errorf(
					"update payout vault after receiver transfer: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 11. Ledger: Payment Vault -> Receiver Wallet
			// ---------------------------------------------------------
			_, err = a.Services.CreateLedgerEntries(
				tx,
				[]*models.LedgerEntry{
					{
						LedgerTransactionID: ledgerTransaction.ID,
						AccountType:         constants.LedgerAccountTypeVault,
						AccountID:           payoutVault.ID,
						EntryType:           constants.LedgerEntryTypeDebit,
						Amount:              payoutRequest.Amount,
						Currency:            "INR",
					},
					{
						LedgerTransactionID: ledgerTransaction.ID,
						AccountType:         constants.LedgerAccountTypeBankAccount,
						AccountID:           receiverBankAccount.ID,
						EntryType:           constants.LedgerEntryTypeCredit,
						Amount:              payoutRequest.Amount,
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
			// 12. Payoutt Vault -> Company Vault
			// ---------------------------------------------------------

			payoutVaultBalanceAfterFees := updatedPayoutVault.Balance.Sub(
				fees,
			)

			_, err = repo.UpdateVaultBalance(
				payoutVaultBalanceAfterFees,
				constants.PayoutVault,
			)
			if err != nil {
				return fmt.Errorf(
					"deduct fees from payout vault: %w",
					err,
				)
			}

			// ---------------------------------------------------------
			// 13. Get Company Vault
			// ---------------------------------------------------------

			companyVault, err := repo.GetVaultByType(
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

			updatedCompanyVault, err := repo.UpdateVaultBalance(
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

			_, err = a.Services.CreateLedgerEntries(
				tx,
				[]*models.LedgerEntry{
					{
						LedgerTransactionID: ledgerTransaction.ID,
						AccountType:         constants.LedgerAccountTypeVault,
						AccountID:           payoutVault.ID,
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

			updatedPayout, err := repo.UpdatePayoutStatus(
				payout.PayoutReference,
				constants.TransactionStatusCompleted,
			)
			if err != nil {
				return fmt.Errorf(
					"update payment status to completed: %w",
					err,
				)
			}

			// Return the updated payment.
			// Returning nil error commits the transaction.
			finalPayoutAfterUpdate = updatedPayout
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return finalPayoutAfterUpdate, nil
}

func (a *Registry) ExecutePayoutBankToBank(
	ctx context.Context,
	request *PayoutWorkflowBankToBanRequest,
) (*models.Payout, error) {

	var finalPayoutAfterUpdate *models.Payout
	err := a.DB.Transaction(
		func(tx *gorm.DB) error {

			payoutRequest := request.Request
			fees := request.Fee
			totalAmount := payoutRequest.Amount.Add(fees)

			// Transaction-aware repositories
			repo := a.Repo.WithTx(tx)

			payoutStatus := constants.TransactionStatusProcessing

			payout, err := repo.UpdatePayoutStatus(
				request.PayoutReference,
				payoutStatus,
			)
			if err != nil {
				return fmt.Errorf(
					"update payout status to processing: %w",
					err,
				)
			}

			isBalanceSufficient, err := a.Services.CheckBankBalance(
				request.Request.SourceBankAccountID,
				totalAmount,
			)
			if err != nil {
				return fmt.Errorf(
					"balance check failed: %w",
					err,
				)
			}

			if !isBalanceSufficient {
				return fmt.Errorf(
					"balance is insufficient: %w",
					err,
				)
			}

			senderBankAccount, err := repo.GetBankAccountByID(request.Request.SourceBankAccountID)
			if err != nil {
				return fmt.Errorf(
					"get sender wallet: %w",
					err,
				)
			}

			updatedSenderBankAccount, err := repo.UpdateBankAccount(
				senderBankAccount.ID,
				&dto.UpdateBankAccountRequest{
					Balance: senderBankAccount.Balance.Sub(totalAmount),
				},
			)
			if err != nil {
				return fmt.Errorf(
					"update sender bank account balance: %w",
					err,
				)
			}

			receiverBankAccount, errRBA := repo.GetBankAccountByID(request.Request.DestinationBankAccountID)
			if errRBA != nil {
				return fmt.Errorf(
					"getting receiver bank account: %w",
					errRBA,
				)
			}

			updateReceiverBankAccount, errURBA := repo.UpdateBankAccount(
				receiverBankAccount.ID,
				&dto.UpdateBankAccountRequest{
					Balance: receiverBankAccount.Balance.Add(totalAmount),
				},
			)
			if errURBA != nil {
				return fmt.Errorf(
					"update receiver bank account: %w",
					errRBA,
				)
			}

			_, errLT := a.Services.PostTransaction(
				tx,
				&dto.PostLedgerTransactionRequest{
					ReferenceID:      payout.PayoutReference,
					Type:             constants.LedgerAccountTypeBankAccount,
					Status:           constants.TransactionStatusCompleted,
					SettlementStatus: constants.TransactionStatusCompleted,
					Currency:         "INR",
				},
				[]*models.LedgerEntry{
					{
						AccountType: constants.LedgerAccountTypeBankAccount,
						AccountID:   updatedSenderBankAccount.ID,
						EntryType:   constants.LedgerEntryTypeDebit,
						Amount:      totalAmount,
						Currency:    "INR",
					},
					{
						AccountType: constants.LedgerAccountTypeBankAccount,
						AccountID:   updateReceiverBankAccount.ID,
						EntryType:   constants.LedgerEntryTypeCredit,
						Amount:      totalAmount,
						Currency:    "INR",
					},
				},
			)
			if errLT != nil {
				return fmt.Errorf(
					"create wallet to payout vault ledger: %w",
					errLT,
				)
			}

			updatedPayout, err := repo.UpdatePayoutStatus(
				payout.PayoutReference,
				constants.TransactionStatusCompleted,
			)
			if err != nil {
				return fmt.Errorf(
					"update payment status to completed: %w",
					err,
				)
			}

			finalPayoutAfterUpdate = updatedPayout
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return finalPayoutAfterUpdate, nil
}
