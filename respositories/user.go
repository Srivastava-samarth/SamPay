package repositories

import (
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
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
	err := ur.db.Where("email = ? AND status = ?", email, constants.MerchantStatusSuspended).First(&user).Error
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

func (ur *UserRepository) UpdateUser(request *models.User) (*models.User, error) {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if request.PasswordHash != "" {
		updates["password_hash"] = request.PasswordHash
	}

	if request.FirstName != "" {
		updates["first_name"] = request.FirstName
	}

	if request.LastName != "" {
		updates["last_name"] = request.LastName
	}

	if request.MustChangePassword == true || request.MustChangePassword == false {
		updates["must_change_password"] = request.MustChangePassword
	}

	err := ur.db.
		Model(&models.User{}).
		Where("id = ?", request.ID).
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

	var updatedUser models.User
	err = ur.db.
		Where("id = ?", request.ID).
		First(&updatedUser).Error

	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}
func(ur *UserRepository) GetUserByID(userId uuid.UUID) (*models.User, error){
	var user *models.User
	err := ur.db.Where("id= ?", userId).First(&user).Error 
	if err != nil{
		return nil, err
	}
	return user, nil
}

func (ur *UserRepository) GetUsers() ([]*models.User, error) {
	var users []*models.User

	if err := ur.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (ur *UserRepository) UpdateUserStatus(status string, userID uuid.UUID) (*models.User, error){
	updates := map[string]interface{}{
		"status":status,
		"updated_at": time.Now(),
	}

	err := ur.db.
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

	var user *models.User
	errU := ur.db.Where("id = ?", userID).First(&user).Error
	if errU != nil{
		return nil, errU
	}

	return user, nil
}
