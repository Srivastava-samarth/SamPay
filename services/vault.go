package services

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/google/uuid"
)

type VaultService struct {
	vaultRepo *repositories.VaultRepository
}

func NewVaultService(
	vaultRepo *repositories.VaultRepository,
) *VaultService{
	return &VaultService{
		vaultRepo: vaultRepo,
	}
}

func (vs *VaultService) CreateVault(request *dto.CreateVaultRequest) (*models.Vault, error){
	vault, errV := vs.vaultRepo.CreateVault(request)
	if errV != nil{
		return nil, errV
	}

	return vault, nil
}

func (vs *VaultService) GetVaults() ([]*models.Vault, error){
	vaults, errV := vs.vaultRepo.GetVaults()
	if errV != nil{
		return nil, errV
	}

	return vaults, nil
}

func (vs *VaultService) GetVaultByID(vaultId uuid.UUID) (*models.Vault, error){
	vault, errV := vs.vaultRepo.GetVault(vaultId)
	if errV != nil{
		return nil, errV
	}
	return vault, nil
}