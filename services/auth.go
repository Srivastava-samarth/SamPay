package services

import (
	"errors"
	"strconv"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/notifications"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/utils"
)

type AuthService struct {
	UserRepo            *repositories.UserRepository
	UserSessionRepo     *repositories.UserSessionRepository
	MerchantUserRepo    *repositories.MerchantUserRepository
	JwtService          utils.Jwt
	NotificationService notifications.EmailService
	PasswordResetRepo   *repositories.PasswordResetTokenRepository
}

func NewAuthService(
	userRepo *repositories.UserRepository,
	userSessionRepo *repositories.UserSessionRepository,
	merchantUserRepo *repositories.MerchantUserRepository,
	jwtService *utils.Jwt,
	notification *notifications.EmailService,
	passwordResetRepo *repositories.PasswordResetTokenRepository,
) *AuthService {
	return &AuthService{
		UserRepo:            userRepo,
		UserSessionRepo:     userSessionRepo,
		MerchantUserRepo:    merchantUserRepo,
		JwtService:          *jwtService,
		NotificationService: *notification,
		PasswordResetRepo: passwordResetRepo,
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
	hashedRefreshToken := utils.HashToken(refreshToken)
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

func (as *AuthService) ForgotPasswod(forgotPasswordRequest *dto.ForgotPasswordRequest) error {
	user, errU := as.UserRepo.GetUserByEmail(forgotPasswordRequest.Email)
	if errU != nil {
		return errU
	}

	if user == nil {
		return errors.New("User doen not exist !")
	}

	resetToken, errRT := as.JwtService.GenerateResetPasswordToken(user.ID, user.Email)
	if errRT != nil {
		return errRT
	}

	hashedResetToken := utils.HashToken(resetToken)

	resetTokenExpirySeconds, err := strconv.ParseInt(
		as.JwtService.Config.AccessExpiry,
		10,
		64,
	)
	if err != nil {
		return err
	}

	resetTokenExpiry := time.Now().Add(
		time.Duration(resetTokenExpirySeconds) * time.Second,
	)

	passwordResetPayload := &models.PasswordResetToken{
		UserID: user.ID,
		ExpiresAt: resetTokenExpiry,
		TokenHash: hashedResetToken,
		UsedAt: nil,
		CreatedAt: time.Now(),
	}

	_, errPR := as.PasswordResetRepo.CreatePasswordReset(passwordResetPayload)
	if errPR != nil{
		return errPR
	}

	errN := as.NotificationService.SendResetPasswordEmail(user.FirstName, user.Email, resetToken)
	if errN != nil {
		return errN
	}
	return nil
}
