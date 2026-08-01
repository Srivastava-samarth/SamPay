# Ledger Domain

## Overview

The Ledger is the financial source of truth for SamPay.

Every movement of money within the platform must create corresponding Ledger entries. The Ledger provides a complete audit trail of all financial transactions and ensures that money is never created or destroyed.

The Ledger follows the principles of double-entry accounting, where every debit has a corresponding credit.

---

# Responsibilities

The Ledger domain is responsible for:

- Recording every financial movement.
- Maintaining immutable financial records.
- Providing a complete audit trail.
- Supporting reconciliation.
- Recording platform fees.

The Ledger domain is **not** responsible for:

- Processing Payments.
- Processing Payouts.
- Processing Refunds.
- Managing Wallet balances.
- Managing Vault balances.
- Settlement execution.

---

# Ledger Structure

A Ledger Entry records:

- Source
- Destination
- Amount
- Currency
- Transaction Type
- Reference
- Audit Information

---

# Ledger Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| source_type | WALLET / PAYMENT_VAULT / PAYOUT_VAULT / REFUND_VAULT / FEE |
| source_id | UUID of the source entity |
| destination_type | WALLET / PAYMENT_VAULT / PAYOUT_VAULT / REFUND_VAULT / FEE |
| destination_id | UUID of the destination entity |
| amount | Amount transferred |
| currency | Supported currency (INR) |
| transaction_type | PAYMENT / PAYOUT / REFUND / SETTLEMENT / FEE |
| reference_type | PAYMENT / PAYOUT / REFUND / SETTLEMENT |
| reference_id | Associated business entity ID |
| description | Optional description |
| created_at | Record creation timestamp |

---

# Double-Entry Accounting

Every financial movement creates matching debit and credit entries.

Money is never created or destroyed.

### Example - Payment

```text
Merchant Wallet      -₹1,000
Payment Vault        +₹1,000
```

### Example - Settlement

```text
Payment Vault        -₹1,000
Recipient            +₹1,000
```

### Example - Refund

```text
Merchant Wallet      -₹500
Refund Vault         +₹500
```

### Example - Payout

```text
Merchant Wallet      -₹2,000
Payout Vault         +₹2,000
```

---

# Business Rules

## Immutability

Ledger entries are immutable.

Ledger records can never be:

- Updated
- Deleted

If an error occurs, a compensating Ledger entry must be created instead of modifying an existing one.

---

## Source & Destination

Every Ledger entry must contain:

- Source
- Destination

Both the source and destination must be explicitly recorded.

---

## Financial Integrity

Every debit must have a corresponding credit.

The total amount debited must always equal the total amount credited.

Money cannot be created or destroyed.

---

## Fees

Platform fees are recorded as independent Ledger entries.

Example:

Payment Amount: ₹1,000

Platform Fee: ₹10

```text
Merchant Wallet      -₹1,000
Payment Vault        +₹990
Fee                  +₹10
```

Fees are never merged into the payment amount.

---

## Currency

Version 1 supports INR only.

---

## References

Every Ledger entry references the business transaction that created it.

Supported Reference Types:

- PAYMENT
- PAYOUT
- REFUND
- SETTLEMENT

---

## Audit

Ledger acts as the permanent financial audit log for the platform.

Historical Ledger entries must never change.

---

## Atomic Financial Operations

Every financial operation that modifies a balance must also create the corresponding Ledger entry within the same database transaction.

The following operations must always succeed or fail together:

- Update balance
- Create history record
- Create Ledger entry

Example:

```text
BEGIN TRANSACTION

1. Update Wallet Balance
2. Create Wallet History
3. Create Ledger Entry

COMMIT
```

If any step fails, the entire transaction must be rolled back.

This guarantees that:

- Every balance change has a corresponding Ledger entry.
- Every Ledger entry represents an actual balance movement.
- Wallet balances and Ledger records are always consistent.
- Financial data remains accurate even during failures.

---

# Ledger Flow Examples

## Payment

```text
Merchant Wallet
        │
        ▼
Payment Vault
```

---

## Settlement

```text
Payment Vault
        │
        ▼
Recipient
```

---

## Payout

```text
Merchant Wallet
        │
        ▼
Payout Vault
```

---

## Refund

```text
Merchant Wallet
        │
        ▼
Refund Vault
```

---

# APIs

The Ledger is an internal financial component.

No public API exists to create Ledger entries.

Ledger entries are automatically created by Payment, Payout, Refund and Settlement workflows.

Read-only APIs may be exposed for administrative purposes.

```http
GET /ledger

GET /ledger/{ledgerId}
```

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Payment | Creates Ledger Entries |
| Payout | Creates Ledger Entries |
| Refund | Creates Ledger Entries |
| Settlement | Creates Ledger Entries |
| Wallet | Source or Destination |
| Payment Vault | Source or Destination |
| Payout Vault | Source or Destination |
| Refund Vault | Source or Destination |

---

# Notes

- Every financial workflow creates one or more Ledger entries.
- Ledger never initiates financial transactions.
- Wallet stores the current balance.
- Ledger stores the complete history of balance movements.
- Every movement of money is permanently recorded.
- Ledger is the financial source of truth for the platform.
- Every balance modification, history record and Ledger entry must be committed atomically within the same database transaction.
- Ledger follows the principles of double-entry accounting to ensure financial consistency and simplify reconciliation.