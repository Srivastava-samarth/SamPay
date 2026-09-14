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
	if userRequest == nil {
		return nil, errors.New("request is required")
	}

	if userRequest.Email == "" {
		return nil, errors.New("email is required")
	}

	if userRequest.FirstName == "" {
		return nil, errors.New("first name is required")
	}

	if userRequest.LastName == "" {
		return nil, errors.New("last name is required")
	}

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
	if userId == uuid.Nil {
		return nil, errors.New("user_id is required")
	}
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
	userID uuid.UUID,
) (*models.User, error) {
	if request.FirstName == "" {
		return nil, errors.New("first_name is required")
	}

	if request.LastName == "" {
		return nil, errors.New("last_name is required")
	}

	if request.PasswordHash == "" {
		return nil, errors.New("password_hash is required")
	}

	if !request.MustChangePassword {
		return nil, errors.New("must_change_password is required")
	}

	updatedUserPayload := &dto.UpdateUserRequest{
		FirstName:          request.FirstName,
		LastName:           request.LastName,
		PasswordHash:       request.PasswordHash,
		MustChangePassword: request.MustChangePassword,
	}

	return us.UserRepo.UpdateUser(userID, updatedUserPayload)
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
	if merchantID == uuid.Nil {
		return nil, errors.New("merchant_id is required")
	}
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
