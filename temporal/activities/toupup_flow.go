package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/security"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)


func (r *Registry) GetWalletByMerchantID(
	ctx context.Context,
	merchantID uuid.UUID,
) (*models.Wallets, error){
	return r.Repo.GetWalletByMerchantId(merchantID)
}

func (r *Registry) GetBankAccountByMerchantID(
	ctx context.Context,
	merchantID uuid.UUID,
) (*models.LinkedBankAccount, error){
	return r.Repo.GetPrimaryBankAccountLinkedByMerchantID(merchantID)
}

func (a *Registry) PublishBankToWalletTransferEvent(
	ctx context.Context,
	request *dto.CreateBankToWalletDebitEvent,
	workflowID string,
) (*models.Event, error) {
	payload := request.Payload
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal response payload: %w", err)
	}

	newEvent, errNE := a.CreateEvent(context.Background(), &dto.CreateEventRequest{
		EventID:    request.EventID,
		EventType:  request.EventType,
		Direction:  string(constants.EventDirectionOutbound),
		Payload:    p,
		WorkflowID: workflowID,
		Status:     string(constants.EventStatusProcessing),
	})
	if errNE != nil {
		return nil, errNE
	}

	jsonPayload := datatypes.JSON(p)

	errP := a.nats.Publish(context.Background(), constants.EventType(newEvent.EventType), security.EventEnvelope{
		EventID:   newEvent.EventID,
		EventType: newEvent.EventType,
		Payload:   json.RawMessage(jsonPayload),
	})

	if errP != nil {
		return nil, errP
	}

	updatedEvent, errUE := a.UpdateEventStatus(context.Background(), newEvent.EventID, string(constants.EventStatusProcessed))
	if errUE != nil {
		return nil, errUE
	}

	return updatedEvent, nil
}

func (a *Registry) ExecuteBankToWallet(
	ctx context.Context,
	event *models.Event,
	merchantID uuid.UUID,
	request *dto.CreateBankToWalletRequest,
) (*models.Wallets, error){
	var finalWallet *models.Wallets
	err := a.DB.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			repo := a.Repo.WithTx(tx)
			updatedEvent, errUE := repo.UpdateEventStatus(ctx, event.EventID, string(constants.EventStatusProcessing))
			if errUE != nil{
				return errUE
			}

			wallet, errW := repo.GetWalletByMerchantId(merchantID)
			if errW != nil{
				return nil
			}

			companyVault, errCV := repo.GetVaultByType(constants.CompanyVault)
			if errCV != nil{
				return errCV
			}

			updatedWalletBalance := wallet.AvailableBalance.Add(request.Amount)
			updatedWallet, errUW := repo.UpdateWalletBalance(merchantID, &dto.UpdateWalletBalanceRequest{
				AvailableBalance: &updatedWalletBalance,
			})
			if errUW != nil{
				return errUW
			}

			updateCompanyVaultBalance := companyVault.Balance.Sub(request.Amount)
			updatedCompanyVault, errUCV := repo.UpdateVaultBalance(updateCompanyVaultBalance, constants.CompanyVault)
			if errUCV != nil{
				return errUCV
			}

			_, err := a.Services.PostTransaction(
				tx,
				&dto.PostLedgerTransactionRequest{
					ReferenceID:      utils.GenerateAutoTopUpReference(),
					Type:             constants.LedgerTransactionTypeTopup,
					Status:           constants.TransactionStatusCompleted,
					SettlementStatus: constants.LedgerSettlementSettled,
					Currency:         "INR",
				},
				[]*models.LedgerEntry{
					{
						AccountType: constants.LedgerAccountTypeVault,
						AccountID:   updatedCompanyVault.ID,
						EntryType:   constants.LedgerEntryTypeDebit,
						Amount:      request.Amount,
						Currency:    "INR",
					},
					{
						AccountType: constants.LedgerAccountTypeWallet,
						AccountID:   updatedWallet.ID,
						EntryType:   constants.LedgerEntryTypeCredit,
						Amount:      request.Amount,
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

			_, errNUE := repo.UpdateEventStatus(ctx, updatedEvent.EventID, string(constants.EventStatusProcessed))
			if errNUE != nil{
				return errNUE
			}

			finalWallet = updatedWallet
			return nil
		},
	)
	
	if err != nil{
		return nil, err
	}

	return finalWallet, nil
}