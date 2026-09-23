package middlewares

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
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
	jti := uuid.New().String()

	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"type":  "password_reset",
		"jti":   jti,
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

func (j *Jwt) ValidateResetPaasword(
	tokenString string,
) (uuid.UUID, string, error) {
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

func (j *Jwt) ValidateAccessToken(
	tokenString string,
) (uuid.UUID, uuid.UUID, string, error) {

	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {

			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("invalid signing method")
			}

			return []byte(j.Config.Secret), nil
		},
	)

	if err != nil {
		return uuid.Nil, uuid.Nil, "", err
	}

	if !token.Valid {
		return uuid.Nil, uuid.Nil, "", errors.New("invalid access token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, uuid.Nil, "", errors.New("invalid token claims")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return uuid.Nil, uuid.Nil, "", errors.New("invalid access token type")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, uuid.Nil, "", errors.New("invalid user id")
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, uuid.Nil, "", errors.New("invalid user id")
	}

	merchantIDString, ok := claims["merchant_id"].(string)
	if !ok {
		return uuid.Nil, uuid.Nil, "", errors.New("invalid merchant id")
	}

	merchantID, err := uuid.Parse(merchantIDString)
	if err != nil {
		return uuid.Nil, uuid.Nil, "", errors.New("invalid merchant id")
	}

	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return uuid.Nil, uuid.Nil, "", errors.New("invalid role")
	}

	return userID, merchantID, role, nil
}
