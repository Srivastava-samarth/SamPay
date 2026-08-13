package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm"
)

type UserRepository struct{
	db * gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository{
	return &UserRepository{
		db: db,
	}
}

func (wr *UserRepository) WithTx(tx *gorm.DB) *UserRepository {
	return &UserRepository{
		db: tx,
	}
}

func(ur *UserRepository) GetBlockedUserByEmail(email string) (bool, error) {
	var user models.User
	err := ur.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func(ur *UserRepository) CreateUser(request *models.User) (*models.User, error) {
	user := &models.User{
		ID:           utils.GenerateUUID(),
		Email:        request.Email,
		PasswordHash: request.PasswordHash,
		FirstName:    request.FirstName,
		LastName:     request.LastName,
		Status:       request.Status,
		MustChangePassword: true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := ur.db.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func(ur *UserRepository) GetUserByEmail(email string) (*models.User, error){
	var user *models.User
	err := ur.db.Where("email= ?",email).First(&user).Error 
	if err != nil{
		return nil, err
	}
	return user, nil
}
