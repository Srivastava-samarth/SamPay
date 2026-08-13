package services

import (
	"errors"
	"strconv"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/utils"
)

type AuthService struct {
	UserRepo         *repositories.UserRepository
	UserSessionRepo  *repositories.UserSessionRepository
	MerchantUserRepo *repositories.MerchantUserRepository
	JwtService       utils.Jwt
}

func NewAuthService(
	userRepo *repositories.UserRepository,
	userSessionRepo *repositories.UserSessionRepository,
	merchantUserRepo *repositories.MerchantUserRepository,
	jwtService *utils.Jwt,
) *AuthService {
	return &AuthService{
		UserRepo:         userRepo,
		UserSessionRepo:  userSessionRepo,
		MerchantUserRepo: merchantUserRepo,
		JwtService:       *jwtService,
	}
}

func (as *AuthService) Authentication(authRequest *dto.AuthRequest) (*dto.AuthResponse, error) {
	var authResponse *dto.AuthResponse

	user, err := as.UserRepo.GetUserByEmail(authRequest.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("User doen not exist !")
	}

	if user.MustChangePassword == true {
		return nil, errors.New("Temporary password can't be used. Please reset the password")
	}

	merchantUser, err := as.MerchantUserRepo.GetMerchantUserByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	password, err := utils.HashPassword(authRequest.Password)
	if err != nil {
		return nil, err
	}

	if password != user.PasswordHash {
		return nil, errors.New("Invalid Credentials: email or password is incorrect")
	}

	token, err := as.JwtService.GenarateTokenAndExpiry(merchantUser.UserID, merchantUser.MerchantID, merchantUser.Role)
	refreshToken, err := as.JwtService.GenerateRefreshToken()
	hashedRefreshToken, err := utils.HashPassword(refreshToken)
	if err != nil {
		return nil, err
	}

	refreshTokenExpirySeconds, err := strconv.ParseInt(
		as.JwtService.Config.RefreshExpiry,
		10,
		64,
	)
	if err != nil {
		return nil, err
	}

	refreshTokenExpiry := time.Now().Add(
		time.Duration(refreshTokenExpirySeconds) * time.Second,
	)

	userSessionRequestPayload := &models.UserSession{
		UserID:           merchantUser.UserID,
		RefreshTokenHash: hashedRefreshToken,
		ExpiresAt:        refreshTokenExpiry,
	}
	_, err = as.UserSessionRepo.CreateUserSession(userSessionRequestPayload)
	if err != nil {
		return nil, err
	}

	authResponse = &dto.AuthResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresIn:    int(refreshTokenExpirySeconds),
	}

	return authResponse, nil

}
