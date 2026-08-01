# Settlement Domain

## Overview

Settlement is responsible for finalizing internal fund transfers within SamPay.

After a successful Payment or Refund, funds are temporarily held in the appropriate Vault. Settlement transfers those funds into the recipient's Wallet by moving the amount from the recipient's Pending Balance to Available Balance.

Settlement is processed asynchronously and guarantees reliable internal movement of funds.

---

# Responsibilities

The Settlement domain is responsible for:

- Processing pending settlements.
- Moving funds from Vaults to Merchant Wallets.
- Crediting Pending Balance.
- Converting Pending Balance into Available Balance.
- Updating settlement status.
- Retrying failed settlements.
- Publishing settlement events.

The Settlement domain is **not** responsible for:

- Processing Payments.
- Processing Refunds.
- Processing Payouts.
- Managing Wallet ownership.
- Ledger management.
- Compliance validation.

---

# Settlement Structure

A Settlement represents one internal transfer waiting to be completed.

Each Settlement belongs to exactly one business transaction.

---

# Settlement Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| reference_type | PAYMENT / REFUND |
| reference_id | Payment or Refund ID |
| source_vault_id | Vault holding the funds |
| destination_wallet_id | Recipient Wallet |
| amount | Settlement amount |
| currency | Supported currency (INR) |
| status | Current settlement status |
| retry_count | Number of retry attempts |
| failure_reason | Last failure reason |
| settled_at | Settlement completion timestamp |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Settlement Status

## PENDING

Settlement has been created and is waiting to be processed.

---

## PROCESSING

Settlement Worker is currently processing the transfer.

---

## COMPLETED

Funds have been successfully settled into the destination Wallet.

---

## FAILED

Settlement could not be completed after the maximum retry attempts.

Funds remain safely held in the Vault until Operations retries the Settlement.

---

# Settlement Workflow

```text
Payment / Refund Completed
        │
        ▼
Create Settlement
        │
        ▼
Status = PENDING
        │
        ▼
Settlement Worker
        │
        ▼
Lock Destination Wallet
        │
        ▼
Credit Pending Balance
        │
        ▼
Move Pending Balance
        │
        ▼
Available Balance
        │
        ▼
Create Ledger Entry
        │
        ▼
Mark Settlement COMPLETED
        │
        ▼
Publish SettlementCompleted Event
```

---

# Business Rules

## Settlement Creation

A Settlement is automatically created after a successful:

- Payment
- Refund

---

## Internal Transfers Only

Settlement only handles internal transfers within SamPay.

Funds always move:

- From a Vault
- To a SamPay Wallet

Settlement never transfers funds to external Bank Accounts.

---

## Pending Balance

Incoming funds are first credited to the destination Wallet's Pending Balance.

Pending Balance is not spendable.

Only Settlement can move Pending Balance into Available Balance.

---

## Available Balance

After Settlement completes:

- Pending Balance decreases.
- Available Balance increases.

The settled funds become immediately spendable.

---

## Retry Policy

Settlement failures are retried automatically.

Maximum retry attempts:

```
5 (configurable)
```

After the retry limit is exceeded:

- Settlement is marked FAILED.
- Funds remain in the Vault.
- Manual retry may be performed by Operations.

---

## Concurrency

Settlement uses row-level locking when updating Wallet balances.

Only one Settlement may modify a Wallet at a time.

---

## Atomic Financial Operations

Every successful Settlement must execute within a single database transaction.

The following operations must succeed or fail together:

- Update destination Wallet balances.
- Create Ledger entry.
- Update Settlement status.

If any operation fails, the transaction is rolled back.

---

## Currency

Version 1 supports INR only.

---

## Immutability

Completed Settlements are immutable.

Historical Settlement records must never be modified.

---

# Domain Events

Settlement publishes:

- SettlementCompleted
- SettlementFailed

These events may be consumed by:

- Notification
- Audit

---

# APIs

Settlement is an internal domain.

No public APIs are exposed for creating Settlements.

Administrative read-only APIs may be provided.

```http
GET /settlements

GET /settlements/{settlementId}

POST /settlements/{settlementId}/retry
```

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Payment | Creates Settlement |
| Refund | Creates Settlement |
| Wallet | Destination |
| Vault | Source |
| Ledger | Records Settlement |
| Notification | Consumes Settlement events |

---

# Notes

- Settlement is an internal asynchronous process.
- Settlement only moves funds within the SamPay ecosystem.
- Settlement never communicates with external banking providers.
- Incoming funds become visible immediately as Pending Balance.
- Only completed Settlements move Pending Balance into Available Balance.
- Every successful Settlement creates corresponding Ledger entries.
- Settlement guarantees reliable and consistent internal fund movement.