package services

import (
	"errors"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/google/uuid"
)

func (ws *Services) CreateWalletForMerchant(
	merchantID uuid.UUID,
	) (*models.Wallets, error) {
	if merchantID == uuid.Nil {
		return nil, errors.New("merchant id is required")
	}

	existingWallet, errEW := ws.Repo.GetWalletByMerchantId(merchantID)
	if errEW != nil {
		return nil, errEW
	}

	if existingWallet != nil {
		return nil, errors.New("wallet already exists for this merchant")
	}

	wallet, walletErr := ws.Repo.CreateWalletForMerchant(merchantID)
	if walletErr != nil {
		return nil, walletErr
	}
	return wallet, nil
}

func (ws *Services) GetWalletByMerchantID(
	merchantID uuid.UUID,
	) (*models.Wallets, error) {
	if merchantID == uuid.Nil {
		return nil, errors.New("merchant id is required")
	}

	wallet, errW := ws.Repo.GetWalletByMerchantId(merchantID)
	if errW != nil {
		return nil, errW
	}
	return wallet, nil
}

func (ws *Services) UpdateWalletStatus(
	merchantId uuid.UUID, 
	status string,
	) (*models.Wallets, error) {
	if merchantId == uuid.Nil {
		return nil, errors.New("merchant id is required")
	}
	updatedWallet, errUW := ws.Repo.UpdateWalletStatus(merchantId, status)
	if errUW != nil {
		return nil, errUW
	}

	return updatedWallet, nil
}
