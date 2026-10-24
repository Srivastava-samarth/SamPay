package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/security"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func(a *Registry) GetVault(
	ctx context.Context,
	vaultType string,
) (*models.Vault, error){
	return a.Repo.GetVaultByType(vaultType)
}

func(a *Registry) PublishVaultTopupEvent(
	ctx context.Context,
	request *dto.CreateVaultTopupEvent,
	workflowID string,
)(*models.Event, error){
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

func(a *Registry) ExecuteVaultTopUp(
	ctx context.Context,
	vault *models.Vault,
	event *models.Event,
	request *dto.VaultTopupRequest,
) (*models.Vault, error){
	var finalVault *models.Vault
	err := a.DB.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			repo := a.Repo.WithTx(tx)

			updatedEvent, errUE := repo.UpdateEventStatus(ctx, event.EventID, string(constants.EventStatusProcessing))
			if errUE != nil{
				return errUE
			}

			updatedVaultBalance := vault.Balance.Add(request.Amount)
			updatedVault, errUV := repo.UpdateVaultBalance(updatedVaultBalance, vault.Type)
			if errUV != nil{
				return errUV
			}

			sampayBankAccountID, errP := uuid.Parse(a.InternalConfig.SampayBankAccountID)
			if errP != nil{
				return errP
			}

			_, err := a.Services.PostTransaction(
				tx,
				&dto.PostLedgerTransactionRequest{
					ReferenceID:      uuid.NewString(),
					Type:             constants.LedgerTransactionTypeTopup,
					Status:           constants.TransactionStatusCompleted,
					SettlementStatus: constants.LedgerSettlementSettled,
					Currency:         "INR",
				},
				[]*models.LedgerEntry{
					{
						AccountType: constants.LedgerAccountTypeVault,
						AccountID:   updatedVault.ID,
						EntryType:   constants.LedgerEntryTypeCredit,
						Amount:      request.Amount,
						Currency:    "INR",
					},
					{
						AccountType: constants.LedgerAccountTypeBankAccount,
						AccountID:   sampayBankAccountID,
						EntryType:   constants.LedgerEntryTypeDebit,
						Amount:      request.Amount,
						Currency:    "INR",
					},
				},
			)
			if err != nil {
				return fmt.Errorf(
					"create vault topup ledger: %w",
					err,
				)
			}

			_, errUNE := repo.UpdateEventStatus(ctx, updatedEvent.EventID, string(constants.EventStatusProcessed))
			if errUNE != nil{
				return errUNE
			}

			finalVault = updatedVault

			return nil
		},
	)
	if err != nil{
		return nil, err
	}

	return finalVault,nil
}