package services

import (
	"errors"
	"strconv"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/Srivastava-samarth/sampay/notifications"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	DB                  *gorm.DB
	UserRepo            *repositories.UserRepository
	UserSessionRepo     *repositories.UserSessionRepository
	MerchantUserRepo    *repositories.MerchantUserRepository
	JwtService          middlewares.Jwt
	NotificationService notifications.EmailService
	PasswordResetRepo   *repositories.PasswordResetTokenRepository
}

func NewAuthService(
	db *gorm.DB,
	userRepo *repositories.UserRepository,
	userSessionRepo *repositories.UserSessionRepository,
	merchantUserRepo *repositories.MerchantUserRepository,
	jwtService *middlewares.Jwt,
	notification *notifications.EmailService,
	passwordResetRepo *repositories.PasswordResetTokenRepository,
) *AuthService {
	return &AuthService{
		DB:                  db,
		UserRepo:            userRepo,
		UserSessionRepo:     userSessionRepo,
		MerchantUserRepo:    merchantUserRepo,
		JwtService:          *jwtService,
		NotificationService: *notification,
		PasswordResetRepo:   passwordResetRepo,
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

	errCHP := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(authRequest.Password),
	)

	if errCHP != nil {
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
		UserID:    user.ID,
		ExpiresAt: resetTokenExpiry,
		TokenHash: hashedResetToken,
		CreatedAt: time.Now(),
	}

	_, errPR := as.PasswordResetRepo.CreatePasswordReset(passwordResetPayload)
	if errPR != nil {
		return errPR
	}

	errN := as.NotificationService.SendResetPasswordEmail(user.FirstName, user.Email, resetToken)
	if errN != nil {
		return errN
	}
	return nil
}

func (as *AuthService) ResetPassword(
	resetPasswordRequest *dto.ResetPasswordRequest,
) error {
	userID, email, err := as.JwtService.ValidateResetPaasword(
		resetPasswordRequest.ResetToken,
	)
	if err != nil {
		return err
	}

	if resetPasswordRequest.Email != email {
		return errors.New("email does not match reset token")
	}

	passwordReset, err := as.PasswordResetRepo.FindByToken(
		resetPasswordRequest.ResetToken,
	)
	if err != nil {
		return err
	}

	if passwordReset.UserID != userID {
		return errors.New("reset token is not valid")
	}

	if !passwordReset.UsedAt.IsZero() {
		return errors.New("reset token has already been used")
	}

	if time.Now().After(passwordReset.ExpiresAt) {
		return errors.New("reset token expired")
	}

	if resetPasswordRequest.NewPassword !=
		resetPasswordRequest.ConfirmPassword {
		return errors.New("passwords do not match")
	}

	hashedPassword, err := utils.HashPassword(
		resetPasswordRequest.NewPassword,
	)
	if err != nil {
		return err
	}

	tx := as.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	userRepo := as.UserRepo.WithTx(tx)
	passwordResetRepo := as.PasswordResetRepo.WithTx(tx)

	updateUserRequest := &models.User{
		ID:                 userID,
		PasswordHash:       hashedPassword,
		MustChangePassword: false,
	}

	_, err = userRepo.UpdateUser(updateUserRequest)
	if err != nil {
		tx.Rollback()
		return err
	}

	updatePasswordResetRequest := &models.PasswordResetToken{
		ID:     passwordReset.ID,
		UsedAt: time.Now(),
	}

	_, err = passwordResetRepo.UpdatePasswordReset(
		updatePasswordResetRequest,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (as *AuthService) RefreshToken(
	refreshTokenRequest *dto.RefreshTokenRequest,
) (*dto.AuthResponse, error) {

	tx := as.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	committed := false

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}

		if !committed {
			tx.Rollback()
		}
	}()

	userSessionRepo := as.UserSessionRepo.WithTx(tx)
	userRepo := as.UserRepo.WithTx(tx)
	merchantUserRepo := as.MerchantUserRepo.WithTx(tx)

	userSession, err := userSessionRepo.
		GetUserSessionByRefreshTokenHash(refreshTokenRequest.RefreshToken)

	if err != nil {
		return nil, err
	}

	now := time.Now()

	if now.After(userSession.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	if userSession.RevokedAt != nil {
	return nil, errors.New("refresh token revoked")
}

	user, err := userRepo.GetUserByID(userSession.UserID)
	if err != nil {
		return nil, err
	}

	merchantUser, err := merchantUserRepo.GetMerchantUserByUserID(userSession.UserID)

	if err != nil {
		return nil, err
	}

	_, err = userSessionRepo.UpdateUserSession(
		&models.UserSession{
			ID:        userSession.ID,
			ExpiresAt: time.Now(),
			RevokedAt: &now,
		},
	)

	if err != nil {
		return nil, err
	}

	accessToken, err := as.JwtService.GenarateTokenAndExpiry(
		user.ID,
		merchantUser.MerchantID,
		merchantUser.Role,
	)

	if err != nil {
		return nil, err
	}

	newRefreshToken, err :=
		as.JwtService.GenerateRefreshToken()

	if err != nil {
		return nil, err
	}

	hashedNewRefreshToken :=
		utils.HashToken(newRefreshToken)

	refreshTokenExpirySeconds, err :=
		strconv.ParseInt(
			as.JwtService.Config.RefreshExpiry,
			10,
			64,
		)

	if err != nil {
		return nil, err
	}

	refreshTokenExpiry := now.Add(
		time.Duration(refreshTokenExpirySeconds) * time.Second,
	)

	_, err = userSessionRepo.CreateUserSession(
		&models.UserSession{
			UserID:           userSession.UserID,
			RefreshTokenHash: hashedNewRefreshToken,
			ExpiresAt:        refreshTokenExpiry,
		},
	)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	committed = true

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int(refreshTokenExpirySeconds),
	}, nil
}
