package repositories

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VaultRepository struct {
	db *gorm.DB
}

func NewVaultRepository(
	db *gorm.DB,
) *VaultRepository{
	return &VaultRepository{
		db: db,
	}
}

func (vr *VaultRepository) CreateVault(
    request *dto.CreateVaultRequest,
) (*models.Vault, error) {

    vault := &models.Vault{
        ID:      utils.GenerateUUID(),
        Type:    request.Type,
        Status:  request.Status,
        Balance: request.Balance,
    }

    if err := vr.db.Create(vault).Error; err != nil {
        return nil, err
    }

    return vault, nil
}

func (vr *VaultRepository) GetVaults() ([]*models.Vault, error){
	var vaults []*models.Vault
	if err := vr.db.Find(&vaults).Error; err != nil{
		return nil, err
	}

	return vaults, nil
}

func (vr *VaultRepository) GetVault(vaultId uuid.UUID) (*models.Vault, error){
	var vault *models.Vault
	if err := vr.db.Where("id = ?", vaultId).First(&vault).Error; err != nil{
		return nil, err
	}

	return vault, nil
}