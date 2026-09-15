package utils

import (
	"strings"
	"testing"

	"github.com/Srivastava-samarth/sampay/constants"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestGenerateUUID(t *testing.T) {
	id := GenerateUUID()

	if id == uuid.Nil {
		t.Fatal("expected non-nil UUID")
	}

	id2 := GenerateUUID()

	if id == id2 {
		t.Fatal("expected generated UUIDs to be different")
	}
}

func TestGenerateMerchantReference(t *testing.T) {
	ref := GenerateMerchantReference()

	if ref == "" {
		t.Fatal("expected reference, got nil")
	}

	if !strings.HasPrefix(ref, "mrc_") {
		t.Fatalf("expected prefix mrc_, got %s", ref)
	}

	if len(ref) != len("mrc_")+12 {
		t.Fatalf("expected reference length %d, got %d",
			len("mrc_")+12,
			len(ref),
		)
	}
}

func TestGenerateLedgerReference(t *testing.T) {
	ref := GenerateLedgerReference()

	if ref == "" {
		t.Fatal("expected reference, got nil")
	}

	if !strings.HasPrefix(ref, "ldg_") {
		t.Fatalf("expected prefix ldg_, got %s", ref)
	}

	if len(ref) != len("ldg_")+12 {
		t.Fatalf("expected reference length %d, got %d",
			len("ldg_")+12,
			len(ref),
		)
	}
}

func TestGenerateAutoTopUpReference(t *testing.T) {
	ref := GenerateAutoTopUpReference()

	if ref == "" {
		t.Fatal("expected reference, got nil")
	}

	if !strings.HasPrefix(ref, "auto_topup_") {
		t.Fatalf("expected prefix auto_topup_, got %s", ref)
	}

	if len(ref) != len("auto_topup_")+12 {
		t.Fatalf("expected reference length %d, got %d",
			len("auto_topup_")+12,
			len(ref),
		)
	}
}

func TestGeneratePaymentReference(t *testing.T) {
	ref := GeneratePaymentReference()

	if ref == nil {
		t.Fatal("expected reference, got nil")
	}

	if !strings.HasPrefix(*ref, "pmt_") {
		t.Fatalf("expected prefix pmt_, got %s", *ref)
	}

	if len(*ref) != len("pmt_")+12 {
		t.Fatalf("expected reference length %d, got %d",
			len("pmt_")+12,
			len(*ref),
		)
	}
}

func TestGeneratePayoutReference(t *testing.T) {
	ref := GeneratePayoutReference()

	if ref == nil {
		t.Fatal("expected reference, got nil")
	}

	if !strings.HasPrefix(*ref, "pyt_") {
		t.Fatalf("expected prefix pyt_, got %s", *ref)
	}

	if len(*ref) != len("pyt_")+12 {
		t.Fatalf("expected reference length %d, got %d",
			len("pyt_")+12,
			len(*ref),
		)
	}
}

func TestGenerateRefundReference(t *testing.T) {
	ref := GenerateRefundReference()

	if ref == nil {
		t.Fatal("expected reference, got nil")
	}

	if !strings.HasPrefix(*ref, "rfd_") {
		t.Fatalf("expected prefix rfd_, got %s", *ref)
	}

	if len(*ref) != len("rfd_")+12 {
		t.Fatalf("expected reference length %d, got %d",
			len("rfd_")+12,
			len(*ref),
		)
	}
}

func TestGenerateCustomerReference(t *testing.T) {
	ref := GenerateCustomerReference()

	if ref == nil {
		t.Fatal("expected reference, got nil")
	}

	if !strings.HasPrefix(*ref, "cust_") {
		t.Fatalf("expected prefix cust_, got %s", *ref)
	}

	if len(*ref) != len("cust_")+12 {
		t.Fatalf("expected reference length %d, got %d",
			len("cust_")+12,
			len(*ref),
		)
	}
}

func TestGenerateBankAccountNumber(t *testing.T) {
	accountNumber := GenerateBankAccountNumber()

	if accountNumber == "" {
		t.Fatal("expected account number, got nil")
	}

	if len(accountNumber) != 13 {
		t.Fatalf(
			"expected 13 digit account number, got %d digits",
			len(accountNumber),
		)
	}

	if accountNumber < "1000000000000" ||
		accountNumber > "9999999999999" {
		t.Fatalf(
			"account number %s is outside valid range",
			accountNumber,
		)
	}

	accountNumber2 := GenerateBankAccountNumber()

	if accountNumber == accountNumber2 {
		t.Fatal("expected different account numbers")
	}
}

func TestGenerateTemporaryPassword(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "short password",
			length: 8,
		},
		{
			name:   "medium password",
			length: 16,
		},
		{
			name:   "long password",
			length: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := GenerateTemporaryPassword(tt.length)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if password == "" {
				t.Fatal("expected password, got nil")
			}

			if len(password) != tt.length {
				t.Fatalf(
					"expected password length %d, got %d",
					tt.length,
					len(password),
				)
			}

			for _, char := range password {
				if !strings.ContainsRune(passwordChars, char) {
					t.Fatalf(
						"password contains invalid character %q",
						char,
					)
				}
			}
		})
	}

	t.Run("different passwords are generated", func(t *testing.T) {
		password1, err := GenerateTemporaryPassword(16)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		password2, err := GenerateTemporaryPassword(16)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if password1 == password2 {
			t.Fatal("expected generated passwords to be different")
		}
	})
}

func TestHashPassword(t *testing.T) {
	password := "my-secret-password"

	hash, err := HashPassword(password)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hash == "" {
		t.Fatal("expected password hash")
	}

	if hash == password {
		t.Fatal("password should not be stored as plain text")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)

	if err != nil {
		t.Fatalf("generated hash does not match password: %v", err)
	}

	hash2, err := HashPassword(password)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hash == hash2 {
		t.Fatal("expected bcrypt hashes to differ because of random salt")
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token"

	hash := HashToken(token)

	if hash == "" {
		t.Fatal("expected hash, got nil")
	}

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	if len(hash) != 64 {
		t.Fatalf(
			"expected SHA-256 hex hash length 64, got %d",
			len(hash),
		)
	}

	hash2 := HashToken(token)

	if hash != hash2 {
		t.Fatal("expected same token to produce same hash")
	}

	differentToken := "different-token"

	hash3 := HashToken(differentToken)

	if hash == hash3 {
		t.Fatal("expected different tokens to produce different hashes")
	}
}

func TestIsValidVaultStatus(t *testing.T) {
	active := constants.VaultStatusActive
	inactive := constants.VaultStatusInactive
	invalid := "invalid"

	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "active status",
			status:   active,
			expected: true,
		},
		{
			name:     "inactive status",
			status:   inactive,
			expected: true,
		},
		{
			name:     "invalid status",
			status:   invalid,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidVaultStatus(tt.status)

			if result != tt.expected {
				t.Fatalf(
					"expected %v, got %v",
					tt.expected,
					result,
				)
			}
		})
	}
}

func TestIsValidVaultType(t *testing.T) {
	paymentVault := constants.PaymentVault
	payoutVault := constants.PayoutVault
	companyVault := constants.CompanyVault
	invalid := "invalid"

	tests := []struct {
		name      string
		vaultType string
		expected  bool
	}{
		{
			name:      "payment vault",
			vaultType: paymentVault,
			expected:  true,
		},
		{
			name:      "payout vault",
			vaultType: payoutVault,
			expected:  true,
		},
		{
			name:      "company vault",
			vaultType: companyVault,
			expected:  true,
		},
		{
			name:      "invalid vault type",
			vaultType: invalid,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidVaultType(tt.vaultType)
			if result != tt.expected {
				t.Fatalf(
					"expected %v, got %v",
					tt.expected,
					result,
				)
			}
		})
	}
}

func TestIsValidMerchantStatus(t *testing.T) {
	active := constants.MerchantStatusActive
	pending := constants.MerchantStatusPending
	suspended := constants.MerchantStatusSuspended
	inactive := constants.MerchantStatusInactive
	invalid := "invalid"

	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "active status",
			status:   active,
			expected: true,
		},
		{
			name:     "pending status",
			status:   pending,
			expected: true,
		},
		{
			name:     "suspended status",
			status:   suspended,
			expected: true,
		},
		{
			name:     "inactive status",
			status:   inactive,
			expected: true,
		},
		{
			name:     "invalid status",
			status:   invalid,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidMerchantStatus(tt.status)
			if result != tt.expected {
				t.Fatalf(
					"expected %v, got %v",
					tt.expected,
					result,
				)
			}
		})
	}
}
