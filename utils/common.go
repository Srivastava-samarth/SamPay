package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"

	"github.com/Srivastava-samarth/sampay/constants"
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
	ref := "pmt_" + hex.EncodeToString(id[:])[:12]
	return &ref
}

func GeneratePayoutReference() *string {
	id := uuid.New()
	ref := "pyt_" + hex.EncodeToString(id[:])[:12]
	return &ref
}

func GenerateRefundReference() *string {
	id := uuid.New()
	ref := "rfd_" + hex.EncodeToString(id[:])[:12]
	return &ref
}

func GenerateCustomerReference() *string {
	id := uuid.New()
	ref := "cust_" + hex.EncodeToString(id[:])[:12]
	return &ref
}

func GenerateBankAccountNumber() string {
	min := big.NewInt(1_000_000_000_000)
	max := big.NewInt(9_999_999_999_999)

	rangeSize := new(big.Int).Sub(max, min)
	rangeSize.Add(rangeSize, big.NewInt(1))

	n, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		panic("failed to generate bank account number")
	}

	n.Add(n, min)

	return n.String()
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

func IsValidVaultStatus(status string) bool {
	switch status {
	case constants.VaultStatusActive,
		constants.VaultStatusInactive:
		return true
	default:
		return false
	}
}

func IsValidVaultType(vaultType string) bool {
	switch vaultType {
	case constants.PaymentVault,
		constants.PayoutVault,
		constants.CompanyVault:
		return true
	default:
		return false
	}
}

func IsValidMerchantStatus(status string) bool {
	switch status {
	case constants.MerchantStatusActive,
		constants.MerchantStatusPending,
		constants.MerchantStatusSuspended,
		constants.MerchantStatusInactive:
		return true
	default:
		return false
	}
}
