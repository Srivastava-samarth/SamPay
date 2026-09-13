package services

import (
	"errors"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
)

type UserService struct {
	UserRepo         *repositories.UserRepository
	MerchantUserRepo *repositories.MerchantUserRepository
}

func NewUserService(
	userRepo *repositories.UserRepository,
	merchantUserRepo *repositories.MerchantUserRepository,
) *UserService {
	return &UserService{
		UserRepo:         userRepo,
		MerchantUserRepo: merchantUserRepo,
	}
}

type CreatedUserResult struct {
	User              *dto.CreateUserResponse
	TemporaryPassword string
}

func (us *UserService) CreateUser(userRequest *dto.CreateUserRequest) (*CreatedUserResult, error) {
	password, errP := utils.GenerateTemporaryPassword(8)
	if errP != nil {
		return nil, errP
	}

	passwordHash, errPH := utils.HashPassword(password)
	if errPH != nil {
		return nil, errPH
	}

	createUserRequestPayload := &models.User{
		Email:        userRequest.Email,
		FirstName:    userRequest.FirstName,
		LastName:     userRequest.LastName,
		PasswordHash: passwordHash,
	}

	user, err := us.UserRepo.CreateUser(createUserRequestPayload)
	if err != nil {
		return nil, err
	}

	userResponse := &CreatedUserResult{
		User: &dto.CreateUserResponse{
			ID:                 user.ID,
			Email:              user.Email,
			FirstName:          user.FirstName,
			LastName:           user.LastName,
			Status:             user.Status,
			MustChangePassword: user.MustChangePassword,
		},
		TemporaryPassword: password,
	}

	return userResponse, nil
}

func (us *UserService) GetUser(userId uuid.UUID) (*models.User, error) {
	user, errU := us.UserRepo.GetUserByID(userId)
	if errU != nil {
		return nil, errU
	}

	return user, nil
}

func (us *UserService) GetUsers() ([]*models.User, error) {
	users, errU := us.UserRepo.GetUsers()
	if errU != nil {
		return nil, errU
	}
	return users, nil
}

func (us *UserService) UpdateUser(
	request *dto.UpdateUserRequest,
	userId uuid.UUID,
) (*models.User, error) {
	updatedUserPayload := &models.User{
		ID: userId,
	}

	if request.FirstName != "" {
		updatedUserPayload.FirstName = request.FirstName
	}

	if request.LastName != "" {
		updatedUserPayload.LastName = request.LastName
	}

	if request.PasswordHash != "" {
		updatedUserPayload.PasswordHash = request.PasswordHash
	}

	if request.MustChangePassword {
		updatedUserPayload.MustChangePassword = request.MustChangePassword
	}

	return us.UserRepo.UpdateUser(updatedUserPayload)
}

func (us *UserService) UpdateUserStatus(status string, userID uuid.UUID) (*models.User, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user_id is required")
	}

	if status == "" {
		return nil, errors.New("status is required")
	}

	if !utils.IsValidMerchantStatus(status) {
		return nil, errors.New("invalid user status")
	}

	user, errU := us.UserRepo.GetUserByID(userID)
	if errU != nil {
		return nil, errU
	}

	if user.Status == status {
		return user, nil
	}

	updatedUser, errUU := us.UserRepo.UpdateUserStatus(status, userID)
	if errUU != nil {
		return nil, errUU
	}

	return updatedUser, nil
}

func (us *UserService) GetUsersByMerchant(merchantID uuid.UUID) ([]*models.User, error) {
	merchantUsers, errMU := us.MerchantUserRepo.GetMerchantUsersByMerchantID(merchantID)
	if errMU != nil {
		return nil, errMU
	}

	var users []*models.User
	for _, merchantUser := range merchantUsers {
		user, errU := us.UserRepo.GetUserByID(merchantUser.UserID)
		if errU != nil {
			return nil, errU
		}

		users = append(users, user)
	}

	return users, nil
}
