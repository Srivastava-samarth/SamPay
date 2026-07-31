# Vault Domain

## Overview

A Vault is an internal financial account owned by SamPay.

Vaults are responsible for temporarily holding and managing funds as they move through different financial workflows such as Payments, Payouts, Refunds and Settlements.

Vaults are internal platform entities and are never directly exposed to Merchants.

---

# Responsibilities

The Vault domain is responsible for:

- Maintaining platform funds.
- Managing balances for different financial workflows.
- Providing a secure intermediary for fund movement.
- Maintaining audit information.

The Vault domain is **not** responsible for:

- Payment Processing.
- Payout Processing.
- Refund Processing.
- Settlement Scheduling.
- Ledger Management.
- Merchant Management.
- Notification Delivery.
- Compliance.

---

# Vault Structure

A Vault owns the following information:

- Vault Type
- Currency
- Current Balance
- Status
- Audit Information

---

# Vault Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| vault_type | Type of Vault |
| currency | Supported currency (INR) |
| balance | Current balance maintained by the Vault |
| status | Current Vault status |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Vault Types

## PAYMENT

Temporarily stores incoming payment funds before settlement.

---

## PAYOUT

Stores funds reserved for outgoing payouts.

---

## REFUND

Stores funds used for refund processing.

---

## COMPANY

Stores platform fees and operational funds.

The Company Vault may also replenish operational Vaults when required.

---

# Vault Status

## ACTIVE

The Vault is operational and can participate in financial workflows.

---

## INACTIVE

The Vault is disabled and cannot participate in any financial workflow.

This status is expected to be rarely used since Vaults are core platform infrastructure.

---

# Business Rules

## Ownership

- Every Vault is owned by SamPay.
- Merchants never own or directly interact with Vaults.

---

## Vault Count

Version 1 maintains exactly one Vault for each Vault Type.

- One Payment Vault
- One Payout Vault
- One Refund Vault
- One Company Vault

---

## Currency

- Every Vault maintains exactly one currency.
- Version 1 supports INR only.

---

## Balance

- Vault balance represents the total funds currently held within that Vault.
- Vault balance can never become negative.
- Money cannot be created or destroyed.

---

## Fund Movement

Vaults only support two financial operations:

- Credit
- Debit

The Vault does not know why funds are moving.

Business workflows such as Payment, Payout, Refund and Settlement determine when a Vault should be credited or debited.

---

## Internal Access

Vaults are internal platform entities.

Only internal financial workflows may modify Vault balances.

Merchants cannot directly access or modify Vaults.

---

## Company Vault

The Company Vault stores:

- Transaction fees
- Platform revenue
- Operational funds

It may transfer funds to operational Vaults when additional liquidity is required.

---

## Concurrency

- Vault balance updates must always be atomic.
- Row-level locking is used during balance modifications.
- Concurrent updates to the same Vault must be prevented.

---

# Domain Events

- VaultCredited
- VaultDebited
- VaultCreated
- VaultStatusChanged

---

# APIs

Vaults do not expose any public APIs.

All interactions occur internally through financial workflows.

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Payment | Uses Payment Vault |
| Payout | Uses Payout Vault |
| Refund | Uses Refund Vault |
| Settlement | Debits Payment Vault |
| Ledger | Records every Vault balance movement |

---

# Notes

- Vaults represent SamPay's internal financial accounts.
- Every balance modification creates a corresponding Ledger entry.
- Every update is recorded in the append-only `vault_history` table.
- Vaults never store business context such as Payment IDs or Refund IDs. That responsibility belongs to the originating financial workflow.

---

# Version 2 Considerations

As SamPay evolves, the Vault domain may support:

- Multiple currencies.
- Regional Vaults.
- Treasury management.
- Automated liquidity balancing.
- Vault sharding for high throughput.
- Real-time liquidity monitoring.

These enhancements are intentionally deferred from Version 1 to prioritize simplicity, correctness and maintainability.