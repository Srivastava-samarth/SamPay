package services

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/utils"
)

type UserService struct{
	UserRepo *repositories.UserRepository
}

func NewUserService(
	UserRepo *repositories.UserRepository,
) *UserService{
	return &UserService{
		UserRepo: UserRepo,
	}
}

type CreatedUserResult  struct {
	User              *dto.CreateUserResponse
	TemporaryPassword string
}

func(us *UserService) CreateUser(userRequest *dto.CreateUserRequest) (*CreatedUserResult, error) {
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
		Status:       "active",
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
