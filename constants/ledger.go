package constants

const (
	LedgerTransactionTypePayment     = "payment"
	LedgerTransactionTypePayout      = "payout"
	LedgerTransactionTypeRefund      = "refund"
	LedgerTransactionTypeSettlement  = "settlement"
)

const (
	LedgerSettlementPending = "pending"
	LedgerSettlementSettled = "settled"
	LedgerSettlementFailed  = "failed"
)

const (
	LedgerAccountTypeWallet      = "wallet"
	LedgerAccountTypeBankAccount = "bank_account"
	LedgerAccountTypeVault       = "vault"
)

const (
	LedgerEntryTypeDebit  = "debit"
	LedgerEntryTypeCredit = "credit"
)