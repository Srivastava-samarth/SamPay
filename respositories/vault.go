package repositories

import (
	"errors"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (vr *Repository) CreateVault(
	request *dto.CreateVaultRequest,
) (*models.Vault, error) {

	vault := &models.Vault{
		ID:        utils.GenerateUUID(),
		Type:      request.Type,
		Status:    request.Status,
		Balance:   request.Balance,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := vr.DB.Create(vault).Error; err != nil {
		return nil, err
	}

	return vault, nil
}

func (vr *Repository) GetVaults() ([]*models.Vault, error) {
	var vaults []*models.Vault
	if err := vr.DB.Find(&vaults).Error; err != nil {
		return nil, err
	}

	return vaults, nil
}

func (vr *Repository) GetVault(
	vaultId uuid.UUID,
	) (*models.Vault, error) {
	var vault *models.Vault
	if err := vr.DB.Where(
		"id = ?",
		vaultId,
	).First(&vault).Error; err != nil {
		return nil, err
	}

	return vault, nil
}

func (vr *Repository) UpdateVaultBalance(
	balance decimal.Decimal, 
	vaultType string,
	) (*models.Vault, error) {
	if balance.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("balance should be greater than zero")
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
		"balance":    balance,
	}

	err := vr.DB.
		Model(&models.Vault{}).
		Where("type = ? AND status = ?", vaultType, constants.VaultStatusActive).
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

	var vault *models.Vault
	errF := vr.DB.Where("type = ?", vaultType).First(&vault).Error
	if errF != nil {
		return nil, errF
	}

	return vault, nil

}

func (vr *Repository) GetVaultByType(
	vaultType string,
	) (*models.Vault, error) {
	var vault *models.Vault
	if err := vr.DB.Where(
		"type = ? AND status = ?",
		vaultType, constants.VaultStatusActive,
	).First(&vault).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return vault, nil
}

func (vr *Repository) UpdateVaultStatus(
	status string, 
	vaultID uuid.UUID,
	) (*models.Vault, error) {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
		"status":     status,
	}

	err := vr.DB.
		Model(&models.Vault{}).
		Where("ID = ?", vaultID).
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

	var vault *models.Vault
	errF := vr.DB.Where("id = ?", vaultID).First(&vault).Error
	if errF != nil {
		return nil, errF
	}

	return vault, nil
}
