# Vault Domain

## Overview

A Vault is an internal SamPay financial account that temporarily holds funds during financial workflows.

Vaults isolate platform fund movement from Merchant Wallets and are never directly accessible by Merchants.

Vaults are internal platform infrastructure used by Payment, Payout, Refund and Settlement workflows.

---

# Responsibilities

The Vault domain is responsible for:

- Maintaining Vault balances.
- Providing secure temporary custody of funds.
- Supporting internal fund movement between financial workflows.

The Vault domain is **not** responsible for:

- Processing Payments.
- Processing Payouts.
- Processing Refunds.
- Executing Settlements.
- Managing the Ledger.
- Managing Merchants.
- Performing Compliance validation.
- Sending Notifications.

---

# Vault Structure

Every Vault maintains:

- Vault Type
- Currency
- Current Balance

---

# Vault Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| vault_type | PAYMENT / PAYOUT / REFUND / COMPANY |
| currency | Supported currency (INR) |
| balance | Current balance (stored in the smallest currency unit) |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Vault Types

## PAYMENT

Temporarily holds incoming payment funds before Settlement.

---

## PAYOUT

Temporarily holds funds reserved for outgoing bank payouts.

---

## REFUND

Temporarily holds refund funds before refund completion.

---

## COMPANY

Stores:

- Platform operational funds
- Platform revenue
- Transaction fees

The Company Vault may replenish operational Vaults whenever additional liquidity is required.

---

# Vault Operations

Vaults support only two balance operations:

- Credit Funds
- Debit Funds

Vaults never perform business validation.

They simply maintain balances.

---

# Business Rules

## Ownership

- Every Vault is owned by SamPay.
- Merchants never own Vaults.
- Merchants cannot directly access Vaults.

---

## Vault Count

Version 1 maintains exactly one Vault for each Vault Type.

- One Payment Vault
- One Payout Vault
- One Refund Vault
- One Company Vault

---

## Currency

- Every Vault supports exactly one currency.
- Version 1 supports INR only.
- Balances are stored using the smallest currency unit (paise) as BIGINT values.

---

## Balance

- Vault balance represents the total funds currently held.
- Vault balance can never become negative.
- Money can never be created or destroyed.

---

## Fund Movement

Vaults support only:

- Credit Funds
- Debit Funds

Vaults do not understand business context.

Business workflows such as Payment, Refund, Payout and Settlement determine when a Vault should be credited or debited.

---

## Internal Access

Vaults are internal platform infrastructure.

Only internal financial workflows may modify Vault balances.

Public APIs never directly expose Vault operations.

---

## Company Vault

The Company Vault stores:

- Platform revenue
- Operational funds
- Transaction fees

The Company Vault may transfer funds into operational Vaults whenever additional liquidity is required.

---

## Concurrency

- Vault balance updates are atomic.
- Row-level locking is used during balance modifications.
- Concurrent modifications to the same Vault are prevented.

---

## Financial Integrity

Every Vault balance modification must:

- Update the Vault balance.
- Create the corresponding Ledger entry.
- Commit atomically within the same database transaction.

---

# Domain Events

- VaultCredited
- VaultDebited

---

# APIs

Vaults expose no public APIs.

All interactions occur internally through financial workflows.

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Payment | Credits / Debits Payment Vault |
| Payout | Uses Payout Vault |
| Refund | Uses Refund Vault |
| Settlement | Moves funds from Vault to Merchant Wallet |
| Ledger | Records every Vault balance movement |

---

# Notes

- Vaults are internal SamPay financial accounts.
- Vaults are never exposed to Merchants.
- Every balance modification creates a corresponding Ledger entry.
- Vaults never store business context such as Payment IDs, Refund IDs or Payout IDs.
- Business workflows own transaction context, while Vaults own only financial balances.

---

# Version 2 Considerations

Future enhancements may include:

- Multi-currency Vaults.
- Regional Vaults.
- Treasury management.
- Automated liquidity balancing.
- Vault sharding.
- Real-time liquidity monitoring.

These enhancements are intentionally deferred from Version 1 to keep the platform simple, reliable and maintainable.

---

# Version

**Architecture Status:** ✅ Frozen (V1)

**Last Updated:** 2026-08-03