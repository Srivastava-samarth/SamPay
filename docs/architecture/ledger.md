# Ledger Domain

## Overview

The Ledger is the financial source of truth for SamPay.

Every movement of money within the platform creates a corresponding immutable Ledger entry.

The Ledger records how money moves between financial accounts and guarantees that money is never created or destroyed.

The Ledger follows the principles of double-entry accounting by recording both the source and destination of every financial movement.

---

# Responsibilities

The Ledger domain is responsible for:

- Recording every financial movement.
- Maintaining immutable financial records.
- Providing the complete financial history.
- Supporting reconciliation.
- Recording platform fees.

The Ledger domain is **not** responsible for:

- Processing Payments.
- Processing Payouts.
- Processing Refunds.
- Managing Wallet balances.
- Managing Vault balances.
- Executing Settlements.

---

# Ledger Structure

A Ledger Entry records:

- From Account
- To Account
- Amount
- Currency
- Business Entity
- Description

---

# Ledger Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| from_account_type | WALLET / PAYMENT_VAULT / PAYOUT_VAULT / REFUND_VAULT / COMPANY_VAULT |
| from_account_id | UUID of the source account |
| to_account_type | WALLET / PAYMENT_VAULT / PAYOUT_VAULT / REFUND_VAULT / COMPANY_VAULT |
| to_account_id | UUID of the destination account |
| amount | Amount transferred (stored in the smallest currency unit) |
| currency | Supported currency (INR) |
| entity_type | PAYMENT / PAYOUT / REFUND / SETTLEMENT |
| entity_id | Associated business entity |
| description | Optional business description |
| created_at | Record creation timestamp |

---

# Account Types

Ledger entries may reference the following account types:

- Merchant Wallet
- Payment Vault
- Payout Vault
- Refund Vault
- Company Vault

---

# Double-Entry Accounting

Every financial movement records both the source account and destination account.

Money is never created or destroyed.

### Payment

```text
From : Merchant Wallet
To   : Payment Vault
Amount : ₹1,000
```

---

### Settlement

```text
From : Payment Vault
To   : Merchant Wallet
Amount : ₹1,000
```

---

### Refund

```text
From : Merchant Wallet
To   : Refund Vault
Amount : ₹500
```

---

### Payout

```text
From : Merchant Wallet
To   : Payout Vault
Amount : ₹2,000
```

---

# Business Rules

## Immutability

Ledger entries are immutable.

Ledger records can never be:

- Updated
- Deleted

If an error occurs, a compensating Ledger entry must be created instead.

---

## Financial Accounts

Every Ledger entry must contain:

- From Account
- To Account

Both accounts must always be explicitly recorded.

---

## Financial Integrity

Every debit must have a corresponding credit.

Money can never be created or destroyed.

The total amount leaving one account must equal the total amount entering another account.

---

## Fees

Platform fees are recorded as independent Ledger entries.

Example:

Payment Amount : ₹1,000

Platform Fee : ₹10

Payment Movement

```text
Merchant Wallet
        │
        ▼
Payment Vault
        ₹990
```

Fee Movement

```text
Merchant Wallet
        │
        ▼
Company Vault
        ₹10
```

Fees are never merged into payment movements.

---

## Currency

Version 1 supports INR only.

All monetary values are stored using the smallest currency unit (paise) as BIGINT values.

Example:

₹100.25 → 10025

---

## References

Every Ledger entry belongs to exactly one business entity.

Supported entity types:

- PAYMENT
- PAYOUT
- REFUND
- SETTLEMENT

---

## Atomic Financial Operations

Every balance modification must create the corresponding Ledger entry within the same database transaction.

The following operations must always succeed or fail together:

```text
BEGIN TRANSACTION

1. Update Wallet or Vault Balance
2. Create Ledger Entry

COMMIT
```

If any step fails, the entire transaction is rolled back.

This guarantees:

- Every balance movement has a corresponding Ledger entry.
- Every Ledger entry represents an actual balance movement.
- Wallets, Vaults and Ledger always remain consistent.

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
Merchant Wallet
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

## Platform Fee

```text
Merchant Wallet
        │
        ▼
Company Vault
```

---

# APIs

The Ledger is an internal financial component.

Ledger entries are automatically created by Payment, Payout, Refund and Settlement workflows.

No public API exists to create Ledger entries.

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
| Company Vault | Source or Destination |

---

# Notes

- Every financial workflow creates one or more Ledger entries.
- Ledger never initiates financial transactions.
- Wallet stores the current financial state.
- Vaults store SamPay's internal financial state.
- Ledger stores the complete immutable financial history.
- Every movement of money is permanently recorded.
- Every balance modification and Ledger entry are committed atomically within the same database transaction.
- Ledger records financial movement, while business workflows own business logic.

---

# Version

**Architecture Status:** ✅ Frozen (V1)

**Last Updated:** 2026-08-03