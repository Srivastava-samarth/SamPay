# Wallet Domain

## Overview

Every Merchant in SamPay owns exactly one Wallet.

A Wallet is the source of funds for Payments, Refunds and Payouts, and the destination of incoming Settlements.

The Wallet maintains the Merchant's balances and ensures that funds are reserved, settled and updated safely during financial operations.

Wallet balances are modified only through financial workflows and are never updated manually.

---

# Responsibilities

The Wallet domain is responsible for:

- Maintaining Merchant balances.
- Reserving funds for outgoing transactions.
- Receiving incoming pending funds.
- Releasing reserved funds on failures.
- Making settled funds available for spending.

The Wallet domain is **not** responsible for:

- Processing Payments.
- Processing Payouts.
- Processing Refunds.
- Settlement execution.
- Ledger management.
- Compliance validation.

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
| available_balance | Spendable balance |
| reserved_balance | Balance reserved for outgoing transactions |
| pending_balance | Incoming balance awaiting settlement |
| status | ACTIVE / SUSPENDED / CLOSED |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Balance Types

## Available Balance

Funds that are immediately available for use.

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
- Cannot be used for Payments, Refunds or Payouts.
- Becomes Available Balance once Settlement completes.

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
Debit Wallet
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
```

Settlement:

```text
Pending Balance
        │
        ▼
Available Balance
```

---

## Refund

Outgoing Refund:

```text
Available Balance
        │
        ▼
Reserved Balance
        │
        ▼
Debit Wallet
```

Incoming Refund:

```text
Refund Vault
        │
        ▼
Pending Balance
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
Bank Transfer
```

If the Payout fails:

```text
Reserved Balance
        │
        ▼
Available Balance
```

---

# Business Rules

## Wallet Creation

A Wallet is automatically created when a Merchant is onboarded.

Every Merchant owns exactly one Wallet.

---

## Currency

Version 1 supports INR only.

---

## Available Balance

Only Available Balance may be spent.

Available Balance can never become negative.

---

## Reserved Balance

Reserved Balance exists only while an outgoing transaction is in progress.

Reserved Balance is automatically released when the transaction fails.

---

## Pending Balance

Pending Balance represents incoming funds awaiting settlement.

Pending Balance:

- Is visible to the Merchant.
- Cannot be transferred.
- Cannot be withdrawn.
- Cannot be reserved.
- Cannot be used for Payments or Refunds.

Only Settlement may move Pending Balance into Available Balance.

---

## Concurrency

Wallet updates use row-level locking.

Only one balance modification may occur at a time.

This prevents double spending during concurrent requests.

---

## Financial Integrity

Every balance modification must:

- Update the Wallet.
- Create a Wallet History record.
- Create a corresponding Ledger entry.

All three operations must be committed within the same database transaction.

---

## Immutability

Wallet balances are modified only by financial workflows.

Direct balance updates are never permitted.

---

# APIs

Wallets are managed internally.

Read-only APIs may be exposed.

```http
GET /wallet

GET /wallet/history
```

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | 1 : 1 |
| Payment | Uses Wallet Balance |
| Refund | Uses Wallet Balance |
| Payout | Uses Wallet Balance |
| Settlement | Credits Pending Balance and settles it into Available Balance |
| Ledger | Records every balance movement |

---

# Notes

- Every Merchant owns exactly one Wallet.
- Available Balance represents spendable funds.
- Reserved Balance represents funds locked for outgoing transactions.
- Pending Balance represents incoming funds awaiting settlement.
- Wallet balances can never become negative.
- Every balance modification creates Wallet History and Ledger entries within the same database transaction.
- Wallet stores the current financial state, while Ledger stores the complete financial history.