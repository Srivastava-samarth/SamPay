package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"golang.org/x/crypto/bcrypt"
)

func (as *Services) Authentication(
	authRequest *dto.AuthRequest,
	) (*dto.AuthResponse, error) {
	var authResponse *dto.AuthResponse

	user, errU := as.Repo.GetUserByEmail(authRequest.Email)
	if errU != nil {
		return nil, errU
	}

	if user == nil {
		return nil, errors.New("user doen not exist")
	}

	if user.Status != constants.UserStatusActive {
		return nil, errors.New("user not valid")
	}

	if user.MustChangePassword {
		return nil, errors.New("temporary password can't be used. Please reset the password")
	}

	merchantUser, errMU := as.Repo.GetMerchantUserByUserID(user.ID)
	if errMU != nil {
		return nil, errMU
	}

	errCHP := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(authRequest.Password),
	)

	if errCHP != nil {
		return nil, errors.New("invalid Credentials: email or password is incorrect")
	}

	token, errT := as.JwtService.GenarateTokenAndExpiry(merchantUser.UserID, merchantUser.MerchantID, merchantUser.Role)
	if errT != nil {
		return nil, errT
	}

	refreshToken, errRT := as.JwtService.GenerateRefreshToken()
	if errRT != nil {
		return nil, errRT
	}

	hashedRefreshToken := utils.HashToken(refreshToken)
	refreshTokenExpirySeconds, errRTE := strconv.ParseInt(
		as.JwtService.Config.RefreshExpiry,
		10,
		64,
	)
	if errRTE != nil {
		return nil, errRTE
	}

	refreshTokenExpiry := time.Now().Add(
		time.Duration(refreshTokenExpirySeconds) * time.Second,
	)

	userSessionRequestPayload := &models.UserSession{
		UserID:           merchantUser.UserID,
		RefreshTokenHash: hashedRefreshToken,
		ExpiresAt:        refreshTokenExpiry,
	}
	_, errUS := as.Repo.CreateUserSession(userSessionRequestPayload)
	if errUS != nil {
		return nil, errUS
	}

	authResponse = &dto.AuthResponse{
		AccessToken:  token,
		RefreshToken: hashedRefreshToken,
		ExpiresIn:    int(refreshTokenExpirySeconds),
	}

	return authResponse, nil

}

func (as *Services) ForgotPasswod(
	forgotPasswordRequest *dto.ForgotPasswordRequest,
	) error {
	start := time.Now()

	fmt.Printf(
		"ForgotPassword START %s\n",
		start.Format("15:04:05.000"),
	)

	defer func() {
		fmt.Printf(
			"ForgotPassword END %s duration=%v\n",
			time.Now().Format("15:04:05.000"),
			time.Since(start),
		)
	}()
	user, errU := as.Repo.GetUserByEmail(forgotPasswordRequest.Email)
	if errU != nil {
		return errU
	}

	if user == nil {
		return errors.New("user doen not exist")
	}

	resetToken, errRT := as.JwtService.GenerateResetPasswordToken(user.ID, user.Email)
	if errRT != nil {
		return errRT
	}

	hashedResetToken := utils.HashToken(resetToken)

	resetTokenExpirySeconds, errRTE := strconv.ParseInt(
		as.JwtService.Config.AccessExpiry,
		10,
		64,
	)
	if errRTE != nil {
		return errRTE
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

	_, errPR := as.Repo.CreatePasswordReset(passwordResetPayload)
	if errPR != nil {
		return errPR
	}

	errN := as.NotificationService.SendResetPasswordEmail(user.FirstName, user.Email, resetToken)
	if errN != nil {
		return errN
	}
	return nil
}

func (as *Services) ResetPassword(
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

	passwordReset, errP := as.Repo.FindByToken(
		resetPasswordRequest.ResetToken,
	)
	if errP != nil {
		return errP
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

	hashedPassword, errHP := utils.HashPassword(
		resetPasswordRequest.NewPassword,
	)
	if errHP != nil {
		return errHP
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

	repo := as.Repo.WithTx(tx)

	mustChangePassword := false
	updateUserRequest := &dto.UpdateUserRequest{
		PasswordHash:       hashedPassword,
		MustChangePassword: &mustChangePassword,
	}

	_, err = repo.UpdateUser(userID, updateUserRequest)
	if err != nil {
		tx.Rollback()
		return err
	}

	updatePasswordResetRequest := &models.PasswordResetToken{
		ID:     passwordReset.ID,
		UsedAt: time.Now(),
	}

	_, err = repo.UpdatePasswordReset(
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

func (as *Services) RefreshToken(
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

	repo := as.Repo.WithTx(tx)

	userSession, err := repo.
		GetUserSessionByRefreshTokenHash(refreshTokenRequest.RefreshToken)

	if err != nil {
		return nil, err
	}

	now := time.Now()

	if now.After(userSession.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	user, err := repo.GetUserByID(userSession.UserID)
	if err != nil {
		return nil, err
	}

	merchantUser, err := repo.GetMerchantUserByUserID(userSession.UserID)

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

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	committed = true

	return &dto.AuthResponse{
		AccessToken: accessToken,
	}, nil
}
