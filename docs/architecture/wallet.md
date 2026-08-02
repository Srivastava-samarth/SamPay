# Wallet Domain

## Overview

Every Merchant in SamPay owns exactly one Wallet.

A Wallet maintains the Merchant's spendable, reserved and pending balances.

Financial operations modify Wallet balances, while Settlement moves pending funds into spendable funds.

Wallet balances may only be modified through financial workflows and are never updated manually.

---

# Responsibilities

The Wallet domain is responsible for:

- Maintaining Merchant balances.
- Reserving funds for outgoing transactions.
- Releasing reserved funds when transactions fail.
- Debiting reserved funds after successful transactions.
- Receiving incoming pending funds.
- Making settled funds available for spending.

The Wallet domain is **not** responsible for:

- Processing Payments.
- Processing Payouts.
- Processing Refunds.
- Executing Settlements.
- Managing the Ledger.
- Performing Compliance validation.

---

# Wallet Structure

Every Merchant owns exactly one Wallet.

A Wallet maintains three independent balances:

- Available Balance
- Reserved Balance
- Pending Balance

---

# Wallet Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| merchant_id | Owner Merchant |
| currency | Supported currency (INR) |
| available_balance | Spendable balance (stored in the smallest currency unit) |
| reserved_balance | Balance reserved for outgoing transactions |
| pending_balance | Incoming balance awaiting settlement |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Balance Types

## Available Balance

Funds immediately available for use.

Available Balance may be used for:

- Payments
- Refunds
- Payouts

---

## Reserved Balance

Funds temporarily locked for an outgoing transaction.

Reserved Balance cannot be spent again until the transaction either:

- Completes successfully, or
- Fails and releases the reservation.

---

## Pending Balance

Funds received from another Merchant but not yet settled.

Pending Balance:

- Is visible to the Merchant.
- Cannot be spent.
- Cannot be transferred.
- Cannot be reserved.
- Becomes Available Balance after Settlement.

---

# Wallet Lifecycle

## Outgoing Payment

```text
Available Balance
        │
        ▼
Reserved Balance
        │
        ▼
Reservation Consumed
```

If the Payment fails:

```text
Reserved Balance
        │
        ▼
Available Balance
```

---

## Incoming Payment

```text
Payment Vault
        │
        ▼
Pending Balance
        │
        ▼
Settlement
        │
        ▼
Available Balance
```

---

## Refund

### Outgoing Refund

```text
Available Balance
        │
        ▼
Reserved Balance
        │
        ▼
Reservation Consumed
```

### Incoming Refund

```text
Refund Vault
        │
        ▼
Pending Balance
        │
        ▼
Settlement
        │
        ▼
Available Balance
```

---

## Payout

```text
Available Balance
        │
        ▼
Reserved Balance
        │
        ▼
Reservation Consumed
        │
        ▼
Bank Transfer
```

If the Wallet has insufficient Available Balance, the Payout workflow may automatically top up the Wallet from the Merchant's Primary Linked Bank Account before reserving funds.

If the Payout fails:

```text
Reserved Balance
        │
        ▼
Available Balance
```

---

# Balance Operations

The Wallet supports the following balance operations:

- Reserve Funds
- Release Reserved Funds
- Consume Reserved Funds
- Credit Pending Balance
- Settle Pending Balance

All balance modifications must occur inside a single database transaction.

---

# Business Rules

## Wallet Creation

- Every Merchant owns exactly one Wallet.
- A Wallet is automatically provisioned during Merchant onboarding.
- A Wallet cannot exist without a Merchant.

---

## Currency

Version 1 supports INR only.

All balances are stored using the smallest currency unit (paise) as BIGINT values.

Example:

₹100.25 → 10025

---

## Available Balance

- Only Available Balance may be spent.
- Available Balance can never become negative.

---

## Reserved Balance

- Reserved Balance exists only while an outgoing transaction is in progress.
- Reserved Balance is automatically released when the transaction fails.

---

## Pending Balance

Pending Balance represents incoming funds awaiting Settlement.

Pending Balance:

- Is visible to the Merchant.
- Cannot be spent.
- Cannot be transferred.
- Cannot be withdrawn.
- Cannot be reserved.

Only Settlement may move Pending Balance into Available Balance.

---

## Concurrency

Wallet updates use row-level locking.

Only one balance modification may occur at a time.

This prevents double spending during concurrent financial operations.

---

## Financial Integrity

Every balance modification must:

- Update the Wallet.
- Create the corresponding Ledger entry.
- Commit atomically within the same database transaction.

---

## Immutability

Wallet balances may only be modified by:

- Payment
- Refund
- Payout
- Settlement

Direct balance updates are never permitted.

---

# APIs

Wallets are managed internally.

Read-only APIs may be exposed.

```http
GET /wallet
```

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | 1 : 1 |
| Payment | Uses Wallet Balance |
| Refund | Uses Wallet Balance |
| Payout | Uses Wallet Balance |
| Settlement | Moves Pending Balance to Available Balance |
| Ledger | Records every balance movement |

---

# Notes

- Every Merchant owns exactly one Wallet.
- Wallet maintains the Merchant's current financial state.
- Available Balance represents spendable funds.
- Reserved Balance represents funds locked for outgoing transactions.
- Pending Balance represents incoming funds awaiting Settlement.
- Wallet balances are stored using the smallest currency unit.
- Wallet balances can never become negative.
- Every balance modification creates a corresponding Ledger entry.
- Ledger stores the complete financial history, while Wallet stores only the current financial state.

---

# Version

**Architecture Status:** ✅ Frozen (V1)

**Last Updated:** 2026-08-03