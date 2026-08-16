package utils

import (
	"encoding/base64"
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

func(j *Jwt) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
