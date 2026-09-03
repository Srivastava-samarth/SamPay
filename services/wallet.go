package services

import (
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/google/uuid"
)

type WalletService struct{
	WalletRepo *repositories.WalletRepository
}

func NewWalletService(
	WalletRepo *repositories.WalletRepository,
) *WalletService{
	return &WalletService{
		WalletRepo: WalletRepo,
	}
}

func(ws *WalletService) CreateWalletForMerchant(merchantID uuid.UUID) (*models.Wallets, error) {
	wallet, walletErr := ws.WalletRepo.CreateWalletForMerchant(merchantID)
	if walletErr != nil {
		return nil, walletErr
	}
	return wallet, nil
}

func(ws *WalletService) GetWalletByMerchantID(merchantID uuid.UUID) (*models.Wallets, error){
	wallet, errW := ws.WalletRepo.GetWalletByMerchantId(merchantID)
	if errW != nil{
		return nil, errW
	}
	return wallet, nil
}