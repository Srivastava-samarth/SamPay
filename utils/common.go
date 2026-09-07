package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strconv"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

func GenerateUUID() uuid.UUID {
	return uuid.New()
}

func GenerateMerchantReference() string {
	id := uuid.New()
	return "mrc_" + hex.EncodeToString(id[:])[:12]
}

func GenerateLedgerReference() string {
	id := uuid.New()
	return "ldg_" + hex.EncodeToString(id[:])[:12]
}

func GenerateAutoTopUpReference() string {
	id := uuid.New()
	return "auto_topup_" + hex.EncodeToString(id[:])[:12]
}

func GeneratePaymentReference() *string {
	id := uuid.New()
	ref :=  "pmt_" + hex.EncodeToString(id[:])[:12]
	return &ref
}

func GeneratePayoutReference() *string {
	id := uuid.New()
	ref :=  "pyt_" + hex.EncodeToString(id[:])[:12]
	return &ref
}

func GenerateCustomerReference() *string {
	id := uuid.New()
	ref :=  "cust_" + hex.EncodeToString(id[:])[:12]
	return &ref
}

func GenerateBankAccountNumber() string {
    n := rand.Int63n(9000000000000) + 1000000000000
    accountNumber := strconv.FormatInt(n, 10)

    return accountNumber
}

func GenerateTemporaryPassword(length int) (string, error) {
	password := make([]byte, length)

	for i := 0; i < length; {
		var b [1]byte

		if _, err := rand.Read(b[:]); err != nil {
			return "", err
		}

		if int(b[0]) >= 256-(256%len(passwordChars)) {
			continue
		}

		password[i] = passwordChars[int(b[0])%len(passwordChars)]
		i++
	}

	return string(password), nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func HashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}