package utils

import (
	"encoding/base64"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type Jwt struct {
	Config *config.JWTConfig
}

func NewJwt(
	jwt *config.JWTConfig,
) *Jwt {
	return &Jwt{
		Config: jwt,
	}
}

func (j *Jwt) GenarateTokenAndExpiry(
	userID uuid.UUID,
	merchantID uuid.UUID,
	role string,
) (string, error) {
	expirySeconds, err := strconv.ParseInt(j.Config.AccessExpiry, 10, 64)
	if err != nil {
		return "", err
	}

	expiry := time.Now().Add(time.Duration(expirySeconds) * time.Second)
	claims := jwt.MapClaims{
		"sub":         userID.String(),
		"merchant_id": merchantID.String(),
		"role":        role,
		"type":        "access",
		"exp":         expiry.Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(j.Config.Secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (j *Jwt) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func (j *Jwt) GenerateResetPasswordToken(
	userID uuid.UUID,
	email string,
) (string, error) {
	expirySeconds, err := strconv.ParseInt(j.Config.AccessExpiry, 10, 64)
	if err != nil {
		return "", err
	}

	expiry := time.Now().Add(time.Duration(expirySeconds) * time.Second)
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"type":  "password_reset",
		"exp":   expiry.Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(j.Config.Secret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func(j *Jwt) ValidateResetPaasword(
	tokenString string,
) (uuid.UUID, string, error){
	token, errT := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {

			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("invalid signing method")
			}

			return []byte(j.Config.Secret), nil
		},
	)

	if errT != nil {
		return uuid.Nil, "", errT
	}

	if !token.Valid {
		return uuid.Nil, "", errors.New("invalid reset token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, "", errors.New("invalid token claims")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "password_reset" {
		return uuid.Nil, "", errors.New("invalid reset token type")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, "", errors.New("invalid user id in token")
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, "", errors.New("invalid user id in token")
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return uuid.Nil, "", errors.New("invalid email in token")
	}

	return userID, email, nil

}