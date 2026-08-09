package repositories

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm"
)

func GetBlockedUserByEmail(email string, db *gorm.DB) (bool, error) {
	var user models.User
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func CreateUser(request models.User, db *gorm.DB) (*models.User, error) {
	user := &models.User{
		ID:           utils.GenerateUUID(),
		Email:        request.Email,
		PasswordHash: request.PasswordHash,
		FirstName:    request.FirstName,
		LastName:     request.LastName,
		Status:       request.Status,
		CreatedAt:    request.CreatedAt,
		UpdatedAt:    request.UpdatedAt,
	}
	err := db.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}
