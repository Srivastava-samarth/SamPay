package utils

import (
	"encoding/hex"
	"math/rand"
	"strconv"

	"github.com/google/uuid"
)

func GenerateUUID() uuid.UUID {
	return uuid.New()
}

func GenerateMerchantReference() string {
	id := uuid.New()
	return "mrc_" + hex.EncodeToString(id[:])[:12]
}

func GenerateBankAccountNumber() string {
    n := rand.Int63n(9000000000000) + 1000000000000
    accountNumber := strconv.FormatInt(n, 10)

    return accountNumber
}