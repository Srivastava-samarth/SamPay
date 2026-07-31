# Wallet Domain

## Overview

A Wallet represents the financial account owned by a Merchant within the SamPay platform.

Every Merchant is assigned exactly one Wallet that stores the funds available for performing financial operations such as Payments, Payouts and Refunds.

The Wallet is the source of truth for a Merchant's current balance within the platform.

---

# Responsibilities

The Wallet domain is responsible for:

- Maintaining the Merchant's available balance.
- Maintaining the Merchant's reserved balance.
- Preventing overspending.
- Ensuring balance updates are atomic.
- Maintaining the current financial state of the Merchant.

The Wallet domain is **not** responsible for:

- Payment Processing
- Payout Processing
- Refund Processing
- Settlement Processing
- Compliance
- Fraud Detection
- Notification Delivery
- Ledger Management

---

# Wallet Structure

A Wallet owns the following information:

- Merchant
- Currency
- Available Balance
- Reserved Balance
- Status
- Audit Information

---

# Wallet Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| wallet_reference | Public Wallet Identifier |
| merchant_id | Merchant owning the Wallet |
| currency | Wallet currency (INR) |
| available_balance | Funds immediately available for financial operations |
| reserved_balance | Funds temporarily reserved by in-progress financial operations |
| status | Current Wallet status |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Wallet Status

## ACTIVE

The Wallet is operational and can participate in all supported financial operations.

---

## SUSPENDED

The Wallet is temporarily disabled.

During this state:

- Payments are blocked.
- Payouts are blocked.
- Refunds are blocked.
- Existing balance remains unchanged.

---

## CLOSED

The Wallet is permanently disabled.

Business Rules:

- Wallet balance must be zero before closure.
- No further financial operations are permitted.

---

# Business Rules

## Ownership

- Every Merchant must have exactly one Wallet.
- A Wallet cannot exist without a Merchant.
- A Wallet cannot be transferred to another Merchant.

---

## Currency

- Every Wallet maintains exactly one currency.
- Version 1 supports INR only.

---

## Balance

- Available Balance represents spendable funds.
- Reserved Balance represents funds temporarily reserved during financial processing.
- Available Balance can never become negative.
- Reserved Balance can never become negative.

---

## Financial Operations

Before an outgoing financial operation begins:

1. Lock the Wallet row.
2. Validate the Available Balance.
3. Move the requested amount from Available Balance to Reserved Balance.

If the operation succeeds:

- Transfer the reserved amount to the appropriate platform Vault.
- Reduce the Reserved Balance.

If the operation fails:

- Move the reserved amount back to the Available Balance.

These balance transitions are performed internally by the Payment or Payout workflow and are not exposed as independent Wallet operations.

---

## Concurrency

- Wallet balance updates must always be atomic.
- Row-level locking is used to prevent concurrent balance modifications.
- Only one financial operation may modify a Wallet at any point in time.

---

## Balance Invariant

Before ownership of funds changes:

```text
Total Wallet Balance =
Available Balance + Reserved Balance
```

This invariant must always hold true.

---

# Domain Events

- WalletCreated
- WalletActivated
- WalletSuspended
- WalletClosed
- WalletBalanceUpdated

---

# APIs

## Wallet

```http
GET /wallets/{walletId}

GET /wallets/{walletId}/balance
```

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | 1 : 1 |
| Payment | 1 : N |
| Payout | 1 : N |
| Refund | 1 : N |
| Ledger | 1 : N |

---

# Notes

- The Wallet stores only the current financial state of the Merchant.
- Transaction history is owned by the Payment, Payout and Refund domains.
- Financial auditability is provided by the Ledger domain.
- All balance modifications occur as part of financial workflows. Direct balance updates are prohibited.

---

# Version 2 Considerations

Version 1 prioritizes correctness, simplicity and maintainability over extreme scalability.

As transaction volume grows, the Wallet domain can evolve to reduce balance update contention.

Potential improvements include:

- Ledger-first balance computation
- Event sourcing
- Balance projections
- Optimistic concurrency control
- Snapshot-based balance reconstruction
- CQRS for separating read and write workloads

These enhancements are intentionally deferred from Version 1 to avoid unnecessary complexity while the platform is still evolving.